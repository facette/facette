package backend

import (
	"fmt"

	"github.com/brettlangdon/forge"
)

func initConfig(config *forge.Section) (string, error) {
	var (
		driver string
		err    error
	)

	if backends := config.Keys(); len(backends) == 0 {
		return "", ErrMissingBackendConfig
	} else if len(backends) > 1 {
		return "", ErrMultipleBackendConfig
	} else {
		driver = backends[0]
	}

	// Set configuration defaults
	section := forge.NewSection()

	switch driver {
	case "mysql":
		section.SetString("host", "localhost")
		section.SetInteger("port", 3306)
		section.SetString("dbname", "facette")
		section.SetString("user", "facette")

	case "pgsql":
		section.SetString("host", "localhost")
		section.SetInteger("port", 5432)
		section.SetString("dbname", "facette")
		section.SetString("user", "facette")

	case "sqlite":
		section.SetString("path", "data.db")
	}

	if err = section.Merge(config); err != nil {
		return "", err
	}

	// Only keep driver related section
	section, err = section.GetSection(driver)
	if err != nil {
		return "", fmt.Errorf("failed to get backend driver configuration: %s", err)
	}

	*config = *section

	return driver, nil
}
