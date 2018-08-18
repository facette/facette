// +build !disable_connector_influxdb

package connector

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"facette.io/facette/catalog"
	"facette.io/facette/series"
	"facette.io/facette/version"
	"facette.io/logger"
	"facette.io/maputil"
	influxdb "github.com/influxdata/influxdb/client/v2"
	influxmodels "github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxql"
	"github.com/pkg/errors"
)

func init() {
	connectors["influxdb"] = func(name string, settings *maputil.Map, logger *logger.Logger) (Connector, error) {
		var (
			pattern string
			mapping maputil.Map
			glue    string
			err     error
		)

		c := &influxDBConnector{
			name: name,
			mapping: &influxDBMapping{
				Glue: ".",
			},
			logger: logger,
		}

		// Load provider configuration
		c.url, err = settings.GetString("url", "")
		if err != nil {
			return nil, err
		} else if c.url == "" {
			return nil, ErrMissingConnectorSetting("url")
		}
		c.url = normalizeURL(c.url)

		c.timeout, err = settings.GetInt("timeout", defaultTimeout)
		if err != nil {
			return nil, err
		}

		c.allowInsecure, err = settings.GetBool("allow_insecure_tls", false)
		if err != nil {
			return nil, err
		}

		c.username, err = settings.GetString("username", "")
		if err != nil {
			return nil, err
		}

		c.password, err = settings.GetString("password", "")
		if err != nil {
			return nil, err
		}

		c.database, err = settings.GetString("database", "")
		if err != nil {
			return nil, err
		} else if c.database == "" {
			return nil, ErrMissingConnectorSetting("database")
		}

		pattern, err = settings.GetString("pattern", "")
		if err != nil {
			return nil, err
		}

		mapping, err = settings.GetMap("mapping", nil)
		if err != nil {
			return nil, err
		}

		if pattern != "" && mapping != nil {
			return nil, fmt.Errorf("connector settings allows either \"pattern\n or \"mapping\", not both")
		} else if pattern == "" && mapping == nil {
			return nil, fmt.Errorf("missing \"pattern\" or \"mapping\" connector settings")
		}

		if pattern != "" {
			c.pattern, err = compilePattern(pattern)
			if err != nil {
				return nil, err
			}
		} else if mapping != nil {
			c.mapping.Source, err = mapping.GetStringSlice("source", nil)
			if err != nil {
				return nil, err
			}

			c.mapping.Metric, err = mapping.GetStringSlice("metric", nil)
			if err != nil {
				return nil, err
			}

			glue, err = mapping.GetString("glue", ".")
			if err != nil {
				return nil, err
			} else if glue != "" {
				c.mapping.Glue = glue
			}
		}

		// Check remote instance URL
		_, err = url.Parse(c.url)
		if err != nil {
			return nil, fmt.Errorf("unable to parse URL: %s", err)
		}

		c.client, err = influxdb.NewHTTPClient(influxdb.HTTPConfig{
			Addr:               c.url,
			Username:           c.username,
			Password:           c.password,
			UserAgent:          "facette/" + version.Version,
			Timeout:            time.Duration(c.timeout) * time.Second,
			InsecureSkipVerify: c.allowInsecure,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create client: %s", err)
		}

		return c, nil
	}

	// Register type for catalog dump
	gob.Register(map[string]string{})
}

type influxDBConnector struct {
	name          string
	url           string
	timeout       int
	allowInsecure bool
	username      string
	password      string
	database      string
	pattern       *regexp.Regexp
	mapping       *influxDBMapping
	client        influxdb.Client
	logger        *logger.Logger
}

func (c *influxDBConnector) Name() string {
	return c.name
}

func (c *influxDBConnector) Points(q *series.Query) ([]series.Series, error) {
	var queries []string

	l := len(q.Metrics)
	if l == 0 {
		return nil, fmt.Errorf("requested metrics list is empty")
	}

	results := make([]series.Series, l)

	// Prepare query
	for _, m := range q.Metrics {
		var (
			terms  map[string]string
			series string
			parts  []string
		)

		column, err := m.Attributes.GetString("column", "")
		if err != nil || column == "" {
			return nil, errors.Wrap(ErrInvalidAttribute, "column")
		}

		if v, err := m.Attributes.GetInterface("terms", nil); err != nil {
			return nil, errors.Wrap(ErrInvalidAttribute, "terms")
		} else if v, ok := v.(map[string]string); !ok {
			return nil, errors.Wrap(ErrInvalidAttribute, "terms")
		} else {
			terms = v
		}

		for term, value := range terms {
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
			column, strconv.Quote(series),
			strings.Join(parts, " and "),
		))
	}

	query := influxdb.Query{
		Command:  strings.Join(queries, "; "),
		Database: c.database,
	}

	// Execute query
	resp, err := c.client.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch points: %s", err)
	} else if resp.Error() != nil {
		return nil, fmt.Errorf("failed to fetch points: %s", resp.Error())
	}

	// Parse results received from back-end
	for i, r := range resp.Results {
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

func (c *influxDBConnector) Refresh(output chan<- *catalog.Record) error {
	// Query back-end for sample rows (used to detect numerical values)
	columnsMap := make(map[string][]string)

	q := influxdb.Query{
		Command:  "select * from /.*/ limit 1",
		Database: c.database,
	}

	resp, err := c.client.Query(q)
	if err != nil {
		return fmt.Errorf("failed to fetch sample rows: %s", err)
	} else if resp.Error() != nil {
		return fmt.Errorf("failed to fetch sample rows: %s", resp.Error())
	}

	if len(resp.Results) != 1 {
		return fmt.Errorf("failed to retrieve sample rows: expected 1 result but got %d", len(resp.Results))
	} else if resp.Results[0].Err != "" {
		return fmt.Errorf("failed to retrieve sample rows: %s", resp.Results[0].Err)
	}

	for _, s := range resp.Results[0].Series {
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
				seriesMatch, err := matchPattern(c.pattern, series)
				if err != nil {
					c.logger.Warning("%s", err)
					continue
				}

				output <- &catalog.Record{
					Origin: c.name,
					Source: seriesMatch[0],
					Metric: seriesMatch[1] + c.mapping.Glue + metric,
					Attributes: &maputil.Map{
						"column": metric,
						"terms":  map[string]string{"": seriesMatch[0] + c.mapping.Glue + seriesMatch[1]},
					},
				}
			}
		}
	} else { // Column-based mapping
		// Query back-end for series list
		q = influxdb.Query{
			Command:  "show series",
			Database: c.database,
		}

		resp, err = c.client.Query(q)
		if err != nil {
			return fmt.Errorf("failed to fetch series: %s", err)
		} else if resp.Error() != nil {
			return fmt.Errorf("failed to fetch series: %s", resp.Error())
		}

		if len(resp.Results) != 1 {
			return fmt.Errorf("failed to retrieve series: expected 1 result but got %d", len(resp.Results))
		} else if resp.Results[0].Err != "" {
			return fmt.Errorf("failed to retrieve series: %s", resp.Results[0].Err)
		}

		// Parse results for sources and metrics
		for _, s := range resp.Results[0].Series {
			for i := range s.Values {
				seriesColumns, err := mapSeriesColumns(s.Values[i][0].(string))
				if err != nil {
					return fmt.Errorf("failed to map columns: %s", err)
				}

				var parts []string

				terms := make(map[string]string)

				// Map source
				for _, item := range c.mapping.Source {
					term, part := mapKey(seriesColumns, item)
					if part != "" {
						terms[term] = part
						parts = append(parts, part)
					}
				}
				sourceName := strings.Join(parts, c.mapping.Glue)

				// Map metric
				parts = []string{}
				for _, item := range c.mapping.Metric {
					term, part := mapKey(seriesColumns, item)
					if part != "" {
						terms[term] = part
						parts = append(parts, part)
					}
				}
				metricName := strings.Join(parts, c.mapping.Glue)

				terms[""] = seriesColumns["name"]

				for _, column := range columnsMap[seriesColumns["name"]] {
					output <- &catalog.Record{
						Origin: c.name,
						Source: sourceName,
						Metric: metricName + c.mapping.Glue + column,
						Attributes: &maputil.Map{
							"column": column,
							"terms":  terms,
						},
					}
				}
			}
		}
	}

	return nil
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

func mapSeriesColumns(series string) (map[string]string, error) {
	idx := strings.Index(series, ",")

	tags := influxmodels.ParseTags([]byte(series))

	columns := tags.Map()
	columns["name"] = series[:idx]

	return columns, nil
}

type influxDBMapping struct {
	Source []string
	Metric []string
	Glue   string
}
