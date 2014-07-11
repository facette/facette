// +build rrd

package connector

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

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
func (connector *RRDConnector) GetPlots(query *types.PlotQuery) (map[string]*types.PlotResult, error) {
	var xport *rrd.Exporter

	if len(query.Group.Series) == 0 {
		return nil, fmt.Errorf("group has no series")
	} else if query.Group.Type != OperGroupTypeNone && len(query.Group.Series) == 1 {
		query.Group.Type = OperGroupTypeNone
	}

	result := make(map[string]*types.PlotResult)
	series := make(map[string]string)

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

			serieTemp := fmt.Sprintf("serie%d", count)
			serieName := serie.Name

			count += 1

			graph.Def(
				serieTemp+"-orig0",
				connector.metrics[serie.Metric.Source][serie.Metric.Name].FilePath,
				connector.metrics[serie.Metric.Source][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			if serie.Scale != 0 {
				graph.CDef(serieTemp+"-orig1", fmt.Sprintf("%s-orig0,%g,*", serieTemp, serie.Scale))
			} else {
				graph.CDef(serieTemp+"-orig1", serieTemp+"-orig0")
			}

			if query.Group.Scale != 0 {
				graph.CDef(serieTemp, fmt.Sprintf("%s-orig1,%g,*", serieTemp, query.Group.Scale))
			} else {
				graph.CDef(serieTemp, serieTemp+"-orig1")
			}

			// Set graph information request
			rrdSetGraph(graph, serieTemp, serieName, query.Percentiles)

			// Set plots request
			xport.Def(
				serieTemp+"-orig0",
				connector.metrics[serie.Metric.Source][serie.Metric.Name].FilePath,
				connector.metrics[serie.Metric.Source][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			if serie.Scale != 0 {
				xport.CDef(serieTemp+"-orig1", fmt.Sprintf("%s-orig0,%g,*", serieTemp, serie.Scale))
			} else {
				xport.CDef(serieTemp+"-orig1", serieTemp+"-orig0")
			}

			if query.Group.Scale != 0 {
				xport.CDef(serieTemp, fmt.Sprintf("%s-orig1,%g,*", serieTemp, query.Group.Scale))
			} else {
				xport.CDef(serieTemp, serieTemp+"-orig1")
			}

			xport.XportDef(serieTemp, serieTemp)

			// Set serie matching
			series[serieTemp] = serieName
		}

	case OperGroupTypeAvg, OperGroupTypeSum:
		serieName := fmt.Sprintf("serie%d", count)
		count += 1

		for index, serie := range query.Group.Series {
			if serie.Metric == nil {
				continue
			}

			serieTemp := serieName + fmt.Sprintf("-tmp%d", index)

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

		graph.CDef(serieName+"-orig", strings.Join(stack, ","))

		if query.Group.Scale != 0 {
			graph.CDef(serieName, fmt.Sprintf("%s-orig,%g,*", serieName, query.Group.Scale))
		} else {
			graph.CDef(serieName, serieName+"-orig")
		}

		// Set graph information request
		rrdSetGraph(graph, serieName, query.Group.Name, query.Percentiles)

		// Set plots request
		xport.CDef(serieName+"-orig", strings.Join(stack, ","))

		if query.Group.Scale != 0 {
			xport.CDef(serieName, fmt.Sprintf("%s-orig,%g,*", serieName, query.Group.Scale))
		} else {
			xport.CDef(serieName, serieName+"-orig")
		}

		xport.XportDef(serieName, serieName)

		// Set serie matching
		series[serieName] = query.Group.Name

	default:
		return nil, fmt.Errorf("unknown operator type %d", query.Group.Type)
	}

	// Get plots
	data := rrd.XportResult{}

	data, err := xport.Xport(query.StartTime, query.EndTime, query.Step)
	if err != nil {
		return nil, err
	}

	for index, serieName := range data.Legends {
		result[series[serieName]] = &types.PlotResult{Info: make(map[string]types.PlotValue)}

		for i := 0; i < data.RowCnt; i++ {
			result[series[serieName]].Plots = append(result[series[serieName]].Plots,
				types.PlotValue(data.ValueAt(index, i)))
		}
	}

	// Parse graph information
	graphInfo, _, err := graph.Graph(query.StartTime, query.EndTime)
	if err != nil {
		return nil, err
	}

	rrdParseInfo(graphInfo, result)

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

func rrdParseInfo(info rrd.GraphInfo, data map[string]*types.PlotResult) {
	for _, value := range info.Print {
		chunks := strings.SplitN(value, ",", 3)

		chunkFloat, err := strconv.ParseFloat(chunks[2], 64)
		if err != nil {
			chunkFloat = math.NaN()
		}

		if data[chunks[0]] == nil {
			data[chunks[0]] = &types.PlotResult{Info: make(map[string]types.PlotValue)}
		}

		data[chunks[0]].Info[chunks[1]] = types.PlotValue(chunkFloat)
	}
}

func rrdSetGraph(graph *rrd.Grapher, serieName, itemName string, percentiles []float64) {
	graph.VDef(serieName+"-min", serieName+",MINIMUM")
	graph.Print(serieName+"-min", itemName+",min,%lf")

	graph.VDef(serieName+"-avg", serieName+",AVERAGE")
	graph.Print(serieName+"-avg", itemName+",avg,%lf")

	graph.VDef(serieName+"-max", serieName+",MAXIMUM")
	graph.Print(serieName+"-max", itemName+",max,%lf")

	graph.VDef(serieName+"-last", serieName+",LAST")
	graph.Print(serieName+"-last", itemName+",last,%lf")

	for index, percentile := range percentiles {
		graph.CDef(fmt.Sprintf("%s-cdef%d", serieName, index),
			fmt.Sprintf("%s,UN,0,%s,IF", serieName, serieName))
		graph.VDef(fmt.Sprintf("%s-vdef%d", serieName, index),
			fmt.Sprintf("%s-cdef%d,%f,PERCENT", serieName, index, percentile))

		if percentile-float64(int(percentile)) != 0 {
			graph.Print(fmt.Sprintf("%s-vdef%d", serieName, index),
				fmt.Sprintf("%s,%.2fth,%%lf", itemName, percentile))
		} else {
			graph.Print(fmt.Sprintf("%s-vdef%d", serieName, index),
				fmt.Sprintf("%s,%.0fth,%%lf", itemName, percentile))
		}
	}
}
