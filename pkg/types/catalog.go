package types

// OriginResponse represents an origin response struct in the server catalog.
type OriginResponse struct {
	Name      string `json:"name"`
	Connector string `json:"connector"`
	Updated   string `json:"updated"`
}

// SourceResponse represents a source response struct in the server catalog.
type SourceResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Updated string   `json:"updated"`
}

// MetricResponse represents a metric response struct in the server catalog.
type MetricResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Sources []string `json:"sources"`
	Updated string   `json:"updated"`
}
