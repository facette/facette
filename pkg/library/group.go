package library

import (
	"sort"
	"strings"

	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/utils"
)

const (
	// LibraryGroupPrefix represents the prefix for sources and metrics groups names.
	LibraryGroupPrefix = "group:"
)

// Group represents a source or metric group.
type Group struct {
	Item
	Type    int           `json:"-"`
	Entries []*GroupEntry `json:"entries"`
}

// GroupEntry represents a group entry.
type GroupEntry struct {
	Pattern string `json:"pattern"`
	Origin  string `json:"origin"`
}

// ExpandSourceGroup expands a source group returning a list of matching items.
func (library *Library) ExpandSourceGroup(name string) []string {
	return library.expandGroup(name, LibraryItemSourceGroup, "")
}

// ExpandMetricGroup expands a metric group returning a list of matching items.
func (library *Library) ExpandMetricGroup(sourceName, name string) []string {
	return library.expandGroup(name, LibraryItemMetricGroup, sourceName)
}

func (library *Library) expandGroup(name string, groupType int, sourceName string) []string {
	item, err := library.GetItemByName(name, groupType)
	if err != nil {
		logger.Log(logger.LevelError, "library", "expand group: unknown group `%s': %s", name, err)
		return []string{}
	}

	// Parse group entries for patterns
	group := item.(*Group)
	result := []string{}

	for _, entry := range group.Entries {
		subResult := []string{}

		if groupType == LibraryItemSourceGroup {
			origin, err := library.Catalog.GetOrigin(entry.Origin)
			if err != nil {
				logger.Log(logger.LevelError, "library", "%s", err)
				continue
			}

			for _, source := range origin.GetSources() {
				if utils.FilterMatch(entry.Pattern, source.Name) {
					subResult = append(subResult, source.Name)
				}
			}
		} else {
			source, err := library.Catalog.GetSource(entry.Origin, sourceName)
			if err != nil {
				logger.Log(logger.LevelError, "library", "%s", err)
				continue
			}

			for _, metric := range source.GetMetrics() {
				if utils.FilterMatch(entry.Pattern, metric.Name) {
					subResult = append(subResult, metric.Name)
				}
			}
		}

		// Preserve manual ordering if grouped with `Single' matching type
		if strings.HasPrefix(entry.Pattern, "glob:") || strings.HasPrefix(entry.Pattern, "regexp:") {
			sort.Strings(subResult)
		}

		// Merge all group items subresults
		result = append(result, subResult...)
	}

	return result
}
