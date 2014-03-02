package connector

import (
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	Path      string
	Pattern   string
	inputChan *chan [2]string
	metrics   map[string]map[string]*rrdMetric
}

// GetPlots calculates and returns plots data based on a time interval.
func (handler *RRDConnector) GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
	percentiles []float64) (map[string]*PlotResult, error) {

	return handler.rrdGetData(query, startTime, endTime, step, percentiles, false)
}

// Refresh triggers a full connector data update.
func (handler *RRDConnector) Refresh() error {
	// Compile pattern
	re := regexp.MustCompile(handler.Pattern)

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

		submatch := re.FindStringSubmatch(filePath[len(handler.Path)+1:])
		if len(submatch) == 0 {
			log.Printf("WARNING: file `%s' does not match pattern", filePath)
			return nil
		}

		if re.SubexpNames()[1] == "source" {
			sourceName = submatch[1]
			metricName = submatch[2]
		} else {
			sourceName = submatch[2]
			metricName = submatch[1]
		}

		if _, ok := handler.metrics[sourceName]; !ok {
			handler.metrics[sourceName] = make(map[string]*rrdMetric)
		}

		// Extract metric information from .rrd file
		info, err := rrd.Info(filePath)
		if err != nil {
			return err
		}

		if _, ok := info["ds.index"]; ok {
			for dsName := range info["ds.index"].(map[string]interface{}) {
				metricFullName := metricName + "/" + dsName

				*handler.inputChan <- [2]string{sourceName, metricFullName}
				handler.metrics[sourceName][metricFullName] = &rrdMetric{Dataset: dsName, FilePath: filePath}
			}
		}

		return err
	}

	err := utils.WalkDir(handler.Path, walkFunc)
	if err != nil {
		return err
	}

	// Close channel once updated
	close(*handler.inputChan)

	return nil
}

func (handler *RRDConnector) rrdGetData(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
	percentiles []float64, infoOnly bool) (map[string]*PlotResult, error) {

	var xport *rrd.Exporter

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("group has no series")
	} else if query.Type != OperGroupTypeNone && len(query.Series) == 1 {
		query.Type = OperGroupTypeNone
	}

	result := make(map[string]*PlotResult)
	series := make(map[string]string)

	stack := make([]string, 0)
	graph := rrd.NewGrapher()

	if !infoOnly {
		xport = rrd.NewExporter()
	}

	count := 0

	switch query.Type {
	case OperGroupTypeNone:
		for _, serie := range query.Series {
			if serie.Metric == nil {
				continue
			}

			serieTemp := fmt.Sprintf("serie%d", count)
			serieName := serie.Name

			count += 1

			graph.Def(
				serieTemp+"-orig0",
				handler.metrics[serie.Metric.SourceName][serie.Metric.Name].FilePath,
				handler.metrics[serie.Metric.SourceName][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			if serie.Scale != 0 {
				graph.CDef(serieTemp+"-orig1", fmt.Sprintf("%s-orig0,%f,*", serieTemp, serie.Scale))
			} else {
				graph.CDef(serieTemp+"-orig1", serieTemp+"-orig0")
			}

			if query.Scale != 0 {
				graph.CDef(serieTemp, fmt.Sprintf("%s-orig1,%f,*", serieTemp, query.Scale))
			} else {
				graph.CDef(serieTemp, serieTemp+"-orig1")
			}

			// Set graph information request
			rrdSetGraph(graph, serieTemp, serieName, percentiles)

			// Set plots request
			if !infoOnly {
				xport.Def(
					serieTemp+"-orig0",
					handler.metrics[serie.Metric.SourceName][serie.Metric.Name].FilePath,
					handler.metrics[serie.Metric.SourceName][serie.Metric.Name].Dataset,
					"AVERAGE",
				)

				if serie.Scale != 0 {
					xport.CDef(serieTemp+"-orig1", fmt.Sprintf("%s-orig0,%f,*", serieTemp, serie.Scale))
				} else {
					xport.CDef(serieTemp+"-orig1", serieTemp+"-orig0")
				}

				if query.Scale != 0 {
					xport.CDef(serieTemp, fmt.Sprintf("%s-orig1,%f,*", serieTemp, query.Scale))
				} else {
					xport.CDef(serieTemp, serieTemp+"-orig1")
				}

				xport.XportDef(serieTemp, serieTemp)
			}

			// Set serie matching
			series[serieTemp] = serieName
		}

		break

	case OperGroupTypeAvg, OperGroupTypeSum:
		serieName := fmt.Sprintf("serie%d", count)
		count += 1

		for index, serie := range query.Series {
			if serie.Metric == nil {
				continue
			}

			serieTemp := serieName + fmt.Sprintf("-tmp%d", index)

			graph.Def(
				serieTemp,
				handler.metrics[serie.Metric.SourceName][serie.Metric.Name].FilePath,
				handler.metrics[serie.Metric.SourceName][serie.Metric.Name].Dataset,
				"AVERAGE",
			)

			if !infoOnly {
				xport.Def(
					serieTemp,
					handler.metrics[serie.Metric.SourceName][serie.Metric.Name].FilePath,
					handler.metrics[serie.Metric.SourceName][serie.Metric.Name].Dataset,
					"AVERAGE",
				)
			}

			if len(stack) == 0 {
				stack = append(stack, serieTemp)
			} else {
				stack = append(stack, serieTemp, "+")
			}
		}

		if query.Type == OperGroupTypeAvg {
			stack = append(stack, strconv.Itoa(len(query.Series)), "/")
		}

		graph.CDef(serieName+"-orig", strings.Join(stack, ","))

		if query.Scale != 0 {
			graph.CDef(serieName, fmt.Sprintf("%s-orig,%f,*", serieName, query.Scale))
		} else {
			graph.CDef(serieName, serieName+"-orig")
		}

		// Set graph information request
		rrdSetGraph(graph, serieName, query.Name, percentiles)

		// Set plots request
		if !infoOnly {
			xport.CDef(serieName+"-orig", strings.Join(stack, ","))

			if query.Scale != 0 {
				xport.CDef(serieName, fmt.Sprintf("%s-orig,%f,*", serieName, query.Scale))
			} else {
				xport.CDef(serieName, serieName+"-orig")
			}

			xport.XportDef(serieName, serieName)
		}

		// Set serie matching
		series[serieName] = query.Name

		break

	default:
		return nil, fmt.Errorf("unknown `%d' operator type", query.Type)
	}

	// Get plots
	data := rrd.XportResult{}

	if !infoOnly {
		data, err := xport.Xport(startTime, endTime, step)
		if err != nil {
			return nil, err
		}

		for index, serieName := range data.Legends {
			result[series[serieName]] = &PlotResult{Info: make(map[string]types.PlotValue)}

			for i := 0; i < data.RowCnt; i++ {
				result[series[serieName]].Plots = append(result[series[serieName]].Plots,
					types.PlotValue(data.ValueAt(index, i)))
			}
		}
	}

	// Parse graph information
	graphInfo, _, err := graph.Graph(startTime, endTime)
	if err != nil {
		return nil, err
	}

	rrdParseInfo(graphInfo, result)

	data.FreeValues()

	return result, nil
}

func rrdParseInfo(info rrd.GraphInfo, data map[string]*PlotResult) {
	for _, value := range info.Print {
		chunks := strings.SplitN(value, ",", 3)

		chunkFloat, err := strconv.ParseFloat(chunks[2], 64)
		if err != nil {
			chunkFloat = math.NaN()
		}

		if data[chunks[0]] == nil {
			data[chunks[0]] = &PlotResult{Info: make(map[string]types.PlotValue)}
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

func init() {
	Connectors["rrd"] = func(inputChan *chan [2]string, config map[string]string) (interface{}, error) {
		if _, ok := config["path"]; !ok {
			return nil, fmt.Errorf("missing `path' mandatory connector setting")
		} else if _, ok := config["pattern"]; !ok {
			return nil, fmt.Errorf("missing `pattern' mandatory connector setting")
		}

		return &RRDConnector{
			Path:      config["path"],
			Pattern:   config["pattern"],
			inputChan: inputChan,
			metrics:   make(map[string]map[string]*rrdMetric),
		}, nil
	}
}
