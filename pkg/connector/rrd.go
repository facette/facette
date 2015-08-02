// +build rrd

package connector

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/plot"
	"github.com/facette/facette/pkg/utils"
	"github.com/fatih/set"
	"github.com/ziutek/rrd"
)

type rrdMetric struct {
	Dataset  string
	FilePath string
	Step     time.Duration
	Cf       string
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

		c := &RRDConnector{
			name:    name,
			metrics: make(map[string]map[string]*rrdMetric),
		}

		if c.path, err = config.GetString(settings, "path", true); err != nil {
			return nil, err
		}
		c.path = strings.TrimRight(c.path, "/")
		if c.path == "" {
			c.path = "."
		}

		if c.daemon, err = config.GetString(settings, "daemon", false); err != nil {
			return nil, err
		}

		if pattern, err = config.GetString(settings, "pattern", true); err != nil {
			return nil, err
		}

		// Check and compile regexp pattern
		if c.re, err = compilePattern(pattern); err != nil {
			return nil, fmt.Errorf("unable to compile regexp pattern: %s", err)
		}

		return c, nil
	}
}

// GetName returns the name of the current connector.
func (c *RRDConnector) GetName() string {
	return c.name
}

// GetPlots retrieves time series data from origin based on a query and a time interval.
func (c *RRDConnector) GetPlots(query *plot.Query) ([]*plot.Series, error) {
	var (
		results []*plot.Series
		xport   *rrd.Exporter
	)

	if len(query.Series) == 0 {
		return nil, fmt.Errorf("rrd[%s]: requested series list is empty", c.name)
	}

	xport = rrd.NewExporter()

	if c.daemon != "" {
		xport.SetDaemon(c.daemon)
	}

	step := time.Duration(0)

	for _, s := range query.Series {
		filePath := strings.Replace(c.metrics[s.Source][s.Metric].FilePath, ":", "\\:", -1)

		// Set plots request
		xport.Def(s.Name+"-def0", filePath, c.metrics[s.Source][s.Metric].Dataset, c.metrics[s.Source][s.Metric].Cf)
		xport.CDef(s.Name, s.Name+"-def0")
		xport.XportDef(s.Name, s.Name)

		if c.metrics[s.Source][s.Metric].Step > step {
			step = c.metrics[s.Source][s.Metric].Step
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

	for idx, name := range data.Legends {
		series := &plot.Series{
			Name: name,
		}

		// FIXME: skip last garbage entry (see https://github.com/ziutek/rrd/pull/13)
		for i := 0; i < data.RowCnt-1; i++ {
			series.Plots = append(series.Plots, plot.Plot{
				Time:  query.StartTime.Add(data.Step * time.Duration(i)),
				Value: plot.Value(data.ValueAt(idx, i)),
			})
		}

		results = append(results, series)
	}

	data.FreeValues()

	return results, nil
}

// Refresh triggers a full connector data update.
func (c *RRDConnector) Refresh(originName string, outputChan chan<- *catalog.Record) error {
	// Search for files and parse their path for source/metric pairs
	walkFunc := func(filePath string, fileInfo os.FileInfo, err error) error {
		var sourceName, metricName string

		// Report errors
		if err != nil {
			logger.Log(logger.LevelWarning, "connector", "rrd[%s]: error while walking: %s", c.name, err)
			return nil
		}

		// Skip non-files
		mode := fileInfo.Mode() & os.ModeType
		if mode != 0 {
			return nil
		}

		// Get pattern matches
		m, err := matchSeriesPattern(c.re, strings.TrimPrefix(filePath, c.path+"/"))
		if err != nil {
			logger.Log(logger.LevelInfo, "connector", "rrd[%s]: file `%s' does not match pattern, ignoring", c.name,
				filePath)
			return nil
		}

		sourceName, metricName = m[0], m[1]

		if _, ok := c.metrics[sourceName]; !ok {
			c.metrics[sourceName] = make(map[string]*rrdMetric)
		}

		// Extract metric information from .rrd file
		info, err := rrd.Info(filePath)
		if err != nil {
			logger.Log(logger.LevelWarning, "connector", "rrd[%s]: %s", c.name, err)
			return nil
		}

		// Extract consolidation functions list
		cfSet := set.New(set.ThreadSafe)

		if cf, ok := info["rra.cf"].([]interface{}); ok {
			for _, entry := range cf {
				if name, ok := entry.(string); ok {
					cfSet.Add(name)
				}
			}
		}

		cfList := set.StringSlice(cfSet)

		if _, ok := info["ds.index"]; ok {
			indexes, ok := info["ds.index"].(map[string]interface{})
			if !ok {
				return nil
			}

			for dsName := range indexes {
				for _, cfName := range cfList {
					metricFullName := metricName + "/" + dsName + "/" + strings.ToLower(cfName)

					c.metrics[sourceName][metricFullName] = &rrdMetric{
						Dataset:  dsName,
						FilePath: filePath,
						Step:     time.Duration(info["step"].(uint)) * time.Second,
						Cf:       cfName,
					}

					outputChan <- &catalog.Record{
						Origin:    originName,
						Source:    sourceName,
						Metric:    metricFullName,
						Connector: c,
					}
				}
			}
		}

		return nil
	}

	if err := utils.WalkDir(c.path, walkFunc); err != nil {
		return err
	}

	return nil
}
