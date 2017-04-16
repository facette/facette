package backend

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"facette/mapper"
)

var (
	dateCreated  time.Time
	mysqlConfig  mapper.Map
	pgsqlConfig  mapper.Map
	sqliteConfig mapper.Map
)

func init() {
	dateCreated = time.Now().UTC().Round(time.Second)

	// MySQL
	mysqlConfig = mapper.Map{"driver": "mysql"}

	if v := os.Getenv("TEST_MYSQL_DBNAME"); v != "" {
		mysqlConfig.Set("dbname", v)
	}
	if v := os.Getenv("TEST_MYSQL_HOST"); v != "" {
		mysqlConfig.Set("host", v)
	}
	if v := os.Getenv("TEST_MYSQL_PORT"); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Fatalf("failed to convert port to integer: %s", err)
		}
		mysqlConfig.Set("port", i)
	}
	if v := os.Getenv("TEST_MYSQL_USER"); v != "" {
		mysqlConfig.Set("user", v)
	}
	if v := os.Getenv("TEST_MYSQL_PASSWORD"); v != "" {
		mysqlConfig.Set("password", v)
	}

	// PostgreSQL
	pgsqlConfig = mapper.Map{
		"driver":  "pgsql",
		"sslmode": "disable",
	}

	if v := os.Getenv("TEST_PGSQL_DBNAME"); v != "" {
		pgsqlConfig.Set("dbname", v)
	}
	if v := os.Getenv("TEST_PGSQL_HOST"); v != "" {
		pgsqlConfig.Set("host", v)
	}
	if v := os.Getenv("TEST_PGSQL_PORT"); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Fatalf("failed to convert port to integer: %s", err)
		}
		pgsqlConfig.Set("port", i)
	}
	if v := os.Getenv("TEST_PGSQL_USER"); v != "" {
		pgsqlConfig.Set("user", v)
	}
	if v := os.Getenv("TEST_PGSQL_PASSWORD"); v != "" {
		pgsqlConfig.Set("password", v)
	}

	// SQLite
	sqliteConfig = mapper.Map{"driver": "sqlite"}

	if v := os.Getenv("TEST_SQLITE_PATH"); v != "" {
		sqliteConfig.Set("path", v)
	} else {
		tmpFile, err := ioutil.TempFile("", "facette")
		if err != nil {
			log.Fatalf("failed to create temporary file: %s", err)
		}
		defer os.Remove(tmpFile.Name())

		sqliteConfig.Set("path", tmpFile.Name())
	}
}

func execTestProvider(config *mapper.Map, t *testing.T) {
	var (
		desc        = "A great provider description"
		descUpdated = "A great provider description (updated)"
	)

	b, err := NewBackend(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	execTest(b, []interface{}{
		&Provider{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000000",
				Name:        "provider1",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Connector: "a",
			Settings: mapper.Map{
				"key1": "value1",
				"key2": 1.23,
			},
		},
		&Provider{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "provider2",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Connector: "b",
			Settings: mapper.Map{
				"key1": "value2",
				"key2": 456,
				"key3": true,
			},
			Filters: []ProviderFilter{
				{Action: "action1", Target: "target1", Pattern: "pattern1", Into: "into1"},
				{Action: "action1", Target: "target1", Pattern: "pattern2"},
			},
			RefreshInterval: 30,
		},
		&Provider{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000002",
				Name:        "provider3-test",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Connector: "a",
			Settings: mapper.Map{
				"key1": "value2",
			},
		},
	}, &Provider{
		Item: Item{
			ID:          "00000000-0000-0000-0000-000000000000",
			Name:        "provider1",
			Description: &descUpdated,
			Created:     dateCreated,
		},
		Connector: "a",
		Settings: mapper.Map{
			"key1": "value1",
			"key2": 12.3,
		},
	}, t)
}

func execTestCollection(config *mapper.Map, t *testing.T) {
	var (
		desc         = "A great collection description"
		descUpdated  = "A great collection description (updated)"
		descTemplate = "A great collection description for {{ .source )}"
	)

	b, err := NewBackend(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	execTest(b, []interface{}{
		&Collection{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000000",
				Name:        "collection1",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Entries: []CollectionEntry{
				{
					ID: "00000000-0000-0000-0000-000000000000",
					Options: map[string]interface{}{
						"title": "A great graph title",
					},
				},
			},
			Options: map[string]interface{}{
				"title": "A great collection title",
			},
			Template: false,
		},
		&Collection{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "collection2",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Entries: []CollectionEntry{
				{
					ID: "00000000-0000-0000-0000-000000000000",
					Options: map[string]interface{}{
						"title": "A great graph title",
					},
				},
			},
			Options: map[string]interface{}{
				"title": "A great collection title",
			},
			Template: false,
		},
		&Collection{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000002",
				Name:        "collection_tmpl1",
				Description: &descTemplate,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Entries: []CollectionEntry{
				{
					ID: "00000000-0000-0000-0000-000000000000",
					Options: map[string]interface{}{
						"title": "A great graph title for {{ .source }}",
					},
				},
			},
			Options: map[string]interface{}{
				"title": "A collection title for {{ .source }}",
			},
			Template: true,
		},
		&Collection{
			Item: Item{
				ID:       "00000000-0000-0000-0000-000000000003",
				Name:     "collection_tmpl1-1",
				Created:  dateCreated,
				Modified: &dateCreated,
			},
			Link: &Collection{
				Item: Item{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "collection2",
					Description: &desc,
					Created:     dateCreated,
					Modified:    &dateCreated,
				},
				Entries: []CollectionEntry{
					{
						ID: "00000000-0000-0000-0000-000000000000",
						Options: map[string]interface{}{
							"title": "A great graph title",
						},
					},
				},
				Options: map[string]interface{}{
					"title": "A great collection title",
				},
				Template: false,
			},
			Attributes: map[string]interface{}{
				"source": "source1",
			},
			Template: false,
		},
	}, &Collection{
		Item: Item{
			ID:          "00000000-0000-0000-0000-000000000000",
			Name:        "collection1",
			Description: &descUpdated,
			Created:     dateCreated,
		},
		Entries: []CollectionEntry{
			{
				ID: "00000000-0000-0000-0000-000000000000",
				Options: map[string]interface{}{
					"title": "A great graph title",
				},
			},
		},
		Options: map[string]interface{}{
			"title": "A great collection title (updated)",
		},
		Template: false,
	}, t)
}

func execTestGraph(config *mapper.Map, t *testing.T) {
	var (
		desc         = "A great graph description"
		descUpdated  = "A great graph description (updated)"
		descTemplate = "A great graph description for {{ .source )}"
	)

	b, err := NewBackend(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	execTest(b, []interface{}{
		&Graph{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000000",
				Name:        "graph1",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Groups: []SeriesGroup{
				{
					Series: []Series{
						{Name: "series1", Origin: "origin1", Source: "source1", Metric: "metric1"},
					},
				},
			},
			Options: map[string]interface{}{
				"title": "A great graph title",
			},
			Template: false,
		},
		&Graph{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "graph2",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Groups: []SeriesGroup{
				{
					Series: []Series{
						{Name: "series1", Origin: "origin1", Source: "source1", Metric: "metric2"},
					},
				},
			},
			Options: map[string]interface{}{
				"title": "A great graph title",
			},
			Template: false,
		},
		&Graph{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000002",
				Name:        "graph_tmpl1",
				Description: &descTemplate,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Groups: []SeriesGroup{
				{
					Series: []Series{
						{Name: "series1", Origin: "origin1", Source: "{{ .source }}", Metric: "metric1"},
					},
				},
			},
			Options: map[string]interface{}{
				"title": "A graph title for {{ .source }}",
			},
			Template: true,
		},
		&Graph{
			Item: Item{
				ID:       "00000000-0000-0000-0000-000000000003",
				Name:     "graph_tmpl1-1",
				Created:  dateCreated,
				Modified: &dateCreated,
			},
			Link: &Graph{
				Item: Item{
					ID:          "00000000-0000-0000-0000-000000000001",
					Name:        "graph2",
					Description: &desc,
					Created:     dateCreated,
					Modified:    &dateCreated,
				},
				Groups: []SeriesGroup{
					{
						Series: []Series{
							{Name: "series1", Origin: "origin1", Source: "source1", Metric: "metric2"},
						},
					},
				},
				Options: map[string]interface{}{
					"title": "A great graph title",
				},
				Template: false,
			},
			Attributes: map[string]interface{}{
				"source": "source1",
			},
			Template: false,
		},
	}, &Graph{
		Item: Item{
			ID:          "00000000-0000-0000-0000-000000000000",
			Name:        "graph1",
			Description: &descUpdated,
			Created:     dateCreated,
		},
		Groups: []SeriesGroup{
			{
				Series: []Series{
					{Name: "series1", Origin: "origin1", Source: "source1", Metric: "metric1"},
					{Name: "series2", Origin: "origin1", Source: "source1", Metric: "metric2"},
				},
			},
		},
		Options: map[string]interface{}{
			"title": "A great graph title (updated)",
		},
		Template: false,
	}, t)
}

func execTestSourceGroup(config *mapper.Map, t *testing.T) {
	var (
		desc        = "A great sourcegroup description"
		descUpdated = "A great sourcegroup description (updated)"
	)

	b, err := NewBackend(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	execTest(b, []interface{}{
		&SourceGroup{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000000",
				Name:        "sourcegroup1",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Patterns: []string{"glob:host*.example.net"},
		},
		&SourceGroup{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "sourcegroup2",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Patterns: []string{"host2.example.net"},
		},
		&SourceGroup{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000002",
				Name:        "sourcegroup3-test",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Patterns: []string{"host3.example.net"},
		},
	}, &SourceGroup{
		Item: Item{
			ID:          "00000000-0000-0000-0000-000000000000",
			Name:        "sourcegroup1",
			Description: &descUpdated,
			Created:     dateCreated,
		},
		Patterns: []string{"glob:host*.example.net"},
	}, t)
}

func execTestMetricGroup(config *mapper.Map, t *testing.T) {
	var (
		desc        = "A great metricgroup description"
		descUpdated = "A great metricgroup description (updated)"
	)

	b, err := NewBackend(config, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	execTest(b, []interface{}{
		&MetricGroup{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000000",
				Name:        "metricgroup1",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Patterns: []string{"glob:metric1.*"},
		},
		&MetricGroup{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "metricgroup2",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Patterns: []string{"metric2"},
		},
		&MetricGroup{
			Item: Item{
				ID:          "00000000-0000-0000-0000-000000000002",
				Name:        "metricgroup3-test",
				Description: &desc,
				Created:     dateCreated,
				Modified:    &dateCreated,
			},
			Patterns: []string{"metric3"},
		},
	}, &MetricGroup{
		Item: Item{
			ID:          "00000000-0000-0000-0000-000000000000",
			Name:        "metricgroup1",
			Description: &descUpdated,
			Created:     dateCreated,
		},
		Patterns: []string{"glob:metric1.*"},
	}, t)
}

func execTest(b *Backend, items []interface{}, update interface{}, t *testing.T) {
	// Check items insertion
	for _, item := range items {
		rv := reflect.ValueOf(item)

		if err := b.Add(item); err != nil {
			t.Fatal(err)
		}

		result := reflect.New(reflect.TypeOf(item).Elem()).Interface()
		if err := b.Get(reflect.Indirect(rv).FieldByName("ID").String(), result); err != nil {
			t.Fatal(err)
		}

		result = reflect.Indirect(reflect.ValueOf(result)).Interface()
		if !deepEqual(reflect.Indirect(rv).Interface(), result) {
			t.Logf("\nExpected %#v\nbut got  %#v", reflect.Indirect(rv).Interface(), result)
			t.Fail()
		}
	}

	// Check items list
	sort := []string{"name"}

	sv := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(items[0]).Elem()), 0, 0)

	s := reflect.New(sv.Type())
	if _, err := b.List(s.Interface(), nil, sort, 0, 0); err != nil {
		t.Fatal(err)
	}
	checkItemList(s, items, t)

	s = reflect.New(sv.Type())
	if _, err := b.List(s.Interface(), nil, sort, 1, 1); err != nil {
		t.Fatal(err)
	}
	checkItemList(s, []interface{}{items[1]}, t)

	s = reflect.New(sv.Type())
	if _, err := b.List(s.Interface(), map[string]interface{}{"name": "glob:*2"}, sort, 0, 0); err != nil {
		t.Fatal(err)
	}
	checkItemList(s, []interface{}{items[1]}, t)

	s = reflect.New(sv.Type())
	if _, err := b.List(s.Interface(), map[string]interface{}{"name": "regexp:^[a-z]+[12]$"}, sort, 0, 0); err != nil {
		t.Fatal(err)
	}
	checkItemList(s, []interface{}{items[0], items[1]}, t)

	// Check item update
	if err := b.Add(update); err != nil {
		t.Fatal(err)
	}

	result := reflect.New(reflect.TypeOf(update).Elem()).Interface()
	if err := b.Get("00000000-0000-0000-0000-000000000000", result); err != nil {
		t.Fatal(err)
	}

	update = reflect.Indirect(reflect.ValueOf(update)).Interface()
	result = reflect.Indirect(reflect.ValueOf(result)).Interface()
	if !deepEqual(update, result) {
		t.Logf("\nExpected %#v\nbut got  %#v", update, result)
		t.Fail()
	}

	for i := len(items) - 1; i >= 0; i-- {
		// Check item deletion
		if err := b.Delete(items[i]); err != nil {
			t.Fatal(err)
		}
	}

	// Check for empty items list
	s = reflect.New(sv.Type())
	if _, err := b.List(s.Interface(), nil, nil, 0, 0); err != nil {
		t.Fatal(err)
	}

	if reflect.Indirect(s).Len() != 0 {
		t.Logf("\nExpected %d\nbut got  %d", 0, reflect.Indirect(s).Len())
		t.Fail()
	}
}

func checkItemList(rv reflect.Value, items []interface{}, t *testing.T) {
	if reflect.Indirect(rv).Len() != len(items) {
		t.Logf("\nExpected %d\nbut got  %d", len(items), reflect.Indirect(rv).Len())
		t.Fail()
		return
	}

	for i, item := range items {
		item = reflect.Indirect(reflect.ValueOf(item)).Interface()
		r := reflect.Indirect(rv).Index(i).Interface()
		if !deepEqual(item, r) {
			t.Logf("\nExpected %#v\nbut got  %#v", item, r)
			t.Fail()
		}
	}
}

func deepEqual(a, b interface{}) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if va.Kind() == reflect.Struct {
		if va.NumField() != vb.NumField() {
			return false
		}

		for i, n := 0, va.NumField(); i < n; i++ {
			ia := va.Field(i).Interface()
			ib := vb.Field(i).Interface()

			if ta, ok := ia.(time.Time); ok {
				if tb, ok := ib.(time.Time); !ok {
					return false
				} else if !ta.Equal(tb) {
					return false
				}
			}

			return deepEqual(ia, ib)
		}
	}

	return reflect.DeepEqual(a, b)
}
