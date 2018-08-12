package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"facette.io/facette/template"

	"facette.io/maputil"
	"github.com/jinzhu/gorm"
)

// Graph represents a library graph item instance.
type Graph struct {
	Item
	Groups     SeriesGroups `gorm:"type:text;not null" json:"groups,omitempty"`
	Link       *Graph       `json:"-"`
	LinkID     *string      `gorm:"column:link;type:varchar(36) DEFAULT NULL REFERENCES graphs (id) ON DELETE CASCADE ON UPDATE CASCADE" json:"link,omitempty"`
	Attributes maputil.Map  `gorm:"type:text" json:"attributes,omitempty"`
	Alias      *string      `gorm:"type:varchar(128);unique_index" json:"alias,omitempty"`
	Options    maputil.Map  `gorm:"type:text" json:"options,omitempty"`
	Template   bool         `gorm:"not null" json:"template"`

	resolved bool
	expanded bool
}

// NewGraph creates a new storage graph item instance.
func (s *Storage) NewGraph() *Graph {
	return &Graph{Item: Item{storage: s}}
}

// BeforeSave handles the ORM 'BeforeSave' callback.
func (g *Graph) BeforeSave(scope *gorm.Scope) error {
	if err := g.Item.BeforeSave(scope); err != nil {
		return err
	} else if g.Alias != nil && *g.Alias != "" && !nameRegexp.MatchString(*g.Alias) {
		return ErrInvalidAlias
	}

	// Ensure optional fields are null if empty
	if g.LinkID != nil && *g.LinkID == "" {
		scope.SetColumn("LinkID", nil)
	}

	if g.Alias != nil && *g.Alias == "" {
		scope.SetColumn("Alias", nil)
	}

	return nil
}

// Clone returns a clone of the graph item instance.
func (g *Graph) Clone() *Graph {
	clone := &Graph{}
	*clone = *g

	clone.Groups = make(SeriesGroups, len(g.Groups))
	for i, group := range g.Groups {
		clone.Groups[i] = &SeriesGroup{}
		*clone.Groups[i] = *group

		if group.Options != nil {
			clone.Groups[i].Options = group.Options.Clone()
		}

		clone.Groups[i].Series = make([]*Series, len(group.Series))
		for j, series := range group.Series {
			clone.Groups[i].Series[j] = &Series{}
			*clone.Groups[i].Series[j] = *series

			if series.Options != nil {
				clone.Groups[i].Series[j].Options = series.Options.Clone()
			}
		}
	}

	if g.Attributes != nil {
		clone.Attributes = g.Attributes.Clone()
	}

	if g.Options != nil {
		clone.Options = g.Options.Clone()
	}

	if g.Link != nil {
		clone.Link = &Graph{}
		*clone.Link = *g.Link
	}

	return clone
}

// Expand expands the graph item instance using its linked instance.
func (g *Graph) Expand(attrs maputil.Map) error {
	var err error

	if g.expanded {
		return nil
	}

	if len(attrs) > 0 {
		g.Attributes.Merge(attrs, true)
	}

	if g.storage != nil && g.LinkID != nil && *g.LinkID != "" {
		err = g.Resolve()
		if err != nil {
			return err
		}

		// Expand template and applies current graph's attributes
		if g.Link == nil {
			return ErrUnresolvableItem
		}

		tmpl := g.Link.Clone()
		tmpl.ID = g.ID
		tmpl.Attributes.Merge(g.Attributes, true)
		tmpl.Options.Merge(g.Options, true)
		tmpl.Template = false

		*g = *tmpl
	}

	if title, ok := g.Options["title"].(string); ok {
		if g.Options["title"], err = template.Expand(title, g.Attributes); err != nil {
			return err
		}
	}

	for _, group := range g.Groups {
		for _, series := range group.Series {
			if series.Name, err = template.Expand(series.Name, g.Attributes); err != nil {
				return err
			} else if series.Origin, err = template.Expand(series.Origin, g.Attributes); err != nil {
				return err
			} else if series.Source, err = template.Expand(series.Source, g.Attributes); err != nil {
				return err
			} else if series.Metric, err = template.Expand(series.Metric, g.Attributes); err != nil {
				return err
			}
		}
	}

	g.expanded = true

	return nil
}

// Resolve resolves the graph item linked instance.
func (g *Graph) Resolve() error {
	if g.resolved {
		return nil
	} else if g.storage == nil {
		return ErrUnresolvableItem
	}

	if g.LinkID != nil && *g.LinkID != "" {
		g.Link = g.storage.NewGraph()
		if err := g.storage.SQL().Get("id", *g.LinkID, g.Link, false); err != nil {
			return err
		}
	}

	g.resolved = true

	return nil
}

// SeriesGroups represents a list of library graph series groups.
type SeriesGroups []*SeriesGroup

// Value marshals the series groups for compatibility with SQL drivers.
func (sg SeriesGroups) Value() (driver.Value, error) {
	data, err := json.Marshal(sg)
	return data, err
}

// Scan unmarshals the series groups retrieved from SQL drivers.
func (sg *SeriesGroups) Scan(v interface{}) error {
	return scanValue(v, sg)
}

// SeriesGroup represents a library graph series group entry instance.
type SeriesGroup struct {
	Name        string      `json:"name"`
	Operator    int         `json:"operator"`
	Consolidate int         `json:"consolidate"`
	Series      []*Series   `json:"series"`
	Options     maputil.Map `json:"options,omitempty"`
}

// Series represents a library graph series entry instance.
type Series struct {
	Name    string      `json:"name"`
	Origin  string      `json:"origin"`
	Source  string      `json:"source"`
	Metric  string      `json:"metric"`
	Options maputil.Map `json:"options,omitempty"`
}

// IsValid checks whether or not the series instance is valid.
func (s Series) IsValid() bool {
	return s.Origin != "" && s.Source != "" && s.Metric != ""
}

// String returns a string representation of the series instance.
func (s Series) String() string {
	return fmt.Sprintf("{Name: %q, Origin: %q, Source: %q, Metric: %q}", s.Name, s.Origin, s.Source, s.Metric)
}
