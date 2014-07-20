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
	Type       int          `json:"type"`
	StackMode  int          `json:"stack_mode"`
	UnitType   int          `json:"unit_type"`
	UnitLegend string       `json:"unit_legend"`
	Groups     []*OperGroup `json:"groups"`
}

func (graph *Graph) String() string {
	return fmt.Sprintf(
		"Graph{ID:\"%s\" Name:\"%s\" Type:%d Groups:[%s]}",
		graph.ID,
		graph.Name,
		graph.Type,
		func(groups []*OperGroup) string {
			groupStrings := make([]string, len(groups))

			for i, group := range groups {
				groupStrings[i] = fmt.Sprintf("%s", group)
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
	Series  []*Serie               `json:"series"`
	Options map[string]interface{} `json:"options"`
}

func (group *OperGroup) String() string {
	return fmt.Sprintf(
		"OperGroup{Name:\"%s\" Type:%d StackID:%d Series:[%s] Options:%v}",
		group.Name,
		group.Type,
		func(series []*Serie) string {
			serieStrings := make([]string, len(series))

			for i, serie := range series {
				serieStrings[i] = fmt.Sprintf("%s", serie)
			}

			return strings.Join(serieStrings, ", ")
		}(group.Series),
		group.Options,
	)
}

// Serie represents a serie entry.
type Serie struct {
	Name    string                 `json:"name"`
	Origin  string                 `json:"origin"`
	Source  string                 `json:"source"`
	Metric  string                 `json:"metric"`
	Options map[string]interface{} `json:"options"`
}

func (serie *Serie) String() string {
	return fmt.Sprintf(
		"Serie{Name:\"%s\" Origin:\"%s\" Source:\"%s\" Metric:\"%s\" Options:%v}",
		serie.Name,
		serie.Origin,
		serie.Source,
		serie.Metric,
		serie.Options,
	)
}
