package library

const (
	_ = iota
	// UnitTypeAbsolute represents an absolute unit value type.
	UnitTypeAbsolute
	// UnitTypeDuration represents a duration unit value type.
	UnitTypeDuration
)

// Unit represents a graph value unit.
type Unit struct {
	Item
	Label string `json:"label"`
}
