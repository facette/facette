package main

import "github.com/brettlangdon/forge"

func initConfig(path string) (*forge.Section, error) {
	var (
		config *forge.Section
		err    error
	)

	if path != "" {
		config, err = forge.ParseFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		config = forge.NewSection()
	}

	root := forge.NewSection()
	root.SetString("listen", "localhost:12003")
	root.SetInteger("graceful_timeout", 30)
	root.SetString("log_path", "")
	root.SetString("log_level", "info")

	frontend := root.AddSection("frontend")
	frontend.SetBoolean("enabled", true)
	frontend.SetString("assets_dir", "assets")

	backend := root.AddSection("backend")
	if section, err := config.GetSection("backend"); err == nil && len(section.Keys()) == 0 {
		backend.AddSection("sqlite")
	}

	if len(config.Keys()) > 0 {
		if err = root.Merge(config); err != nil {
			return nil, err
		}
	}

	return root, nil
}
