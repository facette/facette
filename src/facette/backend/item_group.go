package backend

import (
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"strings"

	"facette/pattern"

	"github.com/jinzhu/gorm"
)

const (
	// GroupPrefix represents the source or metric group prefix.
	GroupPrefix = "group:"
)

// SourceGroup represents a library source group item instance.
type SourceGroup struct {
	Item
	Patterns GroupPatterns `gorm:"type:text;not null" json:"patterns"`
}

// NewSourceGroup creates a new back-end source group item instance.
func (b *Backend) NewSourceGroup() *SourceGroup {
	return &SourceGroup{Item: Item{backend: b}}
}

// BeforeSave handles the ORM 'BeforeSave' callback.
func (sg *SourceGroup) BeforeSave(scope *gorm.Scope) error {
	if err := sg.Item.BeforeSave(scope); err != nil {
		return err
	}

	if len(sg.Patterns) == 0 {
		return ErrEmptyGroup
	}

	for _, p := range sg.Patterns {
		if !strings.HasPrefix(p, pattern.RegexpPrefix) {
			continue
		}

		if _, err := regexp.Compile(strings.TrimPrefix(p, pattern.RegexpPrefix)); err != nil {
			return ErrInvalidPattern
		}
	}

	return nil
}

// TableName returns the table name to use in the database.
func (SourceGroup) TableName() string {
	return "sourcegroups"
}

// MetricGroup represents a library metric group item instance.
type MetricGroup struct {
	Item
	Patterns GroupPatterns `gorm:"type:text;not null" json:"patterns"`
}

// NewMetricGroup creates a new back-end metric group item instance.
func (b *Backend) NewMetricGroup() *MetricGroup {
	return &MetricGroup{Item: Item{backend: b}}
}

// BeforeSave handles the ORM 'BeforeSave' callback.
func (mg *MetricGroup) BeforeSave(scope *gorm.Scope) error {
	if err := mg.Item.BeforeSave(scope); err != nil {
		return err
	}

	for _, p := range mg.Patterns {
		if !strings.HasPrefix(p, pattern.RegexpPrefix) {
			continue
		}

		if _, err := regexp.Compile(strings.TrimPrefix(p, pattern.RegexpPrefix)); err != nil {
			return ErrInvalidPattern
		}
	}

	return nil
}

// TableName returns the table name to use in the database.
func (MetricGroup) TableName() string {
	return "metricgroups"
}

// GroupPatterns represents a list of group patterns.
type GroupPatterns []string

// Value marshals the group patterns for compatibility with SQL drivers.
func (gp GroupPatterns) Value() (driver.Value, error) {
	data, err := json.Marshal(gp)
	return data, err
}

// Scan unmarshals the group patterns retrieved from SQL drivers.
func (gp *GroupPatterns) Scan(v interface{}) error {
	return scanValue(v, gp)
}
