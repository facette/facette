package config

// APIConfig represents the API definition in the configuration system.
type APIConfig struct {
	ReadOnly      bool `json:"read_only"`
	DisableReload bool `json:"disable_reload"`
}
