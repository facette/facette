package types

// OriginResponse represents an origin response struct in the server catalog backend.
type OriginResponse struct {
	Name    string `json:"name"`
	Backend string `json:"backend"`
	Updated string `json:"updated"`
}

// SourceResponse represents a source response struct in the server catalog backend.
type SourceResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Updated string   `json:"updated"`
}

// MetricResponse represents a metric response struct in the server catalog backend.
type MetricResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Sources []string `json:"sources"`
	Updated string   `json:"updated"`
}
