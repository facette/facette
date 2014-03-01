package library

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/facette/facette/thirdparty/github.com/nu7hatch/gouuid"
)

const (
	// GraphTypeArea represents an area graph type.
	GraphTypeArea = iota
	// GraphTypeLine represents a line graph type.
	GraphTypeLine
)

const (
	// StackModeNone represents a null stack mode.
	StackModeNone = iota
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

// GetGraphMetric gets a graph metric item.
func (library *Library) GetGraphMetric(origin, source, metric string) (*Graph, error) {
	if _, ok := library.Catalog.Origins[origin]; !ok {
		return nil, fmt.Errorf("unknown `%s' origin", origin)
	} else if _, ok := library.Catalog.Origins[origin].Sources[source]; !ok {
		return nil, fmt.Errorf("unknown `%s' source for `%s' origin", source, origin)
	} else if _, ok := library.Catalog.Origins[origin].Sources[source].Metrics[metric]; !ok {
		return nil, fmt.Errorf("unknown `%s' metric for `%s' source", metric, source)
	}

	return &Graph{
		Item: Item{ID: origin + "\x30" + metric, Name: metric},
		Stacks: []*Stack{&Stack{
			Name: "stack0",
			Groups: []*OperGroup{&OperGroup{
				Name: metric,
				Series: []*Serie{&Serie{
					Name:   metric,
					Origin: origin,
					Source: source,
					Metric: metric,
				}},
			}},
		}},
	}, nil
}

// GetGraphTemplate gets a graph template item.
func (library *Library) GetGraphTemplate(origin, source, template, filter string) (*Graph, error) {
	id := origin + "\x30" + template + "\x30" + filter

	if _, ok := library.Config.Origins[origin]; !ok {
		return nil, fmt.Errorf("unknown `%s' origin", origin)
	} else if _, ok := library.Config.Origins[origin].Templates[template]; !ok {
		return nil, fmt.Errorf("unknown `%s' template for `%s' origin", template, origin)
	}

	// Load template from filesystem if needed
	if !library.ItemExists(id, LibraryItemGraphTemplate) {
		graph := &Graph{
			Item:      Item{Name: template, Modified: library.Config.Origins[origin].Modified},
			StackMode: library.Config.Origins[origin].Templates[template].StackMode,
		}

		for i, tmplStack := range library.Config.Origins[origin].Templates[template].Stacks {
			stack := &Stack{Name: fmt.Sprintf("stack%d", i)}

			for groupName, tmplGroup := range tmplStack.Groups {
				var re *regexp.Regexp

				if filter != "" {
					re = regexp.MustCompile(strings.Replace(tmplGroup.Pattern, "%s", regexp.QuoteMeta(filter), 1))
				} else {
					re = regexp.MustCompile(tmplGroup.Pattern)
				}

				group := &OperGroup{Name: groupName, Type: tmplGroup.Type}

				for metricName := range library.Catalog.Origins[origin].Sources[source].Metrics {
					if !re.MatchString(metricName) {
						continue
					}

					group.Series = append(group.Series, &Serie{
						Name:   metricName,
						Origin: origin,
						Source: source,
						Metric: metricName,
					})
				}

				if len(group.Series) == 1 {
					group.Series[0].Name = group.Name
				}

				stack.Groups = append(stack.Groups, group)
			}

			graph.Stacks = append(graph.Stacks, stack)
		}

		graph.ID = id
		library.TemplateGraphs[id] = graph
	}

	return library.TemplateGraphs[id], nil
}

func (library *Library) getTemplateID(origin, name string) (string, error) {
	id, err := uuid.NewV3(uuid.NamespaceURL, []byte(origin+name))
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
