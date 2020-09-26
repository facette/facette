// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package catalog

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"facette.io/facette/pkg/labels"
)

var testCatalog = New()

func Test_Catalog_Link(t *testing.T) {
	for _, metric := range testMetrics {
		err := testSection.Insert(metric)
		assert.Nil(t, err)
	}

	testCatalog.Link("test", testSection)
	assert.Len(t, testCatalog.sections, 1)

	_, ok := testCatalog.sections["test"]
	assert.True(t, ok)
}

func Test_Catalog_Labels(t *testing.T) {
	for _, test := range []struct {
		matcher  labels.Matcher
		expected []string
	}{
		{
			matcher:  nil,
			expected: []string{"__name__", "abc", "def"},
		},
		{
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpEq, labels.Name, "foo")},
			expected: []string{"__name__", "abc"},
		},
		{
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpNotEq, "def", "")},
			expected: []string{"__name__", "abc", "def"},
		},
	} {
		assert.Equal(t, test.expected, testCatalog.Labels(&ListOptions{Matcher: test.matcher}))
	}
}

func Test_Catalog_Metrics(t *testing.T) {
	metrics := append([]Metric(nil), testMetrics...)

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].String() < metrics[j].String()
	})

	for _, test := range []struct {
		matcher  labels.Matcher
		expected []Metric
	}{
		{
			matcher:  nil,
			expected: metrics,
		},
		{
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpNotEq, "def", "")},
			expected: metrics[0:1],
		},
		{
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpEq, labels.Name, "foo")},
			expected: metrics[1:3],
		},
		{
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpEq, labels.Name, "bar")},
			expected: metrics[0:1],
		},
		{
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpNotEq, "def", "")},
			expected: metrics[0:1],
		},
	} {
		assert.Equal(t, test.expected, testCatalog.Metrics(&ListOptions{Matcher: test.matcher}))
	}
}

func Test_Catalog_Values(t *testing.T) {
	for _, test := range []struct {
		label    string
		matcher  labels.Matcher
		expected []string
	}{
		{
			label:    labels.Name,
			matcher:  nil,
			expected: []string{"bar", "foo"},
		},
		{
			label:    labels.Name,
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpNotEq, "def", "")},
			expected: []string{"bar"},
		},
		{
			label:    "abc",
			matcher:  nil,
			expected: []string{"123", "456"},
		},
		{
			label:    "abc",
			matcher:  labels.Matcher{mustMatchCond(t, labels.OpEq, labels.Name, "bar")},
			expected: []string{"123"},
		},
	} {
		assert.Equal(t, test.expected, testCatalog.Values(test.label, &ListOptions{Matcher: test.matcher}))
	}
}

func Test_Catalog_Unlink(t *testing.T) {
	testCatalog.Unlink("test")
	assert.Len(t, testCatalog.sections, 0)

	_, ok := testCatalog.sections["test"]
	assert.False(t, ok)
}

func mustMatchCond(t *testing.T, op labels.Op, name, value string) labels.MatchCond {
	cond, err := labels.NewMatchCond(op, name, value)
	assert.Nil(t, err)

	return cond
}
