package library

import (
	"strings"

	"github.com/facette/facette/pkg/config"
)

// Collection represents a collection of graphs.
type Collection struct {
	Item
	Entries  []*CollectionEntry     `json:"entries"`
	Parent   *Collection            `json:"-"`
	ParentID string                 `json:"parent"`
	Options  map[string]interface{} `json:"options"`
	Children []*Collection          `json:"-"`
}

// IndexOfChild returns the index of a child in the children list of `-1' if not found.
func (collection *Collection) IndexOfChild(id string) int {
	for index, entry := range collection.Children {
		if entry.ID == id {
			return index
		}
	}

	return -1
}

// CollectionEntry represents a collection entry.
type CollectionEntry struct {
	ID      string                 `json:"id"`
	Options map[string]interface{} `json:"options"`
}

// PrepareCollection applies options defaults and fallback values, then filters entries by graphs titles and state.
func (library *Library) PrepareCollection(collection *Collection, filter string) *Collection {
	collectionTemp := &Collection{}
	*collectionTemp = *collection
	collectionTemp.Entries = nil

	refreshInterval, _ := config.GetInt(collectionTemp.Options, "refresh_interval", false)

	for _, entry := range collection.Entries {
		item, err := library.GetItem(entry.ID, LibraryItemGraph)
		if err != nil {
			continue
		}

		graph := item.(*Graph)

		// Initialize options map if needed, then set defaults
		if entry.Options == nil {
			entry.Options = make(map[string]interface{})
			entry.Options["enabled"] = true
		}

		// Check for linked graph
		if graph.Link != "" {
			entry.Options["linked"] = true
		}

		// Retrieve missing title from graph if none provided
		if title, ok := entry.Options["title"]; !ok || title == nil {
			if graph.Title != "" {
				entry.Options["title"] = graph.Title
			} else {
				entry.Options["title"] = graph.Name
			}
		}

		// Get global refresh interval if none provided
		if refreshInterval > 0 {
			if _, err := config.GetInt(entry.Options, "refresh_interval", true); err != nil {
				entry.Options["refresh_interval"] = refreshInterval
			}
		}

		if enabled, err := config.GetBool(entry.Options, "enabled", false); err != nil || !enabled {
			continue
		} else if filter != "" {
			if title, err := config.GetString(entry.Options, "title", false); err != nil ||
				!strings.Contains(strings.ToLower(title), strings.ToLower(filter)) {
				continue
			}
		}

		collectionTemp.Entries = append(collectionTemp.Entries, entry)
	}

	return collectionTemp
}
