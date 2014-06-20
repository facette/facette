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
	Type      int          `json:"type"`
	StackMode int          `json:"stack_mode"`
	Groups    []*OperGroup `json:"groups"`
}

// OperGroup represents an operation group entry.
type OperGroup struct {
	Name    string                 `json:"name"`
	Type    int                    `json:"type"`
	StackID int                    `json:"stack_id"`
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

func (graph *Graph) String() string {
	return fmt.Sprintf(
		"Graph{ID:\"%s\" Name:\"%s\" Type:%d Groups:[%s]}",
		graph.Name,
		graph.ID,
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

func (group *OperGroup) String() string {
	return fmt.Sprintf(
		"OperGroup{Name:\"%s\" Type:%d Scale:%f Series:[%s]}",
		group.Name,
		group.Type,
		group.Scale,
		func(series []*Serie) string {
			serieStrings := make([]string, len(series))

			for i, serie := range series {
				serieStrings[i] = fmt.Sprintf("%s", serie)
			}

			return strings.Join(serieStrings, ", ")
		}(group.Series),
	)
}

func (serie *Serie) String() string {
	return fmt.Sprintf(
		"Serie{Name:\"%s\" Origin:\"%s\" Source:\"%s\" Metric:\"%s\" Scale:%f}",
		serie.Name,
		serie.Origin,
		serie.Source,
		serie.Metric,
		serie.Scale,
	)
}
