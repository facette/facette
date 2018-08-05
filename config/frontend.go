package config

// FrontendConfig represents a front-end configuration instance.
type FrontendConfig struct {
	Enabled   bool   `yaml:"enabled"`
	AssetsDir string `yaml:"assets_dir"`
}
