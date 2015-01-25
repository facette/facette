package library

import (
	"fmt"
	"strings"
)

const (
	_ = iota
	// GraphTypeArea represents an area graph type.
	GraphTypeArea
	// GraphTypeLine represents a line graph type.
	GraphTypeLine
)

const (
	_ = iota
	// GraphUnitTypeFixed represents a fixed unit value type.
	GraphUnitTypeFixed
	// GraphUnitTypeMetric represents a metric system unit value type.
	GraphUnitTypeMetric
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
	Template bool `json:"template"`

	// For plain/template graphs definitions
	Title      string       `json:"title"`
	Type       int          `json:"type"`
	StackMode  int          `json:"stack_mode"`
	UnitType   int          `json:"unit_type"`
	UnitLegend string       `json:"unit_legend"`
	Groups     []*OperGroup `json:"groups,omitempty"`

	// For linked graphs
	Link       string                 `json:"link,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func (graph *Graph) String() string {
	if graph.Link != "" {
		return fmt.Sprintf(
			"Graph{ID:%q Link:%s Attributes:%+v}",
			graph.ID,
			graph.Link,
			graph.Attributes,
		)
	}

	return fmt.Sprintf(
		"Graph{ID:%q Name:%q Title:%q Type:%d Groups:[%s]}",
		graph.ID,
		graph.Name,
		graph.Title,
		graph.Type,
		func(groups []*OperGroup) string {
			groupStrings := make([]string, len(groups))
			for i, entry := range groups {
				groupStrings[i] = fmt.Sprintf("%s", entry)
			}

			return strings.Join(groupStrings, ", ")
		}(graph.Groups),
	)
}

// OperGroup represents an operation group entry.
type OperGroup struct {
	Name    string                 `json:"name"`
	Type    int                    `json:"type"`
	StackID int                    `json:"stack_id"`
	Series  []*Series              `json:"series"`
	Options map[string]interface{} `json:"options"`
}

func (group *OperGroup) String() string {
	return fmt.Sprintf(
		"OperGroup{Name:%q Type:%d StackID:%d Series:[%s] Options:%v}",
		group.Name,
		group.Type,
		func(series []*Series) string {
			seriesStrings := make([]string, len(series))
			for i, entry := range series {
				seriesStrings[i] = fmt.Sprintf("%s", entry)
			}

			return strings.Join(seriesStrings, ", ")
		}(group.Series),
		group.Options,
	)
}

// Series represents a series entry.
type Series struct {
	Name    string                 `json:"name"`
	Origin  string                 `json:"origin"`
	Source  string                 `json:"source"`
	Metric  string                 `json:"metric"`
	Options map[string]interface{} `json:"options"`
}

func (series *Series) String() string {
	return fmt.Sprintf(
		"Series{Name:%q Origin:%q Source:%q Metric:%q Options:%v}",
		series.Name,
		series.Origin,
		series.Source,
		series.Metric,
		series.Options,
	)
}
