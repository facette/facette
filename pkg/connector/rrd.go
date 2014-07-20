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
	"github.com/facette/facette/pkg/types"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/ziutek/rrd"
)

type rrdMetric struct {
	Dataset  string
	FilePath string
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
func (connector *RRDConnector) GetPlots(query *types.PlotQuery) ([]*types.PlotResult, error) {
	var xport *rrd.Exporter

	if len(query.Group.Series) == 0 {
		return nil, fmt.Errorf("group has no series")
	} else if query.Group.Type != OperGroupTypeNone && len(query.Group.Series) == 1 {
		query.Group.Type = OperGroupTypeNone
	}

	result := make([]*types.PlotResult, 0)

	stack := make([]string, 0)

	graph := rrd.NewGrapher()

	if connector.daemon != "" {
		graph.SetDaemon(connector.daemon)
	}

	xport = rrd.NewExporter()

	if connector.daemon != "" {
		xport.SetDaemon(connector.daemon)
	}

	count := 0

	switch query.Group.Type {
	case OperGroupTypeNone:
		for _, serie := range query.Group.Series {
			if serie.Metric == nil {
				continue
			}

			itemName := fmt.Sprintf("serie%d", count)
			count += 1

			graph.Def(
				itemName+"-orig0",
				connector.metrics[serie.Metric.Source][serie.Metric.Name].FilePath,
				connector.metrics[serie.Metric.Source][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			serieScale, _ := config.GetFloat(serie.Options, "scale", false)
			groupScale, _ := config.GetFloat(query.Group.Options, "scale", false)

			if serieScale != 0 {
				graph.CDef(itemName+"-orig1", fmt.Sprintf("%s-orig0,%g,*", itemName, serieScale))
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
				connector.metrics[serie.Metric.Source][serie.Metric.Name].FilePath,
				connector.metrics[serie.Metric.Source][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			if serieScale != 0 {
				xport.CDef(itemName+"-orig1", fmt.Sprintf("%s-orig0,%g,*", itemName, serieScale))
			} else {
				xport.CDef(itemName+"-orig1", itemName+"-orig0")
			}

			if groupScale != 0 {
				xport.CDef(itemName, fmt.Sprintf("%s-orig1,%g,*", itemName, groupScale))
			} else {
				xport.CDef(itemName, itemName+"-orig1")
			}

			xport.XportDef(itemName, itemName)
		}

	case OperGroupTypeAvg, OperGroupTypeSum:
		itemName := fmt.Sprintf("serie%d", count)
		count += 1

		for index, serie := range query.Group.Series {
			if serie.Metric == nil {
				continue
			}

			serieTemp := itemName + fmt.Sprintf("-tmp%d", index)

			graph.Def(
				serieTemp+"-ori",
				connector.metrics[serie.Metric.Source][serie.Metric.Name].FilePath,
				connector.metrics[serie.Metric.Source][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			graph.CDef(serieTemp, fmt.Sprintf("%s-ori,UN,0,%s-ori,IF", serieTemp, serieTemp))

			xport.Def(
				serieTemp+"-ori",
				connector.metrics[serie.Metric.Source][serie.Metric.Name].FilePath,
				connector.metrics[serie.Metric.Source][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			xport.CDef(serieTemp, fmt.Sprintf("%s-ori,UN,0,%s-ori,IF", serieTemp, serieTemp))

			if len(stack) == 0 {
				stack = append(stack, serieTemp)
			} else {
				stack = append(stack, serieTemp, "+")
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
	step := query.EndTime.Sub(query.StartTime) / time.Duration(query.Sample)

	data := rrd.XportResult{}

	data, err := xport.Xport(query.StartTime, query.EndTime, step)
	if err != nil {
		return nil, err
	}

	for index, itemName := range data.Legends {
		plotResult := &types.PlotResult{
			Name: itemName,
			Info: make(map[string]types.PlotValue),
		}

		// FIXME: skip last garbage entry (see https://github.com/ziutek/rrd/pull/13)
		for i := 0; i < data.RowCnt-1; i++ {
			plotResult.Plots = append(plotResult.Plots, types.PlotValue(data.ValueAt(index, i)))
		}

		result = append(result, plotResult)
	}

	data.FreeValues()

	return result, nil
}

// Refresh triggers a full connector data update.
func (connector *RRDConnector) Refresh(originName string, outputChan chan *catalog.CatalogRecord) error {
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

				connector.metrics[sourceName][metricFullName] = &rrdMetric{Dataset: dsName, FilePath: filePath}

				outputChan <- &catalog.CatalogRecord{
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
