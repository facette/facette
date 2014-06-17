package server

const (
	_ = iota
	eventInit
	eventRun
	eventCatalogRefresh
	eventShutdown

	_ = iota
	jobSignalRefresh
	jobSignalShutdown
)
