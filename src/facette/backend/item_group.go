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
	Patterns *GroupPatterns `gorm:"type:text;not null" json:"patterns"`
}

// NewSourceGroup creates a new back-end source group item instance.
func (b *Backend) NewSourceGroup() *SourceGroup {
	return &SourceGroup{Item: Item{backend: b}}
}

// TableName returns the table name to use in the database.
func (SourceGroup) TableName() string {
	return "sourcegroups"
}

// MetricGroup represents a library metric group item instance.
type MetricGroup struct {
	Item
	Patterns *GroupPatterns `gorm:"type:text;not null" json:"patterns"`
}

// NewMetricGroup creates a new back-end metric group item instance.
func (b *Backend) NewMetricGroup() *MetricGroup {
	return &MetricGroup{Item: Item{backend: b}}
}

// TableName returns the table name to use in the database.
func (MetricGroup) TableName() string {
	return "metricgroups"
}

// GroupPatterns represents a list of group patterns.
type GroupPatterns []string

// Value marshals the group patterns for compatibility with SQL drivers.
func (l GroupPatterns) Value() (driver.Value, error) {
	data, err := json.Marshal(l)
	return data, err
}

// Scan unmarshals the group patterns retrieved from SQL drivers.
func (l *GroupPatterns) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), l)
}
