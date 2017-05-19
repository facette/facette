package catalog

// Metric represents a catalog metric instance.
type Metric struct {
	Name         string
	OriginalName string
	source       *Source
	connector    interface{}
}

// Source returns the parent source from the catalog metric.
func (m *Metric) Source() *Source {
	m.source.origin.catalog.RLock()
	defer m.source.origin.catalog.RUnlock()

	return m.source
}

// Connector returns the connector handler associated to the catalog metric.
func (m *Metric) Connector() interface{} {
	m.source.origin.catalog.RLock()
	defer m.source.origin.catalog.RUnlock()

	return m.connector
}

type metricList []*Metric

func (l metricList) Len() int {
	return len(l)
}

func (l metricList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l metricList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
