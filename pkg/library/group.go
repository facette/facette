package library

import (
	"log"
	"path"
	"regexp"
	"strings"

	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

const (
	// LibraryGroupPrefix represents the prefix for sources and metrics groups names.
	LibraryGroupPrefix = "group:"
	// LibraryMatchPrefixGlob represents the prefix for glob matching patterns.
	LibraryMatchPrefixGlob = "glob:"
	// LibraryMatchPrefixRegexp represents the prefix for regexp matching patterns.
	LibraryMatchPrefixRegexp = "regexp:"
)

// GroupEntry represents a subset of a Group entry.
type GroupEntry struct {
	Pattern string `json:"pattern"`
	Origin  string `json:"origin"`
}

// Group represents an source/metric group structure.
type Group struct {
	Item
	Type    int           `json:"-"`
	Entries []*GroupEntry `json:"entries"`
}

// ExpandGroup returns the list of items matching the group name based on its groupType.
func (library *Library) ExpandGroup(name string, groupType int) []string {

	item, err := library.GetItemByName(name, groupType)
	if err != nil {
		log.Printf("ERROR: " + err.Error())
		return []string{}
	}

	group := item.(*Group)

	result := set.New()

	for _, entry := range group.Entries {
		var re *regexp.Regexp

		if strings.HasPrefix(entry.Pattern, LibraryMatchPrefixRegexp) {
			re = regexp.MustCompile(strings.TrimPrefix(entry.Pattern, LibraryMatchPrefixRegexp))
		}

		if _, ok := library.Catalog.Origins[entry.Origin]; !ok {
			log.Printf("ERROR: unknown `%s' group entry origin", entry.Origin)
			continue
		}

		if groupType == LibraryItemSourceGroup {
			for _, source := range library.Catalog.Origins[entry.Origin].Sources {
				if strings.HasPrefix(entry.Pattern, LibraryMatchPrefixGlob) {
					if ok, _ := path.Match(strings.TrimPrefix(entry.Pattern, LibraryMatchPrefixGlob),
						source.Name); !ok {
						continue
					}
				} else if strings.HasPrefix(entry.Pattern, LibraryMatchPrefixRegexp) {
					if !re.MatchString(source.Name) {
						continue
					}
				} else if entry.Pattern != source.Name {
					continue
				}

				result.Add(source.Name)
			}
		} else if groupType == LibraryItemMetricGroup {
			for _, source := range library.Catalog.Origins[entry.Origin].Sources {
				for _, metric := range source.Metrics {
					if strings.HasPrefix(entry.Pattern, LibraryMatchPrefixGlob) {
						if ok, _ := path.Match(strings.TrimPrefix(entry.Pattern, LibraryMatchPrefixGlob),
							metric.Name); !ok {
							continue
						}
					} else if strings.HasPrefix(entry.Pattern, LibraryMatchPrefixRegexp) {
						if !re.MatchString(metric.Name) {
							continue
						}
					} else if entry.Pattern != metric.Name {
						continue
					}

					result.Add(metric.Name)
				}
			}
		}
	}

	return result.StringSlice()
}
