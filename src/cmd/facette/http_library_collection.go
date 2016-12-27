package main

import (
	"context"
	"net/http"
	"sort"

	"facette/backend"

	"github.com/facette/httputil"
	"github.com/facette/natsort"
)

type libraryCollectionTreeEntry struct {
	ID       string                    `json:"id"`
	Label    string                    `json:"label"`
	Parent   string                    `json:"parent"`
	Children libraryCollectionTreeList `json:"children"`
}

type libraryCollectionTreeList []*libraryCollectionTreeEntry

func (l libraryCollectionTreeList) Len() int {
	return len(l)
}

func (l libraryCollectionTreeList) Less(i, j int) bool {
	return natsort.Compare(l[i].Label, l[j].Label)
}

func (l libraryCollectionTreeList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (w *httpWorker) httpHandleLibraryCollectionTree(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	tree := map[string]*libraryCollectionTreeEntry{}

	// Handle request filters
	filters := map[string]interface{}{
		"template": false,
	}

	if p := r.URL.Query().Get("parent"); p != "" {
		filters["parent"] = p
	}

	// Fetch non-template collections list
	collections := []*backend.Collection{}
	w.service.backend.List(&collections, filters, nil, 0, 0)

	for _, c := range collections {
		// Expand collection data
		if c.Link != nil && len(c.Link.Options) > 0 {
			c.Link.Options.Merge(c.Options, true)
			c.Options = c.Link.Options
		}
		c.Expand(c.Attributes, w.service.backend)

		// Fill collection tree
		if _, ok := tree[c.ID]; !ok {
			tree[c.ID] = libraryCollectionToTreeItem(c)
		}

		if c.ParentID != "" {
			if _, ok := tree[c.ParentID]; !ok {
				tree[c.ParentID] = libraryCollectionToTreeItem(c.Parent)
			}

			tree[c.ParentID].Children = append(tree[c.ParentID].Children, tree[c.ID])
		}
	}

	// Only keep top-level collections for result
	result := libraryCollectionTreeList{}
	for _, c := range tree {
		if c.Parent == "" {
			result = append(result, c)
			sort.Sort(c.Children)
		}
	}

	sort.Sort(result)

	httputil.WriteJSON(rw, result, http.StatusOK)
}

func libraryCollectionToTreeItem(collection *backend.Collection) *libraryCollectionTreeEntry {
	entry := &libraryCollectionTreeEntry{
		ID:       collection.ID,
		Parent:   collection.ParentID,
		Children: libraryCollectionTreeList{},
	}

	// Use title as label if any or fallback to collection name
	if title, ok := collection.Options["title"].(string); ok {
		entry.Label = title
	} else {
		entry.Label = collection.Name
	}

	return entry
}
