// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package api

// LabelValues are an API label values.
type LabelValues struct {
	Values []string `json:"values"`
	Total  int64    `json:"total"`
}
