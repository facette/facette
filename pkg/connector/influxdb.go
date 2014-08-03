// +build influxdb

package connector

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/plot"
	influxdb "github.com/facette/facette/thirdparty/github.com/influxdb/influxdb/client"
)

// InfluxDBConnector represents the main structure of the InfluxDB connector.
type InfluxDBConnector struct {
	name     string
	host     string
	username string
	password string
	database string
	client   *influxdb.Client
	re       *regexp.Regexp
	series   map[string]map[string]string
}

func init() {
	Connectors["influxdb"] = func(name string, settings map[string]interface{}) (Connector, error) {
		var (
			pattern string
			err     error
		)

		connector := &InfluxDBConnector{
			name:     name,
			host:     "localhost:8086",
			username: "root",
			password: "root",
			series:   make(map[string]map[string]string),
		}

		if connector.host, err = config.GetString(settings, "host", false); err != nil {
			return nil, err
		}

		if connector.username, err = config.GetString(settings, "username", false); err != nil {
			return nil, err
		}

		if connector.password, err = config.GetString(settings, "password", false); err != nil {
			return nil, err
		}

		if connector.database, err = config.GetString(settings, "database", true); err != nil {
			return nil, err
		}

		if pattern, err = config.GetString(settings, "pattern", true); err != nil {
			return nil, err
		}

		// Check and compile regexp pattern
		if connector.re, err = compilePattern(pattern); err != nil {
			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		}

		connector.client, err = influxdb.NewClient(&influxdb.ClientConfig{
			Host:     connector.host,
			Username: connector.username,
			Password: connector.password,
			Database: connector.database,
		})

		if err != nil {
			return nil, fmt.Errorf("unable to create client: %s", err)
		}

		return connector, nil
	}
}

// GetName returns the name of the current connector.
func (connector *InfluxDBConnector) GetName() string {
	return connector.name
}

// GetPlots retrieves time series data from provider based on a query and a time interval.
func (connector *InfluxDBConnector) GetPlots(query *plot.Query) ([]plot.Series, error) {
	seriesLength := len(query.Series)
	if seriesLength == 0 {
		return nil, fmt.Errorf("influxdb[%s]: requested series list is empty", connector.name)
	}

	metricsNames := make([]string, seriesLength)
	for i, series := range query.Series {
		metricsNames[i] = connector.series[series.Source][series.Metric]
	}

	influxdbQuery := fmt.Sprintf(
		"select value from %s where time > %ds and time < %ds order asc",
		strings.Join(metricsNames, ","),
		query.StartTime.Unix(),
		query.EndTime.Unix(),
	)

	queryResult, err := connector.client.Query(influxdbQuery, "s")
	if err != nil {
		return nil, fmt.Errorf("influxdb[%s]: unable to perform query: %s", connector.name, err)
	}

	resultSeries := make([]plot.Series, 0)

	for i, influxdbSeries := range queryResult {
		series := plot.Series{
			Name:    query.Series[i].Name,
			Summary: make(map[string]plot.Value),
			Step:    int(query.EndTime.Sub(query.StartTime) / time.Duration(query.Sample)),
		}

		for _, point := range influxdbSeries.GetPoints() {
			series.Plots = append(series.Plots, plot.Plot{
				Value: plot.Value(point[2].(float64)),
				Time:  time.Unix(int64(point[0].(float64)), 0),
			})
		}

		resultSeries = append(resultSeries, series)
	}

	return resultSeries, nil
}

// Refresh triggers a full connector data update.
func (connector *InfluxDBConnector) Refresh(originName string, outputChan chan *catalog.Record) error {
	seriesList, err := connector.client.QueryWithNumbers("list series")
	if err != nil {
		return fmt.Errorf("influxdb[%s]: unable to fetch series list: %s", connector.name, err)
	}

	for _, series := range seriesList {
		var seriesName, sourceName, metricName string

		seriesName = series.GetName()

		seriesMatch, err := matchSeriesPattern(connector.re, seriesName)
		if err != nil {
			logger.Log(logger.LevelInfo,
				"connector",
				"influxdb[%s]: series `%s' does not match pattern, ignoring",
				connector.name,
				seriesName,
			)
			continue
		}

		sourceName, metricName = seriesMatch[0], seriesMatch[1]

		if _, ok := connector.series[sourceName]; !ok {
			connector.series[sourceName] = make(map[string]string)
		}

		connector.series[sourceName][metricName] = seriesName

		outputChan <- &catalog.Record{
			Origin:    originName,
			Source:    sourceName,
			Metric:    metricName,
			Connector: connector,
		}
	}

	return nil
}
