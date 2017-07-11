package series

// Response represents a point response instance.
type Response struct {
	Start   string                 `json:"start"`
	End     string                 `json:"end"`
	Series  []ResponseSeries       `json:"series"`
	Options map[string]interface{} `json:"options"`
}

// ResponseSeries represents a point response series instance.
type ResponseSeries struct {
	Series
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}
