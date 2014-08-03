// +build rrd

package connector

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/plot"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/ziutek/rrd"
)

type rrdMetric struct {
	Dataset  string
	FilePath string
	Step     time.Duration
}

// RRDConnector represents the main structure of the RRD connector.
type RRDConnector struct {
	name    string
	path    string
	daemon  string
	re      *regexp.Regexp
	metrics map[string]map[string]*rrdMetric
}

func init() {
	Connectors["rrd"] = func(name string, settings map[string]interface{}) (Connector, error) {
		var (
			pattern string
			err     error
		)

		connector := &RRDConnector{
			name:    name,
			metrics: make(map[string]map[string]*rrdMetric),
		}

		if connector.path, err = config.GetString(settings, "path", true); err != nil {
			return nil, err
		}

		if connector.daemon, err = config.GetString(settings, "daemon", false); err != nil {
			return nil, err
		}

		if pattern, err = config.GetString(settings, "pattern", true); err != nil {
			return nil, err
		}

		// Check and compile regexp pattern
		if connector.re, err = compilePattern(pattern); err != nil {
			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		}

		return connector, nil
	}
}

// GetName returns the name of the current connector.
func (connector *RRDConnector) GetName() string {
	return connector.name
}

// GetPlots retrieves time series data from origin based on a query and a time interval.
func (connector *RRDConnector) GetPlots(query *plot.Query) ([]plot.Series, error) {
	var (
		resultSeries []plot.Series
		xport        *rrd.Exporter
	)

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("rrd[%s]: requested series list is empty", connector.name)
	}

	graph := rrd.NewGrapher()

	if connector.daemon != "" {
		graph.SetDaemon(connector.daemon)
	}

	xport = rrd.NewExporter()

	if connector.daemon != "" {
		xport.SetDaemon(connector.daemon)
	}

	step := time.Duration(0)

	for _, series := range query.Series {
		graph.Def(
			series.Name+"-def0",
			connector.metrics[series.Source][series.Metric].FilePath,
			connector.metrics[series.Source][series.Metric].Dataset,
			"AVERAGE",
		)

		graph.CDef(series.Name, series.Name+"-def0")

		// Set plots request
		xport.Def(
			series.Name+"-def0",
			connector.metrics[series.Source][series.Metric].FilePath,
			connector.metrics[series.Source][series.Metric].Dataset,
			"AVERAGE",
		)

		xport.CDef(series.Name, series.Name+"-def0")

		xport.XportDef(series.Name, series.Name)

		if connector.metrics[series.Source][series.Metric].Step > step {
			step = connector.metrics[series.Source][series.Metric].Step
		}
	}

	// Get plots
	if step == 0 {
		step = query.EndTime.Sub(query.StartTime) / time.Duration(config.DefaultPlotSample)
	}

	data := rrd.XportResult{}

	data, err := xport.Xport(query.StartTime, query.EndTime, step)
	if err != nil {
		return nil, err
	}

	for index, itemName := range data.Legends {
		series := plot.Series{
			Name:    itemName,
			Summary: make(map[string]plot.Value),
		}

		// FIXME: skip last garbage entry (see https://github.com/ziutek/rrd/pull/13)
		for i := 0; i < data.RowCnt-1; i++ {
			series.Plots = append(
				series.Plots,
				plot.Plot{
					Value: plot.Value(data.ValueAt(index, i)),
					Time:  query.StartTime.Add(data.Step * time.Duration(i)),
				},
			)
		}

		resultSeries = append(resultSeries, series)
	}

	data.FreeValues()

	return resultSeries, nil
}

// Refresh triggers a full connector data update.
func (connector *RRDConnector) Refresh(originName string, outputChan chan *catalog.Record) error {
	// Search for files and parse their path for source/metric pairs
	walkFunc := func(filePath string, fileInfo os.FileInfo, err error) error {
		var sourceName, metricName string

		// Stop if previous error
		if err != nil {
			return err
		}

		// Skip non-files
		mode := fileInfo.Mode() & os.ModeType
		if mode != 0 {
			return nil
		}

		seriesMatch, err := matchSeriesPattern(connector.re, filePath[len(connector.path)+1:])
		if err != nil {
			logger.Log(
				logger.LevelInfo,
				"connector",
				"rrd[%s]: file `%s' does not match pattern, ignoring",
				connector.name,
				filePath,
			)
			return nil
		}

		sourceName, metricName = seriesMatch[0], seriesMatch[1]

		if _, ok := connector.metrics[sourceName]; !ok {
			connector.metrics[sourceName] = make(map[string]*rrdMetric)
		}

		// Extract metric information from .rrd file
		info, err := rrd.Info(filePath)
		if err != nil {
			logger.Log(logger.LevelWarning, "connector", "rrd[%s]: %s", connector.name, err)
			return nil
		}

		if _, ok := info["ds.index"]; ok {
			for dsName := range info["ds.index"].(map[string]interface{}) {
				metricFullName := metricName + "/" + dsName

				connector.metrics[sourceName][metricFullName] = &rrdMetric{
					Dataset:  dsName,
					FilePath: filePath,
					Step:     time.Duration(info["step"].(uint)) * time.Second,
				}

				outputChan <- &catalog.Record{
					Origin:    originName,
					Source:    sourceName,
					Metric:    metricFullName,
					Connector: connector,
				}
			}
		}

		return nil
	}

	if err := utils.WalkDir(connector.path, walkFunc); err != nil {
		return err
	}

	return nil
}
