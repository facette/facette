package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"facette/backend"

	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
)

type libraryCommand struct {
	Enable bool

	Dump struct {
		Enable bool
		Output string `names:"-o, --ouput" usage:"Dump output file path"`
	} `usage:"Dump data from library" expand:"1"`

	Restore struct {
		Enable bool
		Input  string `names:"-i, --input" usage:"Dump input file path"`
		Merge  bool   `names:"-m, --merge" usage:"Merge data with existing library"`
	} `usage:"Restore data from dump to library" expand:"1"`
}

type collectionRef struct {
	ID       string
	Children []collectionRef
}

var (
	// Backend types are listed according to their restoration order
	backendTypes = []string{
		"collections",
		"graphs",
		"sourcegroups",
		"metricgroups",
		"providers",
	}
)

func execLibrary() error {
	if cmd.Library.Dump.Enable {
		return execLibraryDump()
	} else if cmd.Library.Restore.Enable {
		return execLibraryRestore()
	}

	set, _ := flagSet.FindSubset("library")
	set.Help(false)

	return nil
}

func execLibraryDump() error {
	var errored bool

	output := cmd.Library.Dump.Output
	if output == "" {
		output = fmt.Sprintf("facette-%s.tar.gz", time.Now().Format("200601021504"))
	}

	// Prepare output archive file for data dump
	fd, err := os.OpenFile(output, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open archive")
	}
	defer fd.Close()

	gzw := gzip.NewWriter(fd)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Loop through back-end types and dump data
	for _, typ := range backendTypes {
		if err := dumpLibraryType(tw, typ); err != nil {
			printError("failed to dump %s: %s", typ, err)
			errored = true
			continue
		}
	}

	if !cmd.Quiet {
		if errored {
			fmt.Println(ansi.Color("FAIL", "red"))
		} else {
			fmt.Println(ansi.Color("OK", "green"))
		}
	}

	if errored {
		return errExecFailed
	}

	return nil
}

func execLibraryRestore() error {
	var errored bool

	// Open archive file for data retrieval
	fd, err := os.OpenFile(cmd.Library.Restore.Input, os.O_RDONLY, 0444)
	if err != nil {
		return errors.Wrap(err, "failed to open archive")
	}
	defer fd.Close()

	gzr, err := gzip.NewReader(fd)
	if err != nil {
		return errors.Wrap(err, "failed to create Gzip reader")
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Perform library cleanup if merging is not requested
	if !cmd.Library.Restore.Merge {
		for _, typ := range backendTypes {
			if !cmd.Quiet {
				fmt.Printf("purging %s...\n", typ)
			}

			endpoint := "/" + typ
			if typ != "providers" {
				endpoint = "/library" + endpoint
			}

			if err := apiRequest("DELETE", endpoint, map[string]string{"X-Confirm-Action": "1"}, nil, nil); err != nil {
				printError("failed to delete %s: %s", typ, err)
				continue
			}
		}
	}

	// Restore items
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			die("%s", err)
		}

		switch h.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			endpoint := "/" + h.Name
			if endpoint[:strings.LastIndex(endpoint, "/")] != "/providers" {
				endpoint = "/library" + endpoint
			}

			if !cmd.Quiet {
				fmt.Printf("restoring %q library item...\n", h.Name)
			}

			if err := apiRequest("PUT", endpoint, nil, tr, nil); err != nil {
				printError("%s", err)
				errored = true
				continue
			}
		}
	}

	if !cmd.Quiet {
		if errored {
			fmt.Println(ansi.Color("FAIL", "red"))
		} else {
			fmt.Println(ansi.Color("OK", "green"))
		}
	}

	if errored {
		return errExecFailed
	}

	return nil
}

func dumpLibraryType(w *tar.Writer, typ string) error {
	var params string

	// IMPORTANT:
	// Templates and parent collections *MUST* be dumped first in order to be restored first
	// to preserve back-end items relationships.

	endpoint := "/" + typ
	if typ != "providers" {
		endpoint = "/library" + endpoint
	}

	if typ == "collections" {
		// Dump only collection templates since they are not included in the collections tree dump
		params = "&kind=template"
	} else if typ == "graphs" {
		// Sort graphs by link (null first) to ensure templates are restored before template their instances
		params = "&sort=link"
	}

	// Retrieve items list
	items := []backend.Item{}
	if err := apiRequest("GET", endpoint+"?fields=id,created,modified"+params, nil, nil, &items); err != nil {
		return errors.Wrap(err, "failed to retrieve items")
	}

	if typ == "collections" {
		var (
			root []collectionRef
			cur  []collectionRef
		)

		if err := apiRequest("GET", endpoint+"/tree", nil, nil, &root); err != nil {
			return errors.Wrap(err, "failed to retrieve collections tree")
		}

		stack := [][]collectionRef{root}
		for len(stack) > 0 {
			cur, stack = stack[0], stack[1:]

			for _, c := range cur {
				items = append(items, backend.Item{ID: c.ID})
				if len(c.Children) > 0 {
					stack = append(stack, c.Children)
				}
			}
		}
	}

	for _, item := range items {
		var v interface{}

		if !cmd.Quiet {
			fmt.Printf("dumping %q library item...\n", item.ID)
		}

		if err := apiRequest("GET", endpoint+"/"+item.ID, nil, nil, &v); err != nil {
			printError("%s", err)
			continue
		}

		// Append data to archive
		data, err := json.Marshal(v)
		if err != nil {
			printError("failed to marshal JSON: %s", err)
			continue
		}

		if err = w.WriteHeader(&tar.Header{
			Name:    typ + "/" + item.ID,
			Size:    int64(len(data)),
			Mode:    0644,
			ModTime: item.Modified,
		}); err != nil {
			return errors.Wrap(err, "failed to append data")
		}

		w.Write(data)
	}

	return nil
}
