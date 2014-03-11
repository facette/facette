package library

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
)

// Collection represents a collection of graphs.
type Collection struct {
	Item
	Entries  []*CollectionEntry `json:"entries"`
	Parent   *Collection        `json:"-"`
	ParentID string             `json:"parent"`
	Children []*Collection      `json:"-"`
}

// CollectionEntry represents a collection entry.
type CollectionEntry struct {
	ID      string            `json:"id"`
	Options map[string]string `json:"options"`
}

// FilterCollection filters collection entries by graphs titles.
func (library *Library) FilterCollection(collection *Collection, filter string) *Collection {
	if filter == "" {
		return nil
	}

	collectionTemp := &Collection{}
	*collectionTemp = *collection
	collectionTemp.Entries = nil

	for _, entry := range collection.Entries {
		if _, ok := entry.Options["title"]; !ok {
			continue
		} else if !strings.Contains(strings.ToLower(entry.Options["title"]), strings.ToLower(filter)) {
			continue
		}

		collectionTemp.Entries = append(collectionTemp.Entries, entry)
	}

	return collectionTemp
}

// GetCollectionTemplate generates a collection based on origins templates.
func (library *Library) GetCollectionTemplate(name string) (*Collection, error) {
	found := false

	collection := &Collection{Item: Item{Name: name}}

	for originName, origin := range library.Catalog.Origins {
		if _, ok := origin.Sources[name]; !ok {
			continue
		}

		count := 0
		found = true

		// Get sorted templates list
		templates := make([]string, 0)

		for templateName := range library.Config.Origins[originName].Templates {
			templates = append(templates, templateName)
		}

		sort.Strings(templates)

		// Prepare metrics
		metricSet := set.New()

		for metricName := range library.Catalog.Origins[originName].Sources[name].Metrics {
			metricSet.Add(metricName)
		}

		// Parse template entries
		for _, templateName := range templates {
			template := library.Config.Origins[originName].Templates[templateName]

			if template.SplitPattern != "" {
				splitSet := set.New()

				for metricName := range library.Catalog.Origins[originName].Sources[name].Metrics {
					chunks := template.SplitRegexp.FindStringSubmatch(metricName)
					if len(chunks) != 2 {
						continue
					}

					metricSet.Remove(metricName)
					splitSet.Add(chunks[1])
				}

				splitItems := set.StringSlice(splitSet)
				sort.Strings(splitItems)

				for _, itemName := range splitItems {
					options := make(map[string]string)

					if template.Options != nil {
						utils.Clone(template.Options, &options)
					}

					options["origin"] = originName
					options["source"] = name
					options["template"] = templateName
					options["filter"] = itemName

					if options["title"] != "" {
						options["title"] = strings.Replace(options["title"], "%s", itemName, 1)
					}

					collection.Entries = append(collection.Entries, &CollectionEntry{
						ID:      fmt.Sprintf("unnamed%d", count),
						Options: options,
					})

					count += 1
				}
			} else {
				pattern := ""
				patternMatch := false

				for _, stackItem := range template.Stacks {
					for _, groupItem := range stackItem.Groups {
						if pattern != "" {
							pattern += "|"
						}

						pattern += groupItem.Pattern
					}
				}

				patternRegexp := regexp.MustCompile(pattern)

				for metricName := range library.Catalog.Origins[originName].Sources[name].Metrics {
					if !patternRegexp.MatchString(metricName) {
						continue
					}

					metricSet.Remove(metricName)
					patternMatch = true
				}

				if !patternMatch {
					continue
				}

				options := make(map[string]string)

				if template.Options != nil {
					utils.Clone(template.Options, &options)
				}

				options["origin"] = originName
				options["source"] = name
				options["template"] = templateName

				collection.Entries = append(collection.Entries, &CollectionEntry{
					ID:      fmt.Sprintf("unnamed%d", count),
					Options: options,
				})

				count += 1
			}
		}

		// Handle non-template metrics
		for _, metricName := range set.StringSlice(metricSet) {
			options := make(map[string]string)
			options["origin"] = originName
			options["source"] = name
			options["metric"] = metricName
			options["title"] = metricName

			collection.Entries = append(collection.Entries, &CollectionEntry{
				ID:      fmt.Sprintf("unnamed%d", count),
				Options: options,
			})

			count += 1
		}
	}

	if !found {
		return nil, os.ErrNotExist
	}

	return collection, nil
}
