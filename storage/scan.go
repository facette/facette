package storage

import "encoding/json"

func scanValue(v, dst interface{}) error {
	switch v.(type) {
	case []byte:
		return json.Unmarshal(v.([]byte), dst)

	case string:
		return json.Unmarshal([]byte(v.(string)), dst)
	}

	return ErrUnscannableValue
}
