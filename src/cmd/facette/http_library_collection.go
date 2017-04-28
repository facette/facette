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
	Parent   string                    `json:"parent,omitempty"`
	Children libraryCollectionTreeList `json:"children,omitempty"`
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
	w.service.backend.Storage().List(&collections, filters, nil, 0, 0)

	for _, c := range collections {
		c.Expand(nil)

		// Fill collection tree
		if _, ok := tree[c.ID]; !ok {
			tree[c.ID] = libraryCollectionToTreeItem(c)
		}

		if c.HasParent() {
			parentID := *c.ParentID

			if _, ok := tree[parentID]; !ok {
				tree[parentID] = libraryCollectionToTreeItem(c.Parent)
			}

			tree[parentID].Children = append(tree[parentID].Children, tree[c.ID])

			if parentID == filters["parent"] {
				tree[c.ID].Parent = ""
			}
		}
	}

	// Only keep top-level collections for result
	result := libraryCollectionTreeList{}
	for _, c := range tree {
		if c.Parent == "" && c.ID != filters["parent"] {
			result = append(result, c)
			sort.Sort(c.Children)
		}
	}

	sort.Sort(result)

	httputil.WriteJSON(rw, result, http.StatusOK)
}

func libraryCollectionToTreeItem(c *backend.Collection) *libraryCollectionTreeEntry {
	entry := &libraryCollectionTreeEntry{
		ID:       c.ID,
		Children: libraryCollectionTreeList{},
	}

	if c.HasParent() {
		entry.Parent = *c.ParentID
	}

	// Use title as label if any or fallback to collection name
	if title, ok := c.Options["title"].(string); ok && title != "" {
		entry.Label = title
	} else {
		entry.Label = c.Name
	}

	return entry
}
