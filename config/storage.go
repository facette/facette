package config

import "facette.io/maputil"

func newStorageConfig() *maputil.Map {
	return &maputil.Map{
		"driver": "sqlite",
	}
}
