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

// GetPlots retrieves time series data from provider based on a query and a time interval.
func (connector *InfluxDBConnector) GetPlots(query *plot.Query) ([]plot.Series, error) {
	var resultSeries = make([]plot.Series, 0)

	serieNames := make([]string, len(query.Group.Series))
	for i, serie := range query.Group.Series {
		serieNames[i] = connector.series[serie.Metric.Source][serie.Metric.Name]
	}

	influxdbQuery := fmt.Sprintf(
		"select value from %s where time > %ds and time < %ds order asc",
		strings.Join(serieNames, ","),
		query.StartTime.Unix(),
		query.EndTime.Unix(),
	)

	queryResult, err := connector.client.Query(influxdbQuery, "s")
	if err != nil {
		return nil, fmt.Errorf("influxdb[%s]: unable to perform query: %s", connector.name, err)
	}

	for i, influxdbSeries := range queryResult {
		series := plot.Series{
			Name:    query.Group.Series[i].Metric.Name,
			Summary: make(map[string]plot.Value),
		}

		for _, point := range influxdbSeries.GetPoints() {
			series.Plots = append(series.Plots, plot.Plot{
				Value: plot.Value(point[2].(float64)),
				Time:  time.Unix(int64(point[0].(float64)), 0),
			})
		}

		resultSeries = append(resultSeries, series)
	}

	if query.Group.Type == OperGroupTypeSum {
		sumSeries, err := plot.SumSeries(resultSeries)
		if err != nil {
			return nil, fmt.Errorf("influxdb[%s]: unable to sum series: %s", connector.name, err)
		}
		return []plot.Series{sumSeries}, nil
	} else if query.Group.Type == OperGroupTypeAvg {
		return nil, fmt.Errorf(
			"influxdb[%s]: average series grouping not supported by influxdb connector",
			connector.name,
		)
	} else {
		return resultSeries, nil
	}
}

// Refresh triggers a full connector data update.
func (connector *InfluxDBConnector) Refresh(originName string, outputChan chan *catalog.Record) error {
	var serieName, sourceName, metricName string

	result, err := connector.client.QueryWithNumbers("list series")
	if err != nil {
		return fmt.Errorf("influxdb[%s]: unable to fetch series list: %s", connector.name, err)
	}

	for _, serie := range result {
		serieName = serie.GetName()

		submatch := connector.re.FindStringSubmatch(serieName)
		if len(submatch) == 0 {
			logger.Log(logger.LevelInfo,
				"connector",
				"influxdb[%s]: serie `%s' does not match pattern, ignoring",
				connector.name,
				serieName,
			)
			return nil
		}

		if connector.re.SubexpNames()[1] == "source" {
			sourceName = submatch[1]
			metricName = submatch[2]
		} else {
			sourceName = submatch[2]
			metricName = submatch[1]
		}

		if _, ok := connector.series[sourceName]; !ok {
			connector.series[sourceName] = make(map[string]string)
		}

		connector.series[sourceName][metricName] = serieName

		outputChan <- &catalog.Record{
			Origin:    originName,
			Source:    sourceName,
			Metric:    metricName,
			Connector: connector,
		}
	}

	return nil
}
