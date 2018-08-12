package storage

import (
	"testing"

	"facette.io/maputil"
	"github.com/stretchr/testify/assert"
)

func testProviderNew() []*Provider {
	return []*Provider{
		&Provider{
			Item: Item{
				Name: "item1",
			},
			Connector: "connector1",
			Settings: &maputil.Map{
				"key1": "abc",
				"key2": 123.456,
			},
			Filters: ProviderFilters{
				&ProviderFilter{Action: "action1", Target: "target1", Pattern: "pattern1", Into: "into1"},
				&ProviderFilter{Action: "action1", Target: "target1", Pattern: "pattern2"},
			},
			RefreshInterval: 30,
			Priority:        10,
			Enabled:         true,
		},

		&Provider{
			Item: Item{
				Name: "item2",
			},
			Connector: "connector2",
			Settings: &maputil.Map{
				"key1": "def",
			},
			RefreshInterval: 0,
			Priority:        0,
			Enabled:         true,
		},

		&Provider{
			Item: Item{
				Name: "item3",
			},
			Connector: "connector3",
			Filters: ProviderFilters{
				&ProviderFilter{Action: "action1", Target: "target1", Pattern: "pattern1", Into: "into1"},
			},
			RefreshInterval: 10,
			Priority:        0,
			Enabled:         false,
		},
	}
}

func testProviderCreate(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemCreate(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderCreateInvalid(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemCreateInvalid(s, &Provider{}, testInterfaceToSlice(testProviders), t)
	assert.Equal(t, ErrInvalidInterval, s.SQL().Save(&Provider{Item: Item{Name: "name"}, RefreshInterval: -1}))
	assert.Equal(t, ErrInvalidPriority, s.SQL().Save(&Provider{Item: Item{Name: "name"}, Priority: -1}))
}

func testProviderGet(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemGet(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderGetUnknown(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemGetUnknown(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderUpdate(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemUpdate(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderCount(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemCount(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderList(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemList(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderDelete(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemDelete(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderDeleteAll(s *Storage, testProviders []*Provider, t *testing.T) {
	testItemDeleteAll(s, &Provider{}, testInterfaceToSlice(testProviders), t)
}
