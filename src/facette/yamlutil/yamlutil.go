package yamlutil

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// UnmarshalFile parses the YAML-encoded data read from the file named by filename and stores the result in the
// value pointed to by v.
func UnmarshalFile(filename string, v interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, v)
}
