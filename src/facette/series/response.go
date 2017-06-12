package series

// Response represents a point response instance.
type Response struct {
	Start   string                 `json:"start"`
	End     string                 `json:"end"`
	Series  []SeriesResponse       `json:"series"`
	Options map[string]interface{} `json:"options"`
}

// SeriesResponse represents a point response series instance.
type SeriesResponse struct {
	Series
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}
