package backend

import (
	"facette/template"

	"github.com/facette/maputil"
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
}

func (b *Backend) NewCollection() *Collection {
	return &Collection{Item: Item{Backend: b}}
}

func (c *Collection) BeforeSave(scope *gorm.Scope) error {
	if err := c.Item.BeforeSave(scope); err != nil {
		return err
	} else if c.Alias != nil && !nameRegexp.MatchString(*c.Alias) {
		return ErrInvalidAlias
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

	if len(attrs) > 0 {
		c.Attributes.Merge(attrs, true)
	}

	if c.Backend != nil && c.LinkID != nil && *c.LinkID != "" {
		// Expand template and applies current collection's attributes and options
		tmpl := c.Backend.NewCollection()
		if err := c.Backend.Storage().Get("id", *c.LinkID, tmpl); err != nil {
			return err
		}

		tmpl.Attributes.Merge(c.Attributes, true)
		tmpl.Options.Merge(c.Options, true)

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
			if err := c.Backend.Storage().Get("id", values, &result); err != nil {
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

	return nil
}

func (c *Collection) HasParent() bool {
	return c.ParentID != nil && *c.ParentID != ""
}

// CollectionEntry represents a library collection entry instance.
type CollectionEntry struct {
	Collection   *Collection `json:"-"`
	CollectionID string      `gorm:"column:collection;type:varchar(36);not null" json:"-"`
	Graph        *Graph      `json:"-"`
	GraphID      string      `gorm:"column:graph;type:varchar(36);not null" json:"graph"`
	Attributes   maputil.Map `gorm:"type:text" json:"attributes,omitempty"`
	Options      maputil.Map `gorm:"type:text" json:"options,omitempty"`
}
