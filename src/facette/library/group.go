package library

import (
	"github.com/fatih/goset"
	"log"
	"path"
	"regexp"
	"strings"
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
	var (
		err    error
		group  *Group
		item   interface{}
		re     *regexp.Regexp
		result *set.Set
	)

	if item, err = library.GetItemByName(name, groupType); err != nil {
		log.Printf("ERROR: " + err.Error())
		return []string{}
	}

	group = item.(*Group)

	result = set.New()

	for _, entry := range group.Entries {
		if strings.HasPrefix(entry.Pattern, "regexp:") {
			re = regexp.MustCompile(entry.Pattern[7:])
		}

		if _, ok := library.Catalog.Origins[entry.Origin]; !ok {
			log.Printf("ERROR: unknown `%s' group entry origin", entry.Origin)
			continue
		}

		if groupType == LibraryItemSourceGroup {
			for _, source := range library.Catalog.Origins[entry.Origin].Sources {
				if strings.HasPrefix(entry.Pattern, "glob:") {
					if ok, _ := path.Match(entry.Pattern[5:], source.Name); !ok {
						continue
					}
				} else if strings.HasPrefix(entry.Pattern, "regexp:") {
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
					if strings.HasPrefix(entry.Pattern, "glob:") {
						if ok, _ := path.Match(entry.Pattern[5:], metric.Name); !ok {
							continue
						}
					} else if strings.HasPrefix(entry.Pattern, "regexp:") {
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
