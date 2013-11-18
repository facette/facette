package backend

import (
	"facette/common"
	"facette/utils"
	"fmt"
	"github.com/ziutek/rrd"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RRDBackendHandler represents the main structure of the RRD backend.
type RRDBackendHandler struct {
	Path    string
	Pattern string
	origin  *Origin
}

// GetPlots calculates and returns plot data based on a time interval.
func (handler *RRDBackendHandler) GetPlots(query *GroupQuery, startTime, endTime time.Time, step time.Duration,
	percentiles []float64) (map[string]*PlotResult, error) {
	return rrdGetData(query, startTime, endTime, step, percentiles, false)
}

// GetValue calculates and returns plot data at a specific reference time.
func (handler *RRDBackendHandler) GetValue(query *GroupQuery, refTime time.Time,
	percentiles []float64) (map[string]map[string]common.PlotValue, error) {
	var (
		data   map[string]*PlotResult
		err    error
		result map[string]map[string]common.PlotValue
	)

	result = make(map[string]map[string]common.PlotValue)

	if data, err = rrdGetData(query, refTime.Add(-1*time.Minute), refTime, time.Minute, percentiles, true); err != nil {
		return nil, err
	}

	for serieName := range data {
		result[serieName] = data[serieName].Info
	}

	return result, err
}

// Update triggers a full backend data update.
func (handler *RRDBackendHandler) Update() error {
	var (
		err      error
		groups   map[string]bool
		re       *regexp.Regexp
		walkFunc func(filePath string, fileInfo os.FileInfo, err error) error
	)

	// Compile pattern
	re = regexp.MustCompile(handler.Pattern)

	// Validate pattern keywords
	groups = make(map[string]bool)

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
	walkFunc = func(filePath string, fileInfo os.FileInfo, err error) error {
		var (
			info       map[string]interface{}
			metricPath string
			mode       os.FileMode
			submatch   []string
		)

		if err != nil {
			// Stop if previous error
			return err
		} else if mode = fileInfo.Mode() & os.ModeType; mode != 0 {
			// Skip non-files
			return nil
		} else if submatch = re.FindStringSubmatch(filePath[len(handler.Path)+1:]); len(submatch) == 0 {
			log.Printf("WARNING: file `%s' does not match pattern", filePath)
			return nil
		}

		var source, metric string

		if re.SubexpNames()[1] == "source" {
			source = submatch[1]
			metric = submatch[2]
		} else {
			source = submatch[2]
			metric = submatch[1]
		}

		if _, ok := handler.origin.Sources[source]; !ok {
			handler.origin.AppendSource(source)
		}

		// Extract metric information from .rrd file
		if info, err = rrd.Info(filePath); err != nil {
			return err
		}

		if _, ok := info["ds.index"]; ok {
			for dsName := range info["ds.index"].(map[string]interface{}) {
				metricPath = metric + "/" + dsName

				for _, filter := range handler.origin.catalog.Config.Origins[handler.origin.Name].Filters {
					if !filter.PatternRegexp.MatchString(metricPath) {
						continue
					}

					if filter.Discard {
						if handler.origin.catalog.debugLevel > 2 {
							log.Printf("DEBUG: discarding `%s' metric...", metricPath)
						}

						return nil
					} else if filter.Rewrite != "" {
						metricPath = filter.PatternRegexp.ReplaceAllString(metricPath, filter.Rewrite)
					}
				}

				handler.origin.Sources[source].AppendMetric(metricPath, dsName, filePath)
			}
		}

		return err
	}

	if err = utils.WalkDir(handler.Path, walkFunc); err != nil {
		return err
	}

	return nil
}

func init() {
	BackendHandlers["rrd"] = NewRRDBackendHandler
}

func rrdGetData(query *GroupQuery, startTime, endTime time.Time, step time.Duration, percentiles []float64,
	infoOnly bool) (map[string]*PlotResult, error) {
	var (
		count     int
		data      rrd.XportResult
		err       error
		graph     *rrd.Grapher
		graphInfo rrd.GraphInfo
		i         int
		result    map[string]*PlotResult
		serieName string
		series    map[string]string
		serieTemp string
		stack     []string
		xport     *rrd.Exporter
	)

	result = make(map[string]*PlotResult)
	series = make(map[string]string)

	stack = nil
	graph = rrd.NewGrapher()

	if !infoOnly {
		xport = rrd.NewExporter()
	}

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("group has no series")
	} else if query.Type != OperGroupTypeNone && len(query.Series) == 1 {
		query.Type = OperGroupTypeNone
	}

	switch query.Type {
	case OperGroupTypeNone:
		for _, serie := range query.Series {
			if serie.Metric == nil {
				continue
			}

			serieName = fmt.Sprintf("serie%d", count)
			count += 1

			graph.Def(serieName, serie.Metric.FilePath, serie.Metric.Dataset, "AVERAGE")

			// Set graph information request
			rrdSetGraph(graph, serieName, serie.Name, percentiles)

			// Set plots request
			if !infoOnly {
				xport.Def(serieName, serie.Metric.FilePath, serie.Metric.Dataset, "AVERAGE")
				xport.XportDef(serieName, serieName)
			}

			// Set serie matching
			series[serieName] = serie.Name
		}

		break

	case OperGroupTypeAvg, OperGroupTypeSum:
		serieName = fmt.Sprintf("serie%d", count)
		count += 1

		for index, serie := range query.Series {
			if serie.Metric == nil {
				continue
			}

			serieTemp = serieName + fmt.Sprintf("-tmp%d", index)

			graph.Def(serieTemp, serie.Metric.FilePath, serie.Metric.Dataset, "AVERAGE")

			if !infoOnly {
				xport.Def(serieTemp, serie.Metric.FilePath, serie.Metric.Dataset, "AVERAGE")
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

		graph.CDef(serieName, strings.Join(stack, ","))

		// Set graph information request
		rrdSetGraph(graph, serieName, query.Name, percentiles)

		// Set plots request
		if !infoOnly {
			xport.CDef(serieName, strings.Join(stack, ","))
			xport.XportDef(serieName, serieName)
		}

		// Set serie matching
		series[serieName] = query.Name

		break

	default:
		return nil, fmt.Errorf("unknown `%d' operator type", query.Type)
	}

	// Get plots
	if !infoOnly {
		if data, err = xport.Xport(startTime, endTime, step); err != nil {
			return nil, err
		}

		for index, serieName := range data.Legends {
			result[series[serieName]] = &PlotResult{Info: make(map[string]common.PlotValue)}

			for i = 0; i < data.RowCnt; i++ {
				result[series[serieName]].Plots = append(result[series[serieName]].Plots,
					common.PlotValue(data.ValueAt(index, i)))
			}
		}
	}

	// Parse graph information
	if graphInfo, _, err = graph.Graph(startTime, endTime); err != nil {
		return nil, err
	}

	rrdParseInfo(graphInfo, result)

	data.FreeValues()

	return result, nil
}

func rrdParseInfo(info rrd.GraphInfo, data map[string]*PlotResult) {
	var (
		chunks     []string
		chunkFloat float64
		err        error
	)

	for _, value := range info.Print {
		chunks = strings.SplitN(value, ",", 3)

		if chunkFloat, err = strconv.ParseFloat(chunks[2], 64); err != nil {
			chunkFloat = math.NaN()
		}

		if data[chunks[0]] == nil {
			data[chunks[0]] = &PlotResult{Info: make(map[string]common.PlotValue)}
		}

		data[chunks[0]].Info[chunks[1]] = common.PlotValue(chunkFloat)
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

// NewRRDBackendHandler creates a new instance of BackendHandler.
func NewRRDBackendHandler(origin *Origin, config map[string]string) error {
	if _, ok := config["path"]; !ok {
		return fmt.Errorf("missing `path' mandatory backend definition")
	} else if _, ok := config["pattern"]; !ok {
		return fmt.Errorf("missing `pattern' mandatory backend definition")
	}

	origin.Backend = &RRDBackendHandler{
		Path:    config["path"],
		Pattern: config["pattern"],
		origin:  origin,
	}

	return nil
}
