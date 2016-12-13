package backend

import (
	"database/sql/driver"
	"encoding/json"

	"facette/mapper"
)

// Provider represents a backend provider item instance.
type Provider struct {
	Item
	Connector       string             `orm:"type:varchar(32);not_null" json:"connector"`
	Settings        mapper.Map         `json:"settings"`
	Filters         ProviderFilterList `json:"filters"`
	RefreshInterval int                `orm:"not_null;default:0" json:"refresh_interval"`
	Priority        int                `orm:"not_null;default:0" json:"priority"`
	Enabled         bool               `orm:"not_null;default:true" json:"enabled"`
}

// ProviderFilterList represents a list of backend provider filters.
type ProviderFilterList []ProviderFilter

// Value marshals the provider filter entries for compatibility with SQL drivers.
func (l ProviderFilterList) Value() (driver.Value, error) {
	data, err := json.Marshal(l)
	return data, err
}

// Scan unmarshals the provider filter entries retrieved from SQL drivers.
func (l *ProviderFilterList) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), l)
}

// ProviderFilter represents a backend provider filter entry instance.
type ProviderFilter struct {
	Action  string `json:"action"`
	Target  string `json:"target"`
	Pattern string `json:"pattern"`
	Into    string `json:"into"`
}
