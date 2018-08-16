// +build !disable_connector_rrd

package connector

import (
	"encoding/gob"
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
	"github.com/pkg/errors"
	"github.com/ziutek/rrd"
)

func init() {
	connectors["rrd"] = func(name string, settings *maputil.Map, logger *logger.Logger) (Connector, error) {
		var (
			pattern string
			err     error
		)

		c := &rrdConnector{
			name:   name,
			logger: logger,
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

	// Register type for catalog dump
	gob.Register(time.Duration(0))
}

// rrdConnector represents a RRD connector instance.
type rrdConnector struct {
	name    string
	path    string
	daemon  string
	pattern *regexp.Regexp
	logger  *logger.Logger
}

func (c *rrdConnector) Name() string {
	return c.name
}

func (c *rrdConnector) Points(q *series.Query) ([]series.Series, error) {
	var stepMax time.Duration

	if len(q.Metrics) == 0 {
		return nil, fmt.Errorf("requested metrics list is empty")
	}

	// Initialize new RRD exporter
	xport := rrd.NewExporter()
	if c.daemon != "" {
		xport.SetDaemon(c.daemon)
	}

	// Prepare RRD definitions
	for idx, m := range q.Metrics {
		var step time.Duration

		path, err := m.Attributes.GetString("path", "")
		if err != nil {
			return nil, errors.Wrap(ErrInvalidAttribute, "path")
		}
		path = strings.Replace(path, ":", "\\:", -1)

		ds, err := m.Attributes.GetString("ds", "")
		if err != nil {
			return nil, errors.Wrap(ErrInvalidAttribute, "ds")
		}

		cf, err := m.Attributes.GetString("cf", "")
		if err != nil {
			return nil, errors.Wrap(ErrInvalidAttribute, "cf")
		}

		if v, err := m.Attributes.GetInterface("step", nil); err != nil {
			return nil, errors.Wrap(ErrInvalidAttribute, "step")
		} else if v, ok := v.(time.Duration); !ok {
			return nil, errors.Wrap(ErrInvalidAttribute, "step")
		} else {
			step = v
		}

		name := fmt.Sprintf("series%d", idx)

		xport.Def(name+"_def", path, ds, cf)
		xport.CDef(name+"_cdef", name+"_def")
		xport.XportDef(name+"_cdef", name)

		// Only keep the highest step
		if step > stepMax {
			stepMax = step
		}
	}

	// Set fallback step if none found
	if stepMax == 0 {
		stepMax = q.EndTime.Sub(q.StartTime) / time.Duration(series.DefaultSample)
	}

	// Retrieve data points
	data, err := xport.Xport(q.StartTime, q.EndTime, stepMax)
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
				output <- &catalog.Record{
					Origin: c.name,
					Source: source,
					Metric: metric + "/" + ds + "/" + strings.ToLower(cf),
					Attributes: &maputil.Map{
						"path": path,
						"ds":   ds,
						"cf":   cf,
						"step": time.Duration(info["step"].(uint)) * time.Second,
					},
				}
			}
		}

		return nil
	}

	return c.walk(c.path, "", walkFunc)
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
