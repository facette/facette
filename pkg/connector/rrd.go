// +build rrd

package connector

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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

// GetPlots retrieves time series data from origin based on a query and a time interval.
func (connector *RRDConnector) GetPlots(query *plot.Query) ([]plot.Series, error) {
	var (
		resultSeries []plot.Series
		stack        []string
		xport        *rrd.Exporter
	)

	if len(query.Group.Series) == 0 {
		return nil, fmt.Errorf("rrd[%s]: group has no series", connector.name)
	} else if query.Group.Type != OperGroupTypeNone && len(query.Group.Series) == 1 {
		query.Group.Type = OperGroupTypeNone
	}

	graph := rrd.NewGrapher()

	if connector.daemon != "" {
		graph.SetDaemon(connector.daemon)
	}

	xport = rrd.NewExporter()

	if connector.daemon != "" {
		xport.SetDaemon(connector.daemon)
	}

	count := 0
	step := time.Duration(0)

	switch query.Group.Type {
	case OperGroupTypeNone:
		for _, series := range query.Group.Series {
			if series.Metric == nil {
				continue
			}

			itemName := fmt.Sprintf("series%d", count)
			count++

			graph.Def(
				itemName+"-orig0",
				connector.metrics[series.Metric.Source][series.Metric.Name].FilePath,
				connector.metrics[series.Metric.Source][series.Metric.Name].Dataset,
				"AVERAGE",
			)

			seriesScale, _ := config.GetFloat(series.Options, "scale", false)
			groupScale, _ := config.GetFloat(query.Group.Options, "scale", false)

			if seriesScale != 0 {
				graph.CDef(itemName+"-orig1", fmt.Sprintf("%s-orig0,%g,*", itemName, seriesScale))
			} else {
				graph.CDef(itemName+"-orig1", itemName+"-orig0")
			}

			if groupScale != 0 {
				graph.CDef(itemName, fmt.Sprintf("%s-orig1,%g,*", itemName, groupScale))
			} else {
				graph.CDef(itemName, itemName+"-orig1")
			}

			// Set plots request
			xport.Def(
				itemName+"-orig0",
				connector.metrics[series.Metric.Source][series.Metric.Name].FilePath,
				connector.metrics[series.Metric.Source][series.Metric.Name].Dataset,
				"AVERAGE",
			)

			if seriesScale != 0 {
				xport.CDef(itemName+"-orig1", fmt.Sprintf("%s-orig0,%g,*", itemName, seriesScale))
			} else {
				xport.CDef(itemName+"-orig1", itemName+"-orig0")
			}

			if groupScale != 0 {
				xport.CDef(itemName, fmt.Sprintf("%s-orig1,%g,*", itemName, groupScale))
			} else {
				xport.CDef(itemName, itemName+"-orig1")
			}

			xport.XportDef(itemName, itemName)

			if connector.metrics[series.Metric.Source][series.Metric.Name].Step > step {
				step = connector.metrics[series.Metric.Source][series.Metric.Name].Step
			}
		}

	case OperGroupTypeAvg, OperGroupTypeSum:
		itemName := fmt.Sprintf("series%d", count)
		count++

		for index, series := range query.Group.Series {
			if series.Metric == nil {
				continue
			}

			seriesTemp := itemName + fmt.Sprintf("-tmp%d", index)

			graph.Def(
				seriesTemp+"-ori",
				connector.metrics[series.Metric.Source][series.Metric.Name].FilePath,
				connector.metrics[series.Metric.Source][series.Metric.Name].Dataset,
				"AVERAGE",
			)

			graph.CDef(seriesTemp, fmt.Sprintf("%s-ori,UN,0,%s-ori,IF", seriesTemp, seriesTemp))

			xport.Def(
				seriesTemp+"-ori",
				connector.metrics[series.Metric.Source][series.Metric.Name].FilePath,
				connector.metrics[series.Metric.Source][series.Metric.Name].Dataset,
				"AVERAGE",
			)

			xport.CDef(seriesTemp, fmt.Sprintf("%s-ori,UN,0,%s-ori,IF", seriesTemp, seriesTemp))

			if len(stack) == 0 {
				stack = append(stack, seriesTemp)
			} else {
				stack = append(stack, seriesTemp, "+")
			}

			if connector.metrics[series.Metric.Source][series.Metric.Name].Step > step {
				step = connector.metrics[series.Metric.Source][series.Metric.Name].Step
			}
		}

		if query.Group.Type == OperGroupTypeAvg {
			stack = append(stack, strconv.Itoa(len(query.Group.Series)), "/")
		}

		groupScale, _ := config.GetFloat(query.Group.Options, "scale", false)

		graph.CDef(itemName+"-orig", strings.Join(stack, ","))

		if groupScale != 0 {
			graph.CDef(itemName, fmt.Sprintf("%s-orig,%g,*", itemName, groupScale))
		} else {
			graph.CDef(itemName, itemName+"-orig")
		}

		// Set plots request
		xport.CDef(itemName+"-orig", strings.Join(stack, ","))

		if groupScale != 0 {
			xport.CDef(itemName, fmt.Sprintf("%s-orig,%g,*", itemName, groupScale))
		} else {
			xport.CDef(itemName, itemName+"-orig")
		}

		xport.XportDef(itemName, itemName)

	default:
		return nil, fmt.Errorf("rrd[%s]: unknown operator type %d", connector.name, query.Group.Type)
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
					Time:  query.StartTime.Add(time.Duration(i) * step * time.Second),
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
					Step:     time.Duration(info["step"].(uint)),
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
