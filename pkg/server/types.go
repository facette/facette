package server

import (
	"time"

	"github.com/facette/facette/pkg/types"
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
	Time        string    `json:"time"`
	Range       string    `json:"range"`
	Sample      int       `json:"sample"`
	Constants   []float64 `json:"constants"`
	Percentiles []float64 `json:"percentiles"`
	Graph       string    `json:"graph"`
	Origin      string    `json:"origin"`
	Source      string    `json:"source"`
	Metric      string    `json:"metric"`
	Template    string    `json:"template"`
	Filter      string    `json:"filter"`
}

// OriginResponse represents an origin response structure in the server backend.
type OriginResponse struct {
	Name      string `json:"name"`
	Connector string `json:"connector"`
	Updated   string `json:"updated"`
}

// SourceResponse represents a source response structure in the server backend.
type SourceResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Updated string   `json:"updated"`
}

// MetricResponse represents a metric response structure in the server backend.
type MetricResponse struct {
	Name    string   `json:"name"`
	Origins []string `json:"origins"`
	Sources []string `json:"sources"`
	Updated string   `json:"updated"`
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

func (r StringListResponse) slice(offset, limit int) interface{} {
	return r[offset : offset+limit]
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

func (r ItemListResponse) slice(offset, limit int) interface{} {
	return r[offset : offset+limit]
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

func (r CollectionListResponse) slice(offset, limit int) interface{} {
	return r[offset : offset+limit]
}

// PlotResponse represents a plot response structure in the server backend.
type PlotResponse struct {
	ID          string           `json:"id"`
	Start       string           `json:"start"`
	End         string           `json:"end"`
	Step        float64          `json:"step"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        int              `json:"type"`
	StackMode   int              `json:"stack_mode"`
	Stacks      []*StackResponse `json:"stacks"`
	Modified    time.Time        `json:"modified"`
}

// StackResponse represents a stack response structure in the server backend.
type StackResponse struct {
	Name   string           `json:"name"`
	Series []*SerieResponse `json:"series"`
}

// SerieResponse represents a serie response structure in the server backend.
type SerieResponse struct {
	Name    string                     `json:"name"`
	Plots   []types.PlotValue          `json:"plots"`
	Info    map[string]types.PlotValue `json:"info"`
	Options map[string]interface{}     `json:"options"`
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
	Origins        int    `json:"origins"`
	Sources        int    `json:"sources"`
	Metrics        int    `json:"metrics"`
	CatalogUpdated string `json:"catalog_updated"`

	Graphs      int `json:"graphs"`
	Collections int `json:"collections"`
	Groups      int `json:"groups"`
}

type resourceResponse struct {
	Scales [][2]interface{} `json:"scales"`
}
