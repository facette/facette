// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package api

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

	"facette.io/facette/pkg/connector"
	"facette.io/facette/pkg/errors"
	"facette.io/facette/pkg/filter"
	"facette.io/facette/pkg/series"
	"facette.io/facette/pkg/template"
)

var nameRegexp = regexp.MustCompile(`(?i)^[a-z0-9](?:[a-z0-9\-_]*[a-z0-9])?$`)

// Object is an API object.
type Object interface {
	Excerpt() interface{}
	GetMeta() ObjectMeta
	SetMeta(meta ObjectMeta)
	Validate() error

	object()
}

func (Chart) object()     {}
func (Dashboard) object() {}
func (Provider) object()  {}

// ObjectList is a list of API objects.
type ObjectList interface {
	Objects() []Object
	Len() int

	objectlist()
}

func (ChartList) objectlist()     {}
func (DashboardList) objectlist() {}
func (ProviderList) objectlist()  {}

// ObjectMeta are API object metadata.
type ObjectMeta struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

// Validate checks for API object metadata validity.
func (m ObjectMeta) Validate() error {
	if m.Name == "" {
		return errors.Wrap(ErrInvalid, "missing field: name")
	} else if !nameRegexp.MatchString(m.Name) {
		return errors.Wrapf(ErrInvalid, "invalid name: %s", m.Name)
	}

	if m.ID != "" {
		_, err := uuid.Parse(m.ID)
		if err != nil {
			return errors.Wrapf(ErrInvalid, "invalid identifier: %s", m.ID)
		}
	}

	return nil
}

// Chart is a chart API object.
type Chart struct {
	ObjectMeta `json:",inline"`
	Options    *ChartOptions `json:"options,omitempty"`
	Series     SeriesList    `json:"series,omitempty"`
	Link       string        `json:"link,omitempty"`
	Template   bool          `json:"template"`
}

// Excerpt returns an excerpted version of the chart API object.
func (c Chart) Excerpt() interface{} {
	options := map[string]interface{}{}
	if c.Options.Title != "" {
		options["title"] = c.Options.Title
	}

	return struct {
		ObjectMeta `json:",inline"`
		Options    map[string]interface{} `json:"options,omitempty"`
		Link       string                 `json:"link,omitempty"`
		Template   bool                   `json:"enabled"`
	}{
		ObjectMeta: c.ObjectMeta,
		Options:    options,
		Link:       c.Link,
		Template:   c.Template,
	}
}

// GetMeta returns chart API object metadata.
func (c Chart) GetMeta() ObjectMeta {
	return c.ObjectMeta
}

// SetMeta sets chart API object metadata.
func (c *Chart) SetMeta(meta ObjectMeta) {
	c.ObjectMeta = meta
}

// Validate checks for chart API object validity.
func (c Chart) Validate() error {
	err := c.ObjectMeta.Validate()
	if err != nil {
		return err
	}

	switch {
	case len(c.Series) == 0 && c.Link == "":
		return errors.Wrap(ErrInvalid, "missing field: series or link")

	case len(c.Series) > 0 && c.Link != "":
		return errors.Wrap(ErrInvalid, "conflicting fields: series and link")
	}

	return nil
}

// Variables parses the chart API object for template variables references and
// returns their names if found.
func (c Chart) Variables() ([]string, error) {
	if !c.Template {
		return nil, nil
	}

	var data string

	if c.Options.Axes.Y.Left.Label != "" {
		data += fmt.Sprintf("\xff%s", c.Options.Axes.Y.Left.Label)
	}

	if c.Options.Axes.Y.Right.Label != "" {
		data += fmt.Sprintf("\xff%s", c.Options.Axes.Y.Right.Label)
	}

	if c.Options.Title != "" {
		data += fmt.Sprintf("\xff%s", c.Options.Title)
	}

	for _, series := range c.Series {
		data += fmt.Sprintf("\xff%s", series.Expr)
	}

	tmpl := template.New()

	err := tmpl.Parse(data)
	if err != nil {
		return nil, err
	}

	return tmpl.Variables(), nil
}

// ChartOptions are chart options.
// +store:generate=type
type ChartOptions struct {
	Axes      *ChartAxes         `json:"axes,omitempty"`
	Legend    bool               `json:"legend"`
	Markers   []Marker           `json:"markers,omitempty"`
	Title     string             `json:"title,omitempty"`
	Type      ChartType          `json:"type,omitempty"`
	Variables []TemplateVariable `json:"variables,omitempty"`
}

// ChartAxes are chart axes options.
type ChartAxes struct {
	X *ChartXAxis `json:"x,omitempty"`
	Y *ChartYAxes `json:"y,omitempty"`
}

// ChartXAxis are chart X axis options.
type ChartXAxis struct {
	Show bool `json:"show"`
}

// ChartYAxes are chart Y axes options.
type ChartYAxes struct {
	Center bool        `json:"center,omitempty"`
	Left   *ChartYAxis `json:"left,omitempty"`
	Right  *ChartYAxis `json:"right,omitempty"`
}

// ChartYAxis are chart Y axis options.
type ChartYAxis struct {
	Show  bool      `json:"show"`
	Label string    `json:"label,omitempty"`
	Max   *float64  `json:"max,omitempty"`
	Min   *float64  `json:"min,omitempty"`
	Stack StackMode `json:"stack,omitempty"`
	Unit  *Unit     `json:"unit,omitempty"`
}

// StackMode is a chart series stacking mode.
type StackMode string

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (s *StackMode) UnmarshalText(b []byte) error {
	switch v := StackMode(b); v {
	case StackNone, StackNormal, StackPercent:
		*s = v
		return nil
	}

	return fmt.Errorf("unsupported stack mode: %s", b)
}

// Stack modes:
const (
	StackNone    StackMode = ""
	StackNormal  StackMode = "normal"
	StackPercent StackMode = "percent"
)

// ChartType is a chart type.
type ChartType string

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (c *ChartType) UnmarshalText(b []byte) error {
	switch v := ChartType(b); v {
	case ChartArea, ChartBar, ChartLine:
		*c = v
		return nil
	}

	return fmt.Errorf("unsupported chart type: %s", b)
}

// Chart types:
const (
	ChartArea ChartType = "area"
	ChartBar  ChartType = "bar"
	ChartLine ChartType = "line"
)

// Marker is a chart series marker.
type Marker struct {
	Value series.Value `json:"value"`
	Label string       `json:"label,omitempty"`
	Color string       `json:"color,omitempty"`
	Axis  string       `json:"axis,omitempty"`
}

// Unit is a chart unit.
type Unit struct {
	Type UnitType `json:"type"`
	Base string   `json:"base,omitempty"`
}

// UnitType is a chart unit type.
type UnitType string

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (u *UnitType) UnmarshalText(b []byte) error {
	switch v := UnitType(b); v {
	case UnitBinary, UnitCount, UnitDuration, UnitMetric, UnitNone:
		*u = v
		return nil
	}

	return fmt.Errorf("unsupported unit type: %s", b)
}

// Unit types:
const (
	UnitNone     UnitType = ""
	UnitBinary   UnitType = "binary"
	UnitCount    UnitType = "count"
	UnitDuration UnitType = "duration"
	UnitMetric   UnitType = "metric"
)

// SeriesList is a list of chart series.
// +store:generate=type
type SeriesList []Series

// Series is a chart series.
type Series struct {
	Expr    string         `json:"expr"`
	Options *SeriesOptions `json:"options,omitempty"`
}

// SeriesOptions are chart series options.
type SeriesOptions struct {
	Color string `json:"color,omitempty"`
	Axis  string `json:"axis,omitempty"`
}

// ChartList is a list of chart API objects.
type ChartList []Chart

// Objects satisfies the ObjectList interface.
func (c ChartList) Objects() []Object {
	l := make([]Object, len(c))

	for idx, chart := range c {
		x := chart
		l[idx] = &x
	}

	return l
}

// Len satisfies the ObjectList interface.
func (c ChartList) Len() int {
	return len(c)
}

// Dashboard is a dashboard API object.
type Dashboard struct {
	ObjectMeta `json:",inline"`
	Options    *DashboardOptions `json:"options,omitempty"`
	Layout     GridLayout        `json:"layout,omitempty"`
	Items      DashboardItems    `json:"items,omitempty"`
	Parent     string            `json:"parent,omitempty"`
	Link       string            `json:"link,omitempty"`
	Template   bool              `json:"template"`
	References []Reference       `json:"references,omitempty"`
}

// Excerpt returns an excerpted version of the dashboard API object.
func (d Dashboard) Excerpt() interface{} {
	options := map[string]interface{}{}

	if d.Options.Title != "" {
		options["title"] = d.Options.Title
	}

	if len(d.Options.Variables) > 0 {
		options["variables"] = d.Options.Variables
	}

	return struct {
		ObjectMeta `json:",inline"`
		Options    map[string]interface{} `json:"options,omitempty"`
		Link       string                 `json:"link,omitempty"`
		Template   bool                   `json:"enabled"`
	}{
		ObjectMeta: d.ObjectMeta,
		Options:    options,
		Link:       d.Link,
		Template:   d.Template,
	}
}

// GetMeta returns dashboard API object metadata.
func (d Dashboard) GetMeta() ObjectMeta {
	return d.ObjectMeta
}

// SetMeta sets dashboard API object metadata.
func (d *Dashboard) SetMeta(meta ObjectMeta) {
	d.ObjectMeta = meta
}

// Validate checks for dashboard API object validity.
func (d Dashboard) Validate() error {
	err := d.ObjectMeta.Validate()
	if err != nil {
		return err
	}

	hasItems := len(d.Items) > 0

	switch {
	case !hasItems && d.Link == "":
		return errors.Wrap(ErrInvalid, "missing field: items or link")

	case hasItems && d.Link != "":
		return errors.Wrap(ErrInvalid, "conflicting fields: items and link")
	}

	return nil
}

// Variables parses the dashboard API object for template variables references
// and returns their names if found.
func (d Dashboard) Variables() ([]string, error) {
	if !d.Template {
		return nil, nil
	}

	var data string

	if d.Options.Title != "" {
		data += fmt.Sprintf("\xff%s", d.Options.Title)
	}

	tmpl := template.New()

	err := tmpl.Parse(data)
	if err != nil {
		return nil, err
	}

	return tmpl.Variables(), nil
}

// DashboardOptions are dashboard options.
// +store:generate=type
type DashboardOptions struct {
	Title     string             `json:"title,omitempty"`
	Variables []TemplateVariable `json:"variables,omitempty"`
}

// DashboardItem is a dashboard object.
type DashboardItem struct {
	Type    DashboardItemType      `json:"type"`
	Layout  GridItemLayout         `json:"layout"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// DashboardItems is a list of dashboard items.
// +store:generate=type
type DashboardItems []DashboardItem

// DashboardItemType is a dashboard item type:
type DashboardItemType string

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (d *DashboardItemType) UnmarshalText(b []byte) error {
	switch v := DashboardItemType(b); v {
	case DashboardItemChart, DashboardItemText:
		*d = v
		return nil
	}

	return fmt.Errorf("unsupported dashboard item type: %s", b)
}

// Dashboard item types:
const (
	DashboardItemChart DashboardItemType = "chart"
	DashboardItemText  DashboardItemType = "text"
)

// GridLayout is a dashboard grid layout.
// +store:generate=type
type GridLayout struct {
	Columns   uint `json:"columns"`
	RowHeight uint `json:"rowHeight"`
	Rows      uint `json:"rows"`
}

// GridItemLayout is a dashboard grid item layout.
type GridItemLayout struct {
	X uint `json:"x"`
	Y uint `json:"y"`
	W uint `json:"w"`
	H uint `json:"h"`
}

// DashboardList is a list of dashboard API objects.
type DashboardList []Dashboard

// Objects satisfies the ObjectList interface.
func (d DashboardList) Objects() []Object {
	l := make([]Object, len(d))

	for idx, dashboard := range d {
		x := dashboard
		l[idx] = &x
	}

	return l
}

// Len satisfies the ObjectList interface.
func (d DashboardList) Len() int {
	return len(d)
}

// Provider is a provider API object.
type Provider struct {
	ObjectMeta   `json:",inline"`
	Connector    ProviderConnector `json:"connector"`
	Filters      ProviderFilters   `json:"filters"`
	PollInterval time.Duration     `json:"pollInterval"`
	Enabled      bool              `json:"enabled"`
	Error        string            `json:"error,omitempty"`
}

// Excerpt returns an excerpted version of the provider API object.
func (p Provider) Excerpt() interface{} {
	return struct {
		ObjectMeta `json:",inline"`
		Enabled    bool   `json:"enabled"`
		Error      string `json:"error,omitempty"`
	}{
		ObjectMeta: p.ObjectMeta,
		Enabled:    p.Enabled,
		Error:      p.Error,
	}
}

// GetMeta returns provider API object metadata.
func (p Provider) GetMeta() ObjectMeta {
	return p.ObjectMeta
}

// SetMeta sets provider API object metadata.
func (p *Provider) SetMeta(meta ObjectMeta) {
	p.ObjectMeta = meta
}

// Validate checks for provider API object validity.
func (p Provider) Validate() error {
	err := p.ObjectMeta.Validate()
	if err != nil {
		return err
	}

	if p.Connector.Type == "" {
		return errors.Wrap(ErrInvalid, "missing field: connector.type")
	} else if len(p.Connector.Settings) == 0 {
		return errors.Wrap(ErrInvalid, "missing field: connector.settings")
	}

	return nil
}

// ProviderConnector is a provider connector configuration.
// +store:generate=type
type ProviderConnector connector.Config

// ProviderFilters are provider filters.
// +store:generate=type
type ProviderFilters []filter.Rule

// ProviderList is a list of provider API objects.
type ProviderList []Provider

// Objects satisfies the ObjectList interface.
func (p ProviderList) Objects() []Object {
	l := make([]Object, len(p))

	for idx, provider := range p {
		x := provider
		l[idx] = &x
	}

	return l
}

// Len satisfies the ObjectList interface.
func (p ProviderList) Len() int {
	return len(p)
}

// Reference is a value reference.
type Reference struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// Template is an API object template interface.
type Template interface {
	Object
	Variables() ([]string, error)

	template()
}

func (Chart) template()     {}
func (Dashboard) template() {}

// TemplateVariable is a template variable.
type TemplateVariable struct {
	Name    string `json:"name"`
	Value   string `json:"value,omitempty"`
	Label   string `json:"label,omitempty"`
	Filter  string `json:"filter,omitempty"`
	Dynamic bool   `json:"dynamic"`
}
