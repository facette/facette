package server

import (
	"encoding/json"
	"math"
)

type plotValue float64

func (value plotValue) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(value)) {
		return json.Marshal(nil)
	}

	return json.Marshal(float64(value))
}

type plotList []plotValue

func (list plotList) Len() int {
	return len(list)
}

func (list plotList) Less(i, j int) bool {
	return list[i] < list[j]
}

func (list plotList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
