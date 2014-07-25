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
	path    string
	pattern string
	daemon  string
	metrics map[string]map[string]*rrdMetric
}

func init() {
	Connectors["rrd"] = func(settings map[string]interface{}) (Connector, error) {
		var err error

		connector := &RRDConnector{
			metrics: make(map[string]map[string]*rrdMetric),
		}

		if connector.path, err = config.GetString(settings, "path", true); err != nil {
			return nil, err
		}

		if connector.pattern, err = config.GetString(settings, "pattern", true); err != nil {
			return nil, err
		}

		if connector.daemon, err = config.GetString(settings, "daemon", false); err != nil {
			return nil, err
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
		return nil, fmt.Errorf("group has no series")
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
		return nil, fmt.Errorf("unknown operator type %d", query.Group.Type)
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
	// Compile pattern
	re := regexp.MustCompile(connector.pattern)

	// Validate pattern keywords
	groups := make(map[string]bool)

	for _, key := range re.SubexpNames() {
		if key == "" {
			continue
		} else if key == "source" || key == "metric" {
			groups[key] = true
		} else {
			return fmt.Errorf("invalid pattern keyword `%s'", key)
		}
	}

	if !groups["source"] {
		return fmt.Errorf("missing pattern keyword `source'")
	} else if !groups["metric"] {
		return fmt.Errorf("missing pattern keyword `metric'")
	}

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

		submatch := re.FindStringSubmatch(filePath[len(connector.path)+1:])
		if len(submatch) == 0 {
			logger.Log(logger.LevelInfo, "connector: rrd", "file `%s' does not match pattern, ignoring", filePath)
			return nil
		}

		if re.SubexpNames()[1] == "source" {
			sourceName = submatch[1]
			metricName = submatch[2]
		} else {
			sourceName = submatch[2]
			metricName = submatch[1]
		}

		if _, ok := connector.metrics[sourceName]; !ok {
			connector.metrics[sourceName] = make(map[string]*rrdMetric)
		}

		// Extract metric information from .rrd file
		info, err := rrd.Info(filePath)
		if err != nil {
			logger.Log(logger.LevelWarning, "connector: rrd", "%s", err)
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
