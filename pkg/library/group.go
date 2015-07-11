package library

import (
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/facette/facette/pkg/logger"
	"github.com/fatih/set"
)

const (
	// LibraryGroupPrefix represents the prefix for sources and metrics groups names.
	LibraryGroupPrefix = "group:"
	// LibraryMatchPrefixGlob represents the prefix for glob matching patterns.
	LibraryMatchPrefixGlob = "glob:"
	// LibraryMatchPrefixRegexp represents the prefix for regexp matching patterns.
	LibraryMatchPrefixRegexp = "regexp:"
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

	// Launch expansion goroutine
	itemSet := set.New(set.ThreadSafe)
	itemChan := make(chan [2]string)
	itemWg := &sync.WaitGroup{}

	go func(itemSet set.Interface, itemChan chan [2]string, itemWg *sync.WaitGroup) {
		var re *regexp.Regexp

		itemWg.Add(1)

		for entry := range itemChan {

			if strings.HasPrefix(entry[0], LibraryMatchPrefixGlob) {
				if ok, _ := path.Match(strings.TrimPrefix(entry[0], LibraryMatchPrefixGlob), entry[1]); !ok {
					continue
				}
			} else if strings.HasPrefix(entry[0], LibraryMatchPrefixRegexp) {
				re = regexp.MustCompile(strings.TrimPrefix(entry[0], LibraryMatchPrefixRegexp))

				if !re.MatchString(entry[1]) {
					continue
				}
			} else if entry[0] != entry[1] {
				continue
			}

			itemSet.Add(entry[1])
		}

		itemWg.Done()
	}(itemSet, itemChan, itemWg)

	// Parse group entries for patterns
	group := item.(*Group)

	for _, entry := range group.Entries {
		if groupType == LibraryItemSourceGroup {
			origin, err := library.Catalog.GetOrigin(entry.Origin)
			if err != nil {
				logger.Log(logger.LevelError, "library", "%s", err)
				continue
			}

			for _, source := range origin.GetSources() {
				itemChan <- [2]string{entry.Pattern, source.Name}
			}
		} else {
			source, err := library.Catalog.GetSource(entry.Origin, sourceName)
			if err != nil {
				logger.Log(logger.LevelError, "library", "%s", err)
				continue
			}

			for _, metric := range source.GetMetrics() {
				itemChan <- [2]string{entry.Pattern, metric.Name}
			}
		}
	}

	close(itemChan)

	itemWg.Wait()

	result := set.StringSlice(itemSet)
	sort.Strings(result)

	return result
}
