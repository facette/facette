package library

import (
	"facette/common"
	"facette/utils"
	"fmt"
	"github.com/fatih/set"
	"os"
	"regexp"
	"sort"
	"strings"
)

// CollectionEntry represents a Collection entry structure.
type CollectionEntry struct {
	ID      string            `json:"id"`
	Options map[string]string `json:"options"`
}

// Collection represents a Collection entry in a Library.
type Collection struct {
	Item
	Entries  []*CollectionEntry `json:"entries"`
	Parent   *Collection        `json:"-"`
	Children []*Collection      `json:"-"`
}

// FilterCollection filters collection entries by graphs titles.
func (library *Library) FilterCollection(collection *Collection, filter string) *Collection {
	var (
		collectionTemp *Collection
	)

	if filter == "" {
		return nil
	}

	collectionTemp = &Collection{}
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

// GetCollectionTemplate generates a Collection based on origins templates.
func (library *Library) GetCollectionTemplate(name string) (*Collection, error) {
	var (
		chunks        []string
		collection    *Collection
		count         int
		found         bool
		metricSet     *set.Set
		options       map[string]string
		pattern       string
		patternMatch  bool
		patternRegexp *regexp.Regexp
		splitItems    []string
		splitSet      *set.Set
		template      *common.TemplateConfig
		templates     []string
	)

	collection = &Collection{Item: Item{Name: name}}

	for originName, origin := range library.Catalog.Origins {
		if _, ok := origin.Sources[name]; !ok {
			continue
		}

		found = true

		// Get sorted templates list
		for templateName := range library.Config.Origins[originName].Templates {
			templates = append(templates, templateName)
		}

		sort.Strings(templates)

		// Prepare metrics
		metricSet = set.New()

		for metricName := range library.Catalog.Origins[originName].Sources[name].Metrics {
			metricSet.Add(metricName)
		}

		// Parse template entries
		for _, templateName := range templates {
			template = library.Config.Origins[originName].Templates[templateName]

			if template.SplitPattern != "" {
				splitSet = set.New()

				for metricName := range library.Catalog.Origins[originName].Sources[name].Metrics {
					if chunks = template.SplitRegexp.FindStringSubmatch(metricName); len(chunks) != 2 {
						continue
					}

					metricSet.Remove(metricName)
					splitSet.Add(chunks[1])
				}

				splitItems = splitSet.StringSlice()
				sort.Strings(splitItems)

				for _, itemName := range splitItems {
					options = make(map[string]string)

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
				pattern = ""
				patternMatch = false

				for _, stackItem := range template.Stacks {
					for _, groupItem := range stackItem.Groups {
						if pattern != "" {
							pattern += "|"
						}

						pattern += groupItem.Pattern
					}
				}

				patternRegexp = regexp.MustCompile(pattern)

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

				options = make(map[string]string)

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
		for _, metricName := range metricSet.StringSlice() {
			options = make(map[string]string)
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
