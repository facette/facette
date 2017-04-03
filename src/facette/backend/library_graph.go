package backend

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"facette/mapper"
	"facette/template"
)

// Graph represents a library graph item instance.
type Graph struct {
	Item
	Groups     SeriesGroupList `json:"groups,omitempty"`
	Link       *Graph          `orm:"type:varchar(36);foreign_key:ID" json:"-"`
	LinkID     string          `orm:"-" json:"link,omitempty"`
	Attributes mapper.Map      `json:"attributes,omitempty"`
	Alias      string          `orm:"type:varchar(128);unique" json:"alias,omitempty"`
	Options    mapper.Map      `json:"options,omitempty"`
	Template   bool            `orm:"not_null;default:false" json:"template"`
}

// Validate checks whether or not the graph item instance is valid.
func (g Graph) Validate(backend *Backend) error {
	if err := g.Item.Validate(backend); err != nil {
		return err
	} else if !g.Template && g.LinkID == "" && len(g.Groups) == 0 && len(g.Attributes) == 0 {
		return ErrEmptyGraph
	} else if g.Template && (len(g.Attributes) > 0 || g.Alias != "") {
		return ErrIncompatibleAttributes
	} else if g.Alias != "" && !authorizedAliasChars.MatchString(g.Alias) {
		return ErrInvalidAlias
	}

	return nil
}

// Expand expands the graph template with attributes passed as parameter.
func (g *Graph) Expand(attrs mapper.Map) error {
	var err error

	g.Template = false

	if title, ok := g.Options["title"].(string); ok {
		if g.Options["title"], err = template.Expand(title, attrs); err != nil {
			return err
		}
	}

	for i := range g.Groups {
		for j := range g.Groups[i].Series {
			series := &g.Groups[i].Series[j]
			if series.Name, err = template.Expand(series.Name, attrs); err != nil {
				return err
			} else if series.Origin, err = template.Expand(series.Origin, attrs); err != nil {
				return err
			} else if series.Source, err = template.Expand(series.Source, attrs); err != nil {
				return err
			} else if series.Metric, err = template.Expand(series.Metric, attrs); err != nil {
				return err
			}
		}
	}

	return nil
}

// SeriesGroupList represents a list of library graph series group entries.
type SeriesGroupList []SeriesGroup

// Value marshals the series group entries for compatibility with SQL drivers.
func (l SeriesGroupList) Value() (driver.Value, error) {
	data, err := json.Marshal(l)
	return data, err
}

// Scan unmarshals the series group entries retrieved from SQL drivers.
func (l *SeriesGroupList) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), l)
}

// SeriesGroup represents a library graph series group entry instance.
type SeriesGroup struct {
	Name        string     `json:"name"`
	Operator    int        `json:"operator"`
	Consolidate int        `json:"consolidate"`
	Series      []Series   `json:"series"`
	Options     mapper.Map `json:"options,omitempty"`
}

// Series represents a library graph series entry instance.
type Series struct {
	Name    string     `json:"name"`
	Origin  string     `json:"origin"`
	Source  string     `json:"source"`
	Metric  string     `json:"metric"`
	Options mapper.Map `json:"options,omitempty"`
}

// IsValid checks whether or not the series instance is valid.
func (s Series) IsValid() bool {
	return s.Origin != "" && s.Source != "" && s.Metric != ""
}

// String returns a string representation of the series instance.
func (s Series) String() string {
	return fmt.Sprintf("{Name: %q, Origin: %q, Source: %q, Metric: %q}", s.Name, s.Origin, s.Source, s.Metric)
}
