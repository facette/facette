// +build !disable_connector_rrd

package connector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"facette.io/facette/catalog"
	"facette.io/facette/series"
	"facette.io/facette/set"
	"facette.io/logger"
	"facette.io/maputil"
	"github.com/ziutek/rrd"
)

func init() {
	connectors["rrd"] = func(name string, settings *maputil.Map, logger *logger.Logger) (Connector, error) {
		var (
			pattern string
			err     error
		)

		c := &rrdConnector{
			name:    name,
			metrics: make(map[string]map[string]*rrdMetric),
			logger:  logger,
		}

		// Get connector handler settings
		c.path, err = settings.GetString("path", ".")
		if err != nil {
			return nil, err
		}
		c.path = strings.TrimRight(c.path, "/")

		c.daemon, err = settings.GetString("daemon", "")
		if err != nil {
			return nil, err
		}

		pattern, err = settings.GetString("pattern", "")
		if err != nil {
			return nil, err
		} else if pattern == "" {
			return nil, ErrMissingConnectorSetting("pattern")
		}

		// Check and compile regexp pattern
		c.pattern, err = compilePattern(pattern)
		if err != nil {
			return nil, err
		}

		return c, nil
	}
}

// rrdConnector represents a RRD connector instance.
type rrdConnector struct {
	name    string
	path    string
	daemon  string
	pattern *regexp.Regexp
	metrics map[string]map[string]*rrdMetric
	logger  *logger.Logger
}

// Name returns the name of the current connector.
func (c *rrdConnector) Name() string {
	return c.name
}

// Refresh triggers the connector data refresh.
func (c *rrdConnector) Refresh(output chan<- *catalog.Record) error {
	// Search for files and parse their path for source/metric pairs
	walkFunc := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			c.logger.Error("%s", err)
			return nil
		}

		// Skip non-files
		mode := fi.Mode() & os.ModeType
		if mode != 0 {
			return nil
		}

		// Get matching pattern elements
		m, err := matchPattern(c.pattern, strings.TrimPrefix(path, c.path+"/"))
		if err != nil {
			c.logger.Error("%s", err)
			return nil
		}

		source, metric := m[0], m[1]

		if _, ok := c.metrics[source]; !ok {
			c.metrics[source] = make(map[string]*rrdMetric)
		}

		// Extract information from .rrd file
		info, err := rrd.Info(path)
		if err != nil {
			c.logger.Error("failed to extract info: %s", err)
			return nil
		}

		// Extract consolidation functions list
		cfs := set.New()
		if cf, ok := info["rra.cf"].([]interface{}); ok {
			for _, entry := range cf {
				if v, ok := entry.(string); ok {
					cfs.Add(v)
				}
			}
		}

		// Parse RRD information for indexes
		indexes, ok := info["ds.index"].(map[string]interface{})
		if !ok {
			return nil
		}

		for ds := range indexes {
			for _, cf := range set.StringSlice(cfs) {
				metric = metric + "/" + ds + "/" + strings.ToLower(cf)

				c.metrics[source][metric] = &rrdMetric{
					DS:   ds,
					Path: path,
					Step: time.Duration(info["step"].(uint)) * time.Second,
					CF:   cf,
				}

				output <- &catalog.Record{
					Origin:    c.name,
					Source:    source,
					Metric:    metric,
					Connector: c,
				}
			}
		}

		return nil
	}

	return c.walk(c.path, "", walkFunc)
}

// Points retrieves the time series data according to the query parameters and a time interval.
func (c *rrdConnector) Points(q *series.Query) ([]series.Series, error) {
	var step time.Duration

	if len(q.Series) == 0 {
		return nil, series.ErrEmptySeries
	}

	// Initialize new RRD exporter
	xport := rrd.NewExporter()
	if c.daemon != "" {
		xport.SetDaemon(c.daemon)
	}

	// Prepare RRD definitions
	for i, s := range q.Series {
		if _, ok := c.metrics[s.Source]; !ok {
			return nil, ErrUnknownSource
		} else if _, ok := c.metrics[s.Source][s.Metric]; !ok {
			return nil, ErrUnknownMetric
		}

		name := fmt.Sprintf("series%d", i)
		path := strings.Replace(c.metrics[s.Source][s.Metric].Path, ":", "\\:", -1)

		xport.Def(name+"_def", path, c.metrics[s.Source][s.Metric].DS, c.metrics[s.Source][s.Metric].CF)
		xport.CDef(name+"_cdef", name+"_def")
		xport.XportDef(name+"_cdef", name)

		// Only keep the highest step
		if c.metrics[s.Source][s.Metric].Step > step {
			step = c.metrics[s.Source][s.Metric].Step
		}
	}

	// Set fallback step if none found
	if step == 0 {
		step = q.EndTime.Sub(q.StartTime) / time.Duration(series.DefaultSample)
	}

	// Retrieve data points
	data, err := xport.Xport(q.StartTime, q.EndTime, step)
	if err != nil {
		return nil, err
	}

	result := []series.Series{}
	for idx := range data.Legends {
		s := series.Series{}

		// FIXME: skip last garbage entry (see https://github.com/ziutek/rrd/pull/13)
		for i, n := 0, data.RowCnt-1; i < n; i++ {
			s.Points = append(s.Points, series.Point{
				Time:  q.StartTime.Add(data.Step * time.Duration(i)),
				Value: series.Value(data.ValueAt(idx, i)),
			})
		}

		result = append(result, s)
	}

	data.FreeValues()

	return result, nil
}

func (c *rrdConnector) walk(root, originalRoot string, walkFunc filepath.WalkFunc) error {
	if _, err := os.Stat(root); err != nil {
		c.logger.Error("%s", err)
		return nil
	}

	// Walk root directory
	return filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		var realPath string

		if err != nil {
			c.logger.Error("%s", err)
			return nil
		}

		mode := fi.Mode() & os.ModeType
		if mode == os.ModeSymlink {
			// Follow symbolic link if evaluation succeeds
			realPath, err = filepath.EvalSymlinks(path)
			if err != nil {
				c.logger.Error("%s", err)
				return nil
			}

			return c.walk(realPath, path, walkFunc)
		}

		if originalRoot != "" {
			path = originalRoot + strings.TrimPrefix(path, root)
		}

		return walkFunc(path, fi, err)
	})
}

type rrdMetric struct {
	DS   string
	Path string
	Step time.Duration
	CF   string
}
