package main

import (
	"fmt"

	"facette/backend"

	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
)

type catalogCommand struct {
	Enable bool

	Refresh struct {
		Enable bool
	} `usage:"Refresh catalog data" expand:"1"`
}

func execCatalog() error {
	if cmd.Catalog.Refresh.Enable {
		return execCatalogRefresh()
	}

	set, _ := flagSet.FindSubset("catalog")
	set.Help(false)

	return nil
}

func execCatalogRefresh() error {
	var errored bool

	providers := []backend.Provider{}
	if err := apiRequest("GET", "/providers/?fields=id,name", nil, nil, &providers); err != nil {
		return errors.Wrap(err, "failed to retrieve providers")
	}

	for _, prov := range providers {
		if !cmd.Quiet {
			fmt.Printf("refreshing %q provider...\n", prov.Name)
		}

		if err := apiRequest("POST", fmt.Sprintf("/providers/%s/refresh", prov.ID), nil, nil, nil); err != nil {
			printError("failed to refresh provider: %s", err)
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
