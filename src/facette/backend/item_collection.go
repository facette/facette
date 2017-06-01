package backend

import (
	"facette/template"
	"sort"

	"github.com/facette/maputil"
	"github.com/facette/natsort"
	"github.com/facette/sliceutil"
	"github.com/jinzhu/gorm"
)

// Collection represents a back-end collection item instance.
type Collection struct {
	Item
	Entries    []*CollectionEntry `json:"entries,omitempty"`
	Link       *Collection        `json:"-"`
	LinkID     *string            `gorm:"column:link;type:varchar(36) DEFAULT NULL REFERENCES collections (id) ON DELETE CASCADE ON UPDATE CASCADE" json:"link,omitempty"`
	Attributes maputil.Map        `gorm:"type:text" json:"attributes,omitempty"`
	Alias      *string            `gorm:"type:varchar(128);unique_index" json:"alias,omitempty"`
	Options    maputil.Map        `gorm:"type:text" json:"options,omitempty"`
	Parent     *Collection        `json:"-"`
	ParentID   *string            `gorm:"column:parent;type:varchar(36) DEFAULT NULL REFERENCES collections (id) ON DELETE SET NULL ON UPDATE SET NULL" json:"parent,omitempty"`
	Template   bool               `gorm:"not null" json:"template"`

	resolved bool
	expanded bool
}

// NewCollection creates a new collection item instance.
func (b *Backend) NewCollection() *Collection {
	return &Collection{Item: Item{backend: b}}
}

// BeforeSave handles the ORM 'BeforeSave' callback.
func (c *Collection) BeforeSave(scope *gorm.Scope) error {
	if err := c.Item.BeforeSave(scope); err != nil {
		return err
	} else if c.Alias != nil && *c.Alias != "" && !nameRegexp.MatchString(*c.Alias) {
		return ErrInvalidAlias
	}

	for idx, entry := range c.Entries {
		entry.Index = idx + 1
	}

	// Ensure optional fields are null if empty
	if c.LinkID != nil && *c.LinkID == "" {
		scope.SetColumn("LinkID", nil)
	}

	if c.Alias != nil && *c.Alias == "" {
		scope.SetColumn("Alias", nil)
	}

	if c.ParentID != nil && *c.ParentID == "" {
		scope.SetColumn("ParentID", nil)
	}

	return nil
}

// Clone returns a clone of the collection item instance.
func (c *Collection) Clone() *Collection {
	clone := &Collection{}
	*clone = *c

	clone.Entries = make([]*CollectionEntry, len(c.Entries))
	for i, entry := range c.Entries {
		clone.Entries[i] = &CollectionEntry{}
		*clone.Entries[i] = *entry

		if entry.Options != nil {
			clone.Entries[i].Options = entry.Options.Clone()
		}
	}

	if c.Attributes != nil {
		clone.Attributes = c.Attributes.Clone()
	}

	if c.Options != nil {
		clone.Options = c.Options.Clone()
	}

	if c.Link != nil {
		clone.Link = &Collection{}
		*clone.Link = *c.Link
	}

	if c.Parent != nil {
		clone.Parent = &Collection{}
		*clone.Parent = *c.Parent
	}

	return clone
}

// Expand expands the collection item instance using its linked instance.
func (c *Collection) Expand(attrs maputil.Map) error {
	var err error

	if c.expanded {
		return nil
	}

	if len(attrs) > 0 {
		c.Attributes.Merge(attrs, true)
	}

	if c.backend != nil && c.LinkID != nil && *c.LinkID != "" {
		if err := c.Resolve(nil); err != nil {
			return err
		}

		// Expand template and applies current collection's attributes and options
		tmpl := c.Link.Clone()
		tmpl.ID = c.ID
		tmpl.Attributes.Merge(c.Attributes, true)
		tmpl.Options.Merge(c.Options, true)
		tmpl.Template = false

		if c.ParentID != nil && *c.ParentID != "" {
			tmpl.Parent = c.Parent.Clone()
			tmpl.ParentID = c.ParentID
		}

		if title, ok := tmpl.Options["title"].(string); ok {
			if tmpl.Options["title"], err = template.Expand(title, tmpl.Attributes); err != nil {
				return err
			}
		}

		*c = *tmpl
	}

	if len(c.Entries) > 0 {
		if !c.resolved {
			if err := c.Resolve(nil); err != nil {
				return err
			}
		}

		for _, entry := range c.Entries {
			attrs := maputil.Map{}
			attrs.Merge(c.Attributes, true)
			attrs.Merge(entry.Attributes, true)

			entry.Graph.backend = c.backend
			entry.Graph.Expand(attrs)

			if v, ok := entry.Graph.Options["title"]; ok {
				if entry.Options == nil {
					entry.Options = maputil.Map{}
				}
				entry.Options["title"] = v
			}
		}
	}

	c.expanded = true

	return nil
}

// HasParent returns whether or not the collection item has a parent instance.
func (c *Collection) HasParent() bool {
	return c.ParentID != nil && *c.ParentID != ""
}

// Resolve resolves both the collection item linked and parent instances.
func (c *Collection) Resolve(cache map[string]*Collection) error {
	if c.resolved {
		return nil
	} else if c.backend == nil {
		return ErrUnresolvableItem
	}

	if c.LinkID != nil && *c.LinkID != "" {
		if cache != nil {
			if link, ok := cache[*c.LinkID]; ok {
				c.Link = link
			}
		} else {
			c.Link = c.backend.NewCollection()
			if err := c.backend.Storage().Get("id", *c.LinkID, c.Link, true); err != nil {
				return err
			}
		}
	}

	if c.ParentID != nil && *c.ParentID != "" {
		if cache != nil {
			if parent, ok := cache[*c.ParentID]; ok {
				c.Parent = parent
			}
		} else {
			c.Parent = c.backend.NewCollection()
			if err := c.backend.Storage().Get("id", *c.ParentID, c.Parent, true); err != nil {
				return err
			}
		}
	}

	// Resolve associated graphs if any
	ids := []string{}
	for _, entry := range c.Entries {
		if !sliceutil.Has(ids, entry.GraphID) {
			ids = append(ids, entry.GraphID)
		}
	}

	if len(ids) > 0 {
		graphs := []*Graph{}
		if err := c.backend.Storage().Get("id", ids, &graphs, false); err != nil {
			return err
		}

		graphsMap := map[string]*Graph{}
		for _, g := range graphs {
			graphsMap[g.ID] = g
		}

		for _, entry := range c.Entries {
			if g, ok := graphsMap[entry.GraphID]; ok {
				entry.Graph = g.Clone()
			}
		}
	}

	c.resolved = true

	return nil
}

func (c *Collection) treeEntry() *CollectionTreeEntry {
	entry := &CollectionTreeEntry{
		ID:       c.ID,
		Children: &CollectionTree{},
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

// CollectionEntry represents a back-end collection entry instance.
type CollectionEntry struct {
	Index        int         `gorm:"type:int NOT NULL;primary_key" json:"-"`
	Collection   *Collection `json:"-"`
	CollectionID string      `gorm:"column:collection;type:varchar(36) NOT NULL REFERENCES collections (id) ON DELETE CASCADE ON UPDATE CASCADE;primary_key" json:"-"`
	Graph        *Graph      `json:"-"`
	GraphID      string      `gorm:"column:graph;type:varchar(36) NOT NULL REFERENCES graphs (id) ON DELETE CASCADE ON UPDATE CASCADE;primary_key" json:"graph"`
	Attributes   maputil.Map `gorm:"type:text" json:"attributes,omitempty"`
	Options      maputil.Map `gorm:"type:text" json:"options,omitempty"`
}

// CollectionTree represents a back-end collection tree instance.
type CollectionTree []*CollectionTreeEntry

// NewCollectionTree creates a new back-end collection tree instance.
func (b *Backend) NewCollectionTree(root string) (*CollectionTree, error) {
	collections := []*Collection{}
	if _, err := b.Storage().List(&collections, nil, nil, 0, 0, false); err != nil {
		return nil, err
	}

	collectionsCache := map[string]*Collection{}
	for _, c := range collections {
		collectionsCache[c.ID] = c
	}

	entries := map[string]*CollectionTreeEntry{}
	for _, c := range collections {
		if c.Template {
			continue
		}

		c.backend = b
		c.Resolve(collectionsCache)
		c.Expand(nil)

		// Fill collections tree
		if _, ok := entries[c.ID]; !ok {
			entries[c.ID] = c.treeEntry()
		}

		if c.HasParent() {
			parentID := *c.ParentID

			if _, ok := entries[parentID]; !ok {
				c.Parent.backend = b
				c.Parent.Resolve(collectionsCache)
				c.Parent.Expand(nil)

				entries[parentID] = c.Parent.treeEntry()
			}

			*entries[parentID].Children = append(*entries[parentID].Children, entries[c.ID])
		}
	}

	tree := &CollectionTree{}
	for _, entry := range entries {
		if entry.Parent == root {
			*tree = append(*tree, entry)
			sort.Sort(entry.Children)
		}
	}

	sort.Sort(tree)

	return tree, nil
}

func (c CollectionTree) Len() int {
	return len(c)
}

func (c CollectionTree) Less(i, j int) bool {
	return natsort.Compare(c[i].Label, c[j].Label)
}

func (c CollectionTree) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// CollectionTreeEntry represents a back-end collections tree entry instance.
type CollectionTreeEntry struct {
	ID       string          `json:"id,omitempty"`
	Label    string          `json:"label,omitempty"`
	Parent   string          `json:"parent,omitempty"`
	Children *CollectionTree `json:"children,omitempty"`
}
