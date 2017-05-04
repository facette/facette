package backend

import (
	"facette/template"
	"sort"

	"github.com/facette/maputil"
	"github.com/facette/natsort"
	"github.com/facette/sliceutil"
	"github.com/jinzhu/gorm"
)

// Collection represents a library collection item instance.
type Collection struct {
	Item
	Entries    []*CollectionEntry `json:"entries,omitempty"`
	Link       *Collection        `json:"-"`
	LinkID     *string            `gorm:"column:link;type:varchar(36);default:NULL" json:"link,omitempty"`
	Attributes maputil.Map        `gorm:"type:text" json:"attributes,omitempty"`
	Alias      *string            `gorm:"type:varchar(128);unique_index" json:"alias,omitempty"`
	Options    maputil.Map        `gorm:"type:text" json:"options,omitempty"`
	Parent     *Collection        `json:"-"`
	ParentID   *string            `gorm:"column:parent;type:varchar(36);default:NULL" json:"parent,omitempty"`
	Template   bool               `gorm:"not null" json:"template"`

	resolved bool
	expanded bool
}

func (b *Backend) NewCollection() *Collection {
	return &Collection{Item: Item{backend: b}}
}

func (c *Collection) BeforeSave(scope *gorm.Scope) error {
	if err := c.Item.BeforeSave(scope); err != nil {
		return err
	} else if c.Alias != nil && !nameRegexp.MatchString(*c.Alias) {
		return ErrInvalidAlias
	}

	for idx, entry := range c.Entries {
		entry.Index = idx + 1
	}

	// Ensure optional fields are null if empty
	if c.LinkID != nil && *c.LinkID == "" {
		c.LinkID = nil
	}

	if c.Alias != nil && *c.Alias == "" {
		c.Alias = nil
	}

	if c.ParentID != nil && *c.ParentID == "" {
		c.ParentID = nil
	}

	return nil
}

func (c *Collection) Expand(attrs maputil.Map) error {
	var err error

	if c.expanded {
		return nil
	}

	if len(attrs) > 0 {
		c.Attributes.Merge(attrs, true)
	}

	if c.backend != nil && c.LinkID != nil && *c.LinkID != "" {
		if err := c.Resolve(); err != nil {
			return err
		}

		// Expand template and applies current collection's attributes and options
		tmpl := c.Link
		tmpl.ID = c.ID
		tmpl.Attributes.Merge(c.Attributes, true)
		tmpl.Options.Merge(c.Options, true)
		tmpl.Template = false

		if title, ok := tmpl.Options["title"].(string); ok {
			if tmpl.Options["title"], err = template.Expand(title, tmpl.Attributes); err != nil {
				return err
			}
		}

		// Fetch graph entries titles
		graphs := map[string]*Graph{}

		values := []string{}
		for _, entry := range tmpl.Entries {
			if !sliceutil.Has(values, entry.GraphID) {
				values = append(values, entry.GraphID)
			}
		}

		if len(values) > 0 {
			result := []*Graph{}
			if err := c.backend.Storage().Get("id", values, &result); err != nil {
				return err
			}

			for _, entry := range result {
				graphs[entry.ID] = entry
			}
		}

		for _, entry := range tmpl.Entries {
			g, ok := graphs[entry.GraphID]
			if !ok {
				continue
			}

			attrs := maputil.Map{}
			attrs.Merge(c.Attributes, true)
			attrs.Merge(g.Attributes, true)

			opts := maputil.Map{}
			opts.Merge(g.Options, true)
			opts.Merge(entry.Options, true)

			if opts.Has("title") {
				title, _ := opts.GetString("title", "")

				if opts["title"], err = template.Expand(title, attrs); err != nil {
					return err
				}
			}

			entry.Options = opts
		}

		*c = *tmpl
	}

	c.expanded = true

	return nil
}

func (c *Collection) HasParent() bool {
	return c.ParentID != nil && *c.ParentID != ""
}

func (c *Collection) Resolve() error {
	if c.resolved {
		return nil
	} else if c.backend == nil {
		return ErrUnresolvableItem
	}

	if c.LinkID != nil && *c.LinkID != "" {
		c.Link = c.backend.NewCollection()
		if err := c.backend.Storage().Get("id", *c.LinkID, c.Link); err != nil {
			return err
		}
	}

	if c.ParentID != nil && *c.ParentID != "" {
		c.Parent = c.backend.NewCollection()
		if err := c.backend.Storage().Get("id", *c.ParentID, c.Parent); err != nil {
			return err
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

// CollectionEntry represents a library collection entry instance.
type CollectionEntry struct {
	Index        int         `gorm:"type:int;not null;primary_key" json:"-"`
	Collection   *Collection `json:"-"`
	CollectionID string      `gorm:"column:collection;type:varchar(36);not null;primary_key" json:"-"`
	Graph        *Graph      `json:"-"`
	GraphID      string      `gorm:"column:graph;type:varchar(36);not null;primary_key" json:"graph"`
	Attributes   maputil.Map `gorm:"type:text" json:"attributes,omitempty"`
	Options      maputil.Map `gorm:"type:text" json:"options,omitempty"`
}

type CollectionTree []*CollectionTreeEntry

func (b *Backend) NewCollectionTree(root string) (*CollectionTree, error) {
	filters := map[string]interface{}{"template": false}
	if root != "" {
		filters["parent"] = root
	}

	collections := []*Collection{}
	if _, err := b.Storage().List(&collections, filters, nil, 0, 0); err != nil {
		return nil, err
	}

	entries := map[string]*CollectionTreeEntry{}
	for _, c := range collections {
		c.backend = b
		c.Resolve()
		c.Expand(nil)

		// Fill collections tree
		if _, ok := entries[c.ID]; !ok {
			entries[c.ID] = c.treeEntry()
		}

		if c.HasParent() {
			parentID := *c.ParentID

			if _, ok := entries[parentID]; !ok {
				c.Parent.Resolve()
				c.Parent.Expand(nil)

				entries[parentID] = c.Parent.treeEntry()
			}

			*entries[parentID].Children = append(*entries[parentID].Children, entries[c.ID])

			if parentID == filters["parent"] {
				entries[c.ID].Parent = ""
			}
		}
	}

	tree := &CollectionTree{}
	for _, entry := range entries {
		if entry.Parent == "" && entry.ID != filters["parent"] {
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

type CollectionTreeEntry struct {
	ID       string          `json:"id,omitempty"`
	Label    string          `json:"label,omitempty"`
	Parent   string          `json:"parent,omitempty"`
	Children *CollectionTree `json:"children,omitempty"`
}
