package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"facette/backend"

	"github.com/facette/httputil"
)

type backendError struct {
	Message string `json:"message"`
}

type backendType struct {
	prefix      string
	reflectType reflect.Type
}

var (
	backendTypes = []string{
		"units",
		"scales",
		"metricgroups",
		"sourcegroups",
		"graphs",
		"collections",
		"providers",
	}

	backendAttrs = map[string]backendType{
		"collections":  {"library/", reflect.TypeOf(backend.Collection{})},
		"graphs":       {"library/", reflect.TypeOf(backend.Graph{})},
		"metricgroups": {"library/", reflect.TypeOf(backend.MetricGroup{})},
		"providers":    {"", reflect.TypeOf(backend.Provider{})},
		"scales":       {"library/", reflect.TypeOf(backend.Scale{})},
		"sourcegroups": {"library/", reflect.TypeOf(backend.SourceGroup{})},
		"units":        {"library/", reflect.TypeOf(backend.Unit{})},
	}
)

func execBackupDump(output string) {
	if output == "" {
		output = fmt.Sprintf("facette-%s.tar.gz", time.Now().Format("200601021504"))
	}

	// Prepare output archive file for data dump
	fd, err := os.OpenFile(output, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		die("failed to open archive: %s", err)
	}
	defer fd.Close()

	gzw := gzip.NewWriter(fd)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Loop through backend types and dump data
	for _, name := range backendTypes {
		if err := dumpBackupType(tw, name); err != nil {
			printError("%s", err)
		}
	}
}

func execBackupRestore(input string, merge bool) {
	// Open archive file for data retrieval
	fd, err := os.OpenFile(input, os.O_RDONLY, 0444)
	if err != nil {
		printError("%s", err)
		return
	}
	defer fd.Close()

	gzr, err := gzip.NewReader(fd)
	if err != nil {
		die("%s", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Create new HTTP request
	hc := httputil.NewClient(time.Duration(upstreamTimeout)*time.Second, true, false)

	if !merge {
		for _, typ := range backendTypes {
			path := typ
			if typ != "providers" {
				path = "library/" + path
			}

			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/%s/", upstreamAddress, path), nil)
			if err != nil {
				printError("%s", err)
				continue
			}

			req.Header.Add("User-Agent", "facettectl/"+version)
			req.Header.Add("X-Confirm-Action", "1")

			resp, err := hc.Do(req)
			if err != nil {
				printError("%s", err)
				continue
			}
			defer resp.Body.Close()
		}
	}

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			die("%s", err)
		}

		switch h.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			path := h.Name
			if path[:strings.LastIndex(path, "/")] != "providers" {
				path = "library/" + path
			}

			// Register item into library
			req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/%s", upstreamAddress, path), tr)
			if err != nil {
				printError("%s", err)
				return
			}

			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("User-Agent", "facettectl/"+version)

			resp, err := hc.Do(req)
			if err != nil {
				printError("%s", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				printError("unable to restore %q item, returned: %s", path, resp.Status)
			}
		}
	}
}

func dumpBackupType(w *tar.Writer, name string) error {
	var extra string

	if name == "collections" {
		extra = "?sort=parent"
	}

	ba, _ := backendAttrs[name]

	// Create new HTTP request
	hc := httputil.NewClient(time.Duration(upstreamTimeout)*time.Second, true, false)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/%s%s%s", upstreamAddress, ba.prefix, name, extra), nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", "facettectl/"+version)

	// Retrieve items list from library
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var v backendError
		if err := httputil.BindJSON(resp, &v); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %s", err)
		}

		return fmt.Errorf("failed to list items: %s", v.Message)
	}

	rv := reflect.New(reflect.MakeSlice(reflect.SliceOf(ba.reflectType), 0, 0).Type())
	if err := httputil.BindJSON(resp, rv.Interface()); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %s", err)
	}

	n := reflect.Indirect(rv).Len()
	for i := 0; i < n; i++ {
		var mt time.Time

		v := reflect.Indirect(rv).Index(i)

		id := v.FieldByName("ID").String()

		// Retrieve item data from library
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/%s%s/%s", upstreamAddress, ba.prefix, name, id), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %s", err)
		}

		req.Header.Add("User-Agent", "facettectl/"+version)

		resp, err := hc.Do(req)
		if err != nil {
			return fmt.Errorf("failed to retrieve data: %s", err)
		}
		defer resp.Body.Close()

		rv := reflect.New(ba.reflectType)
		if err := httputil.BindJSON(resp, rv.Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %s", err)
		}

		data, err := json.Marshal(rv.Interface())
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %s", err)
		}

		// Get modification date or fall back to creation one
		f := reflect.Indirect(rv).FieldByName("Modified")
		if !reflect.DeepEqual(f.Interface(), reflect.Zero(f.Type()).Interface()) {
			mt = *f.Interface().(*time.Time)
		} else {
			mt = reflect.Indirect(rv).FieldByName("Created").Interface().(time.Time)
		}

		// Append data to archive
		if err = w.WriteHeader(&tar.Header{
			Name:    name + "/" + id,
			Size:    int64(len(data)),
			Mode:    0644,
			ModTime: mt,
		}); err != nil {
			return fmt.Errorf("failed to append data: %s", err)
		}

		w.Write(data)
	}

	return nil
}
