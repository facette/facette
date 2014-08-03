package server

import (
	"time"

	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/library"
	"github.com/facette/facette/pkg/plot"
)

// ExpandRequest represents an expand request structure in the server backend.
type ExpandRequest [][3]string

func (e ExpandRequest) Len() int {
	return len(e)
}

func (e ExpandRequest) Less(i, j int) bool {
	return e[i][0]+e[i][1]+e[i][2] < e[j][0]+e[j][1]+e[j][2]
}

func (e ExpandRequest) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// PlotRequest represents a plot request structure in the server backend.
type PlotRequest struct {
	Time        time.Time      `json:"time"`
	Range       string         `json:"range"`
	Sample      int            `json:"sample"`
	Constants   []float64      `json:"constants"`
	Percentiles []float64      `json:"percentiles"`
	ID          string         `json:"id"`
	Graph       *library.Graph `json:"graph"`
	startTime   time.Time
	endTime     time.Time
}

// OriginResponse represents an origin response structure in the server backend.
type OriginResponse struct {
	Name      string `json:"name"`
	Connector string `json:"connector"`
}

// SourceResponse represents a source response structure in the server backend.
type SourceResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
}

// MetricResponse represents a metric response structure in the server backend.
type MetricResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Sources []string `json:"sources"`
}

// StringListResponse represents a list of strings response structure in the server backend.
type StringListResponse []string

func (r StringListResponse) Len() int {
	return len(r)
}

func (r StringListResponse) Less(i, j int) bool {
	return r[i] < r[j]
}

func (r StringListResponse) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r StringListResponse) slice(i, j int) interface{} {
	return r[i:j]
}

// ItemResponse represents an item response structure in the server backend.
type ItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Modified    string `json:"modified"`
}

// ItemListResponse represents a list of items response structure in the server backend.
type ItemListResponse []*ItemResponse

func (r ItemListResponse) Len() int {
	return len(r)
}

func (r ItemListResponse) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

func (r ItemListResponse) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ItemListResponse) slice(i, j int) interface{} {
	return r[i:j]
}

// CollectionResponse represents a collection response structure in the server backend.
type CollectionResponse struct {
	ItemResponse
	Parent      *string `json:"parent"`
	HasChildren bool    `json:"has_children"`
}

// CollectionListResponse represents a list of collections response structure in the backend server.
type CollectionListResponse []*CollectionResponse

func (r CollectionListResponse) Len() int {
	return len(r)
}

func (r CollectionListResponse) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

func (r CollectionListResponse) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r CollectionListResponse) slice(i, j int) interface{} {
	return r[i:j]
}

// ScaleValueResponse represents a scale value response structure in the server backend.
type ScaleValueResponse struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// ScaleValueListResponse represents a list of scale values structure in the backend server.
type ScaleValueListResponse []*ScaleValueResponse

func (r ScaleValueListResponse) Len() int {
	return len(r)
}

func (r ScaleValueListResponse) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

func (r ScaleValueListResponse) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ScaleValueListResponse) slice(i, j int) interface{} {
	return r[i:j]
}

// UnitValueResponse represents an unit value response structure in the server backend.
type UnitValueResponse struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// UnitValueListResponse represents a list of unit values structure in the backend server.
type UnitValueListResponse []*UnitValueResponse

func (r UnitValueListResponse) Len() int {
	return len(r)
}

func (r UnitValueListResponse) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

func (r UnitValueListResponse) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r UnitValueListResponse) slice(i, j int) interface{} {
	return r[i:j]
}

// PlotResponse represents a plot response structure in the server backend.
type PlotResponse struct {
	ID          string            `json:"id"`
	Start       string            `json:"start"`
	End         string            `json:"end"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        int               `json:"type"`
	StackMode   int               `json:"stack_mode"`
	UnitType    int               `json:"unit_type"`
	UnitLegend  string            `json:"unit_legend"`
	Series      []*SeriesResponse `json:"series"`
	Modified    time.Time         `json:"modified"`
}

// SeriesResponse represents a series response structure in the server backend.
type SeriesResponse struct {
	Name    string                 `json:"name"`
	StackID int                    `json:"stack_id"`
	Plots   []plot.Plot            `json:"plots"`
	Summary map[string]plot.Value  `json:"summary"`
	Options map[string]interface{} `json:"options"`
}

// Unexported types
type listResponse struct {
	list   sortableListResponse
	offset int
	limit  int
}

type sortableListResponse interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
	slice(offset, limit int) interface{}
}

type serverResponse struct {
	Message string `json:"message"`
}

type statsResponse struct {
	Origins      int `json:"origins"`
	Sources      int `json:"sources"`
	Metrics      int `json:"metrics"`
	Graphs       int `json:"graphs"`
	Collections  int `json:"collections"`
	SourceGroups int `json:"sourcegroups"`
	MetricGroups int `json:"metricgroups"`
}

type providerQuery struct {
	query     plot.Query
	queryMap  []providerQueryMap
	connector connector.Connector
}

type providerQueryMap struct {
	seriesName string
	sourceName string
	metricName string
}
