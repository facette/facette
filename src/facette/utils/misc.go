package utils

import (
	"bytes"
	"encoding/gob"
)

// Clone performs a deep copy of an interface.
func Clone(src, dst interface{}) {
	var (
		buffer  *bytes.Buffer
		decoder *gob.Decoder
		encoder *gob.Encoder
	)

	buffer = new(bytes.Buffer)

	encoder = gob.NewEncoder(buffer)
	decoder = gob.NewDecoder(buffer)

	encoder.Encode(src)
	decoder.Decode(dst)
}
