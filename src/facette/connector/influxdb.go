// +build !disable_connector_influxdb

package connector

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"facette/catalog"
	"facette/series"

	"github.com/facette/logger"
	"github.com/facette/maputil"
	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/influxql"
)

type influxDBMap struct {
	source []string
	metric []string
	glue   string
	maps   map[string]map[string]influxDBMapEntry
}

type influxDBMapEntry struct {
	column string
	terms  map[string]string
}

// influxdbConnector implements the connector handler for another InfluxDB instance.
type influxdbConnector struct {
	name          string
	url           string
	timeout       int
	allowInsecure bool
	username      string
	password      string
	database      string
	pattern       *regexp.Regexp
	mapping       influxDBMap
	client        influxdb.Client
}

func init() {
	connectors["influxdb"] = func(name string, settings *maputil.Map, log *logger.Logger) (Connector, error) {
		var err error

		c := &influxdbConnector{
			name: name,
			mapping: influxDBMap{
				glue: ".",
				maps: make(map[string]map[string]influxDBMapEntry),
			},
		}

		// Load provider configuration
		if c.url, err = settings.GetString("url", ""); err != nil {
			return nil, err
		} else if c.url == "" {
			return nil, ErrMissingConnectorSetting("url")
		}
		normalizeURL(&c.url)

		if c.timeout, err = settings.GetInt("timeout", connectorDefaultTimeout); err != nil {
			return nil, err
		}

		if c.allowInsecure, err = settings.GetBool("allow_insecure_tls", false); err != nil {
			return nil, err
		}

		if c.username, err = settings.GetString("username", ""); err != nil {
			return nil, err
		}

		if c.password, err = settings.GetString("password", ""); err != nil {
			return nil, err
		}

		if c.database, err = settings.GetString("database", ""); err != nil {
			return nil, err
		} else if c.database == "" {
			return nil, ErrMissingConnectorSetting("database")
		}

		if v, _ := settings.GetString("pattern", ""); v != "" && settings.Has("mapping") {
			return nil, fmt.Errorf("connector settings allows either %q or %q, not both", "pattern", "mapping")
		} else if !settings.Has("pattern") && !settings.Has("mapping") {
			return nil, fmt.Errorf("connector settings %q or %q must be specified", "pattern", "mapping")
		}

		pattern, err := settings.GetString("pattern", "")
		if err != nil {
			return nil, err
		}

		if pattern != "" {
			if c.pattern, err = compilePattern(pattern); err != nil {
				return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
			}
		}

		mapping, err := settings.GetMap("mapping", nil)
		if err != nil {
			return nil, err
		}

		if mapping != nil {
			if c.mapping.source, err = mapping.GetStringSlice("source", nil); err != nil {
				return nil, err
			}

			if c.mapping.metric, err = mapping.GetStringSlice("metric", nil); err != nil {
				return nil, err
			}

			glue, err := mapping.GetString("glue", ".")
			if err != nil {
				return nil, err
			} else if glue != "" {
				c.mapping.glue = glue
			}
		}

		// Check remote instance URL
		if _, err := url.Parse(c.url); err != nil {
			return nil, fmt.Errorf("unable to parse URL: %s", err)
		}

		// Create new client instance
		c.client, err = influxdb.NewHTTPClient(influxdb.HTTPConfig{
			Addr:               c.url,
			Username:           c.username,
			Password:           c.password,
			Timeout:            time.Duration(c.timeout) * time.Second,
			InsecureSkipVerify: c.allowInsecure,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create client: %s", err)
		}

		return c, nil
	}
}

// Name returns the name of the current connector.
func (c *influxdbConnector) Name() string {
	return c.name
}

// Refresh triggers the connector data refresh.
func (c *influxdbConnector) Refresh(output chan<- *catalog.Record) error {
	// Query back-end for sample rows (used to detect numerical values)
	columnsMap := make(map[string][]string, 0)

	q := influxdb.Query{
		Command:  "select * from /.*/ limit 1",
		Database: c.database,
	}

	response, err := c.client.Query(q)
	if err != nil {
		return fmt.Errorf("failed to fetch sample rows: %s", err)
	} else if response.Error() != nil {
		return fmt.Errorf("failed to fetch sample rows: %s", response.Error())
	}

	if len(response.Results) != 1 {
		return fmt.Errorf("failed to retrieve sample rows: expected 1 result but got %d", len(response.Results))
	} else if response.Results[0].Err != "" {
		return fmt.Errorf("failed to retrieve sample rows: %s", response.Results[0].Err)
	}

	for _, s := range response.Results[0].Series {
		if len(s.Values) == 0 {
			continue
		}

		if _, ok := columnsMap[s.Name]; !ok {
			columnsMap[s.Name] = make([]string, 0)
		}

		for i, v := range s.Values[0] {
			if _, ok := v.(json.Number); !ok {
				continue
			}

			columnsMap[s.Name] = append(columnsMap[s.Name], s.Columns[i])
		}
	}

	if c.pattern != nil { // Pattern-based mapping
		for series, metricColumns := range columnsMap {
			for _, metric := range metricColumns {
				// FIXME: we should return the matchPattern() error to the caller via the eventChan
				seriesMatch, _ := matchPattern(c.pattern, series)

				if _, ok := c.mapping.maps[seriesMatch[0]]; !ok {
					c.mapping.maps[seriesMatch[0]] = make(map[string]influxDBMapEntry)
				}

				c.mapping.maps[seriesMatch[0]][seriesMatch[1]+c.mapping.glue+metric] = influxDBMapEntry{
					terms:  map[string]string{"": seriesMatch[0] + c.mapping.glue + seriesMatch[1]},
					column: metric,
				}

				// Send record to catalog
				output <- &catalog.Record{
					Origin:    c.name,
					Source:    seriesMatch[0],
					Metric:    seriesMatch[1] + c.mapping.glue + metric,
					Connector: c,
				}
			}
		}
	} else { // Column-based mapping
		// Query back-end for series list
		q = influxdb.Query{
			Command:  "show series",
			Database: c.database,
		}

		response, err = c.client.Query(q)
		if err != nil {
			return fmt.Errorf("failed to fetch series: %s", err)
		} else if response.Error() != nil {
			return fmt.Errorf("failed to fetch series: %s", response.Error())
		}

		if len(response.Results) != 1 {
			return fmt.Errorf("failed to retrieve series: expected 1 result but got %d", len(response.Results))
		} else if response.Results[0].Err != "" {
			return fmt.Errorf("failed to retrieve series: %s", response.Results[0].Err)
		}

		// Parse results for sources and metrics
		for _, s := range response.Results[0].Series {
			for i := range s.Values {
				seriesColumns := mapSeriesColumns(s.Values[i][0].(string))

				var parts []string

				terms := make(map[string]string)

				// Map source
				for _, item := range c.mapping.source {
					term, part := mapKey(seriesColumns, item)
					if part != "" {
						terms[term] = part
						parts = append(parts, part)
					}
				}
				sourceName := strings.Join(parts, c.mapping.glue)

				// Map metric
				parts = []string{}
				for _, item := range c.mapping.metric {
					term, part := mapKey(seriesColumns, item)
					if part != "" {
						terms[term] = part
						parts = append(parts, part)
					}
				}
				metricName := strings.Join(parts, c.mapping.glue)

				terms[""] = seriesColumns["name"]

				// Initialize metric mapping terms if needed
				if _, ok := c.mapping.maps[sourceName]; !ok {
					c.mapping.maps[sourceName] = make(map[string]influxDBMapEntry)
				}

				for _, col := range columnsMap[seriesColumns["name"]] {
					c.mapping.maps[sourceName][metricName+c.mapping.glue+col] = influxDBMapEntry{
						column: col,
						terms:  terms,
					}

					// Send record to catalog
					output <- &catalog.Record{
						Origin:    c.name,
						Source:    sourceName,
						Metric:    metricName + c.mapping.glue + col,
						Connector: c,
					}
				}
			}
		}
	}

	return nil
}

// Points retrieves the time series data according to the query parameters and a time interval.
func (c *influxdbConnector) Points(q *series.Query) ([]series.Series, error) {
	var queries []string

	l := len(q.Series)
	if l == 0 {
		return nil, fmt.Errorf("influxdb[%s]: requested series list is empty", c.name)
	}

	results := make([]series.Series, l)

	// Prepare query
	for _, s := range q.Series {
		var (
			series string
			parts  []string
		)

		if _, ok := c.mapping.maps[s.Source]; !ok {
			return nil, fmt.Errorf("unknown series source `%s'", s.Source)
		} else if _, ok := c.mapping.maps[s.Source][s.Metric]; !ok {
			return nil, fmt.Errorf("unknown series metric `%s' for source `%s'", s.Source, s.Metric)
		}

		mapping := c.mapping.maps[s.Source][s.Metric]

		for term, value := range mapping.terms {
			if term == "" {
				series = value
			} else {
				parts = append(parts, fmt.Sprintf("%s = %s", influxql.QuoteIdent(term), influxql.QuoteString(value)))
			}
		}

		parts = append(parts, fmt.Sprintf(
			"time > %ds and time < %ds order by asc",
			q.StartTime.Unix(),
			q.EndTime.Unix(),
		))

		queries = append(queries, fmt.Sprintf(
			"select %s, time from %s where %s",
			mapping.column, strconv.Quote(series),
			strings.Join(parts, " and "),
		))
	}

	query := influxdb.Query{
		Command:  strings.Join(queries, "; "),
		Database: c.database,
	}

	// Execute query
	response, err := c.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch points: %s", err)
	} else if response.Error() != nil {
		return nil, fmt.Errorf("failed to fetch points: %s", response.Error())
	}

	// Parse results received from back-end
	for i, r := range response.Results {
		if r.Err != "" {
			continue
		}

		results[i] = series.Series{}
		for _, s := range r.Series {
			for _, v := range s.Values {
				time, err := time.Parse(time.RFC3339Nano, v[0].(string))
				if err != nil {
					return nil, fmt.Errorf("failed to parse time: %s", v[0])
				}

				value, err := v[1].(json.Number).Float64()
				if err != nil {
					return nil, fmt.Errorf("failed to parse value: %s", v[1])
				}

				results[i].Points = append(results[i].Points, series.Point{
					Time:  time,
					Value: series.Value(value),
				})
			}
		}
	}

	return results, nil
}

func mapKey(seriesColumns map[string]string, item string) (string, string) {
	if item == "name" {
		return "", seriesColumns["name"]
	} else if strings.HasPrefix(item, "column:") {
		// Try to match row column
		name := strings.TrimPrefix(item, "column:")
		if value, ok := seriesColumns[name]; ok {
			return name, value
		}
	}

	// Nothing matched
	return "", ""
}

func mapSeriesColumns(series string) map[string]string {
	columns := make(map[string]string)

	// From InfluxDB documentation of "SHOW SERIES":
	// 		"Everything before the first comma is the measurement name.
	// 		Everything after the first comma is either a tag key or a tag value."

	columns["name"] = series[:strings.Index(series, ",")]
	for _, col := range strings.Split(series[strings.Index(series, ",")+1:], ",") {
		kv := strings.Split(col, "=")
		columns[kv[0]] = kv[1]
	}

	return columns
}
