package library

import "strings"

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
