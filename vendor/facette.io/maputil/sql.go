package maputil

import (
	"database/sql/driver"
	"encoding/json"
)

// Value marshals the keys mapping for compatibility with SQL drivers.
func (m Map) Value() (driver.Value, error) {
	data, err := json.Marshal(m)
	return data, err
}

// Scan unmarshals the keys mapping retrieved from SQL drivers.
func (m *Map) Scan(v interface{}) error {
	switch v.(type) {
	case []byte:
		return json.Unmarshal(v.([]byte), m)

	case string:
		return json.Unmarshal([]byte(v.(string)), m)
	}

	return ErrUnscannableValue
}
