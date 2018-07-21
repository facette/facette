package backend

import "testing"

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

func testSourceGroupCreate(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemCreate(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupCreateInvalid(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemCreateInvalid(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)

	if err := b.Storage().Save(&SourceGroup{
		Item:     Item{Name: "name"},
		Patterns: GroupPatterns{"regexp:(.*"}},
	); err != ErrInvalidPattern {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidPattern, err)
		t.Fail()
	}
}

func testSourceGroupGet(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemGet(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupGetUnknown(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemGetUnknown(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupUpdate(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemUpdate(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupCount(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemCount(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupList(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemList(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupDelete(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemDelete(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testSourceGroupDeleteAll(b *Backend, testSourceGroups []*SourceGroup, t *testing.T) {
	testItemDeleteAll(b, &SourceGroup{}, testInterfaceToSlice(testSourceGroups), t)
}

func testMetricGroupCreate(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemCreate(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupCreateInvalid(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemCreateInvalid(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)

	if err := b.Storage().Save(&MetricGroup{
		Item:     Item{Name: "name"},
		Patterns: GroupPatterns{"regexp:(.*"}},
	); err != ErrInvalidPattern {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidPattern, err)
		t.Fail()
	}
}

func testMetricGroupGet(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemGet(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupGetUnknown(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemGetUnknown(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupUpdate(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemUpdate(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupCount(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemCount(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupList(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemList(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupDelete(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemDelete(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}

func testMetricGroupDeleteAll(b *Backend, testMetricGroups []*MetricGroup, t *testing.T) {
	testItemDeleteAll(b, &MetricGroup{}, testInterfaceToSlice(testMetricGroups), t)
}
