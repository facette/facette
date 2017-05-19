package backend

import (
	"testing"

	"github.com/facette/maputil"
)

func testProviderNew() []*Provider {
	return []*Provider{
		&Provider{
			Item: Item{
				Name: "item1",
			},
			Connector: "connector1",
			Settings: maputil.Map{
				"key1": "abc",
				"key2": 123.456,
			},
			Filters: ProviderFilters{
				ProviderFilter{Action: "action1", Target: "target1", Pattern: "pattern1", Into: "into1"},
				ProviderFilter{Action: "action1", Target: "target1", Pattern: "pattern2"},
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
			Settings: maputil.Map{
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
				ProviderFilter{Action: "action1", Target: "target1", Pattern: "pattern1", Into: "into1"},
			},
			RefreshInterval: 10,
			Priority:        0,
			Enabled:         false,
		},
	}
}

func testProviderCreate(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemCreate(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderCreateInvalid(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemCreateInvalid(b, &Provider{}, testInterfaceToSlice(testProviders), t)

	if err := b.Storage().Save(&Provider{Item: Item{Name: "name"}, RefreshInterval: -1}); err != ErrInvalidInterval {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidInterval, err)
		t.Fail()
	}

	if err := b.Storage().Save(&Provider{Item: Item{Name: "name"}, Priority: -1}); err != ErrInvalidPriority {
		t.Logf("\nExpected %#v\nbut got  %#v", ErrInvalidPriority, err)
		t.Fail()
	}
}

func testProviderGet(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemGet(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderGetUnknown(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemGetUnknown(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderUpdate(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemUpdate(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderCount(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemCount(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderList(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemList(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderDelete(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemDelete(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}

func testProviderDeleteAll(b *Backend, testProviders []*Provider, t *testing.T) {
	testItemDeleteAll(b, &Provider{}, testInterfaceToSlice(testProviders), t)
}
