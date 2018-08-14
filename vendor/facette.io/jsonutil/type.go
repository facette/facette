package jsonutil

import (
	"bytes"
	"encoding/json"
)

// NullString represents a nullable string type.
type NullString string

// MarshalJSON satisfies the json.Marshaller interface.
func (s NullString) MarshalJSON() ([]byte, error) {
	if s == "" {
		return []byte("null"), nil
	}
	return json.Marshal(string(s))
}

// UnmarshalJSON satisfies the json.Unmarshaller interface.
func (s NullString) UnmarshalJSON(b []byte) error {
	if !bytes.Equal(b, []byte("null")) {
		return json.Unmarshal(b, &s)
	}
	return nil
}
