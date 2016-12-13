package backend

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	// GroupPrefix is the source or metric group prefix.
	GroupPrefix = "group:"
)

// SourceGroup represents a library source group item instance.
type SourceGroup struct {
	Item
	Patterns GroupPatternList `orm:"not_null" json:"patterns"`
}

// TableName returns the table name to use in the database.
func (SourceGroup) TableName() string {
	return "sourcegroups"
}

// Validate checks whether or not the source group item instance is valid.
func (g SourceGroup) Validate(backend *Backend) error {
	if err := g.Item.Validate(backend); err != nil {
		return err
	} else if len(g.Patterns) == 0 {
		return ErrEmptyGroup
	}

	return nil
}

// MetricGroup represents a library metric group item instance.
type MetricGroup struct {
	Item
	Patterns GroupPatternList `orm:"not_null" json:"patterns"`
}

// TableName returns the table name to use in the database.
func (MetricGroup) TableName() string {
	return "metricgroups"
}

// Validate checks whether or not the metric group item instance is valid.
func (g MetricGroup) Validate(backend *Backend) error {
	if err := g.Item.Validate(backend); err != nil {
		return err
	} else if len(g.Patterns) == 0 {
		return ErrEmptyGroup
	}

	return nil
}

// GroupPatternList represents a list of group pattern entries.
type GroupPatternList []string

// Value marshals the group pattern entries for compatibility with SQL drivers.
func (l GroupPatternList) Value() (driver.Value, error) {
	data, err := json.Marshal(l)
	return data, err
}

// Scan unmarshals the group pattern entries retrieved from SQL drivers.
func (l *GroupPatternList) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), l)
}
