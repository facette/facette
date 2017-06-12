// +build !disable_connector_rrd

package connector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"facette/catalog"
	"facette/series"

	"github.com/facette/logger"
	"github.com/facette/maputil"
	"github.com/fatih/set"
	"github.com/ziutek/rrd"
)

type rrdMetric struct {
	ds   string
	path string
	step time.Duration
	cf   string
}

// rrdConnector implements the connector handler for RRD files.
type rrdConnector struct {
	name    string
	path    string
	daemon  string
	pattern *regexp.Regexp
	metrics map[string]map[string]*rrdMetric
	log     *logger.Logger
}

func init() {
	connectors["rrd"] = func(name string, settings *maputil.Map, log *logger.Logger) (Connector, error) {
		var err error

		c := &rrdConnector{
			name:    name,
			metrics: make(map[string]map[string]*rrdMetric),
			log:     log,
		}

		// Get connector handler settings
		if c.path, err = settings.GetString("path", "."); err != nil {
			return nil, err
		}
		c.path = strings.TrimRight(c.path, "/")

		if c.daemon, err = settings.GetString("daemon", ""); err != nil {
			return nil, err
		}

		pattern, err := settings.GetString("pattern", "")
		if err != nil {
			return nil, err
		} else if pattern == "" {
			return nil, ErrMissingConnectorSetting("pattern")
		}

		// Check and compile regexp pattern
		if c.pattern, err = compilePattern(pattern); err != nil {
			return nil, fmt.Errorf("unable to compile pattern: %s", err)
		}

		return c, nil
	}
}

// Name returns the name of the current connector.
func (c *rrdConnector) Name() string {
	return c.name
}

// Refresh triggers the connector data refresh.
func (c *rrdConnector) Refresh(output chan<- *catalog.Record) error {
	// Search for files and parse their path for source/metric pairs
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.log.Error("%s", err)
			return nil
		}

		// Skip non-files
		mode := info.Mode() & os.ModeType
		if mode != 0 {
			return nil
		}

		// Get matching pattern elements
		m, err := matchPattern(c.pattern, strings.TrimPrefix(path, c.path+"/"))
		if err != nil {
			c.log.Error("%s", err)
			return nil
		}

		source, metric := m[0], m[1]

		if _, ok := c.metrics[source]; !ok {
			c.metrics[source] = make(map[string]*rrdMetric)
		}

		// Extract information from .rrd file
		rinfo, err := rrd.Info(path)
		if err != nil {
			c.log.Error("failed to extract info: %s", err)
			return nil
		}

		// Extract consolidation functions list
		cfs := set.New()
		if cf, ok := rinfo["rra.cf"].([]interface{}); ok {
			for _, entry := range cf {
				if name, ok := entry.(string); ok {
					cfs.Add(name)
				}
			}
		}

		// Parse RRD information for indexes
		indexes, ok := rinfo["ds.index"].(map[string]interface{})
		if !ok {
			return nil
		}

		for ds := range indexes {
			for _, cf := range set.StringSlice(cfs) {
				metric := metric + "/" + ds + "/" + strings.ToLower(cf)

				c.metrics[source][metric] = &rrdMetric{
					ds:   ds,
					path: path,
					step: time.Duration(rinfo["step"].(uint)) * time.Second,
					cf:   cf,
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
		path := strings.Replace(c.metrics[s.Source][s.Metric].path, ":", "\\:", -1)

		xport.Def(name+"_def", path, c.metrics[s.Source][s.Metric].ds, c.metrics[s.Source][s.Metric].cf)
		xport.CDef(name+"_cdef", name+"_def")
		xport.XportDef(name+"_cdef", name)

		// Only keep the highest step
		if c.metrics[s.Source][s.Metric].step > step {
			step = c.metrics[s.Source][s.Metric].step
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
		c.log.Error("%s", err)
		return nil
	}

	// Walk root directory
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.log.Error("%s", err)
			return nil
		}

		mode := info.Mode() & os.ModeType
		if mode == os.ModeSymlink {
			// Follow symbolic link if evaluation succeeds
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				c.log.Error("%s", err)
				return nil
			}

			return c.walk(realPath, path, walkFunc)
		}

		if originalRoot != "" {
			path = originalRoot + strings.TrimPrefix(path, root)
		}

		return walkFunc(path, info, err)
	})
}
