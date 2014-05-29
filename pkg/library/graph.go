package library

const (
	_ = iota
	// GraphTypeArea represents an area graph type.
	GraphTypeArea
	// GraphTypeLine represents a line graph type.
	GraphTypeLine
)

const (
	_ = iota
	// StackModeNone represents a null stack mode.
	StackModeNone
	// StackModeNormal represents a normal stack mode.
	StackModeNormal
	// StackModePercent represents a percentage stack mode.
	StackModePercent
)

// Graph represents a graph containing list of series.
type Graph struct {
	Item
	Type      int      `json:"type"`
	StackMode int      `json:"stack_mode"`
	Stacks    []*Stack `json:"stacks"`
	Volatile  bool     `json:"-"`
}

// Stack represents a set of operation group entries.
type Stack struct {
	Name   string       `json:"name"`
	Groups []*OperGroup `json:"groups"`
}

// OperGroup represents an operation group entry.
type OperGroup struct {
	Name    string                 `json:"name"`
	Type    int                    `json:"type"`
	Series  []*Serie               `json:"series"`
	Scale   float64                `json:"scale"`
	Options map[string]interface{} `json:"options"`
}

// Serie represents a serie entry.
type Serie struct {
	Name   string  `json:"name"`
	Origin string  `json:"origin"`
	Source string  `json:"source"`
	Metric string  `json:"metric"`
	Scale  float64 `json:"scale"`
}
