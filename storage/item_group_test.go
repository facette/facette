package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSourceGroupNew() []*SourceGroup {
	return []*SourceGroup{
		&SourceGroup{
			Item: Item{
				Name: "item1",
			},
			Patterns: GroupPatterns{
				"pattern1",
				"pattern2",
			},
		},

		&SourceGroup{
			Item: Item{
				Name: "item2",
			},
			Patterns: GroupPatterns{
				"pattern2",
			},
		},

		&SourceGroup{
			Item: Item{
				Name: "item3",
			},
			Patterns: GroupPatterns{
				"pattern1",
			},
		},
	}
}

func testMetricGroupNew() []*MetricGroup {
	return []*MetricGroup{
		&MetricGroup{
			Item: Item{
				Name: "item1",
			},
			Patterns: GroupPatterns{
				"pattern1",
				"pattern2",
			},
		},

		&MetricGroup{
			Item: Item{
				Name: "item2",
			},
			Patterns: GroupPatterns{
				"pattern2",
			},
		},

		&MetricGroup{
			Item: Item{
				Name: "item3",
			},
			Patterns: GroupPatterns{
				"pattern1",
			},
		},
	}
}

func testSourceGroupCreate(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemCreate(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupCreateInvalid(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemCreateInvalid(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
	assert.Equal(t, ErrInvalidPattern, s.SQL().Save(&SourceGroup{
		Item:     Item{Name: "name"},
		Patterns: GroupPatterns{"regexp:(.*"},
	}))
}

func testSourceGroupGet(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemGet(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupGetUnknown(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemGetUnknown(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupUpdate(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemUpdate(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupCount(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemCount(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupList(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemList(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupDelete(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemDelete(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupDeleteAll(s *Storage, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemDeleteAll(s, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testMetricGroupCreate(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemCreate(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupCreateInvalid(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemCreateInvalid(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
	assert.Equal(t, ErrInvalidPattern, s.SQL().Save(&MetricGroup{
		Item:     Item{Name: "name"},
		Patterns: GroupPatterns{"regexp:(.*"},
	}))
}

func testMetricGroupGet(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemGet(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupGetUnknown(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemGetUnknown(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupUpdate(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemUpdate(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupCount(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemCount(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupList(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemList(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupDelete(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemDelete(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupDeleteAll(s *Storage, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemDeleteAll(s, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}
