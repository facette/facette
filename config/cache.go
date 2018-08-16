package config

// CacheConfig represents a cache configuration instance.
type CacheConfig struct {
	Path string `yaml:"path"`
}

func newCacheConfig() *CacheConfig {
	return &CacheConfig{
		Path: "var/cache",
	}
}
