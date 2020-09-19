// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

// Package labels provides the metrics labeling system.
package labels

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"facette.io/facette/pkg/errors"
)

// Reserved label names:
const (
	Name     = "__name__"
	Provider = "__provider__"
)

// Labels is a labels list.
type Labels []Label

// New creates a new labels list instance.
func New(ls ...Label) Labels {
	return append(Labels{}, ls...)
}

// Append appends a label to the list.
func (l *Labels) Append(ls ...Label) {
	*l = append(*l, ls...)
}

// Copy returns a copy of the labels list.
func (l Labels) Copy() Labels {
	return append(Labels(nil), l...)
}

// Delete deletes a label given its name.
func (l *Labels) Delete(name string) {
	for idx, label := range *l {
		if label.Name == name {
			*l = append((*l)[:idx], (*l)[idx+1:]...)
			break
		}
	}
}

// Get returns the value associated with the given name. It returns an empty
// string if not found.
func (l Labels) Get(name string) string {
	for _, label := range l {
		if label.Name == name {
			return label.Value
		}
	}

	return ""
}

// Len satisfies the sort.Interface interface.
func (l Labels) Len() int {
	return len(l)
}

// Less satisfies the sort.Interface interface.
func (l Labels) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

// Match returns whether the labels list matches with the given matcher.
func (l Labels) Match(matcher Matcher) bool {
	if matcher == nil {
		return true
	}

	var count int

	expected := len(matcher)

	for _, cond := range matcher {
		for _, label := range l {
			if cond.Match(label) {
				count++
			}

			// Stop iteration if expected matching conditions count is reached
			if count == expected {
				return true
			}
		}
	}

	return false
}

// Pop deletes a label given its name returning its associated value. It
// returns an empty string if the label wasn't found.
func (l *Labels) Pop(name string) string {
	var v string

	pos := -1

	for idx, label := range *l {
		if label.Name == name {
			pos = idx
			break
		}
	}

	if pos != -1 {
		v = (*l)[pos].Value
		*l = append((*l)[:pos], (*l)[pos+1:]...)
	}

	return v
}

// Set sets the value for the given label name. It will replace any preexisting
// value present for that name.
func (l *Labels) Set(name, value string) {
	for idx, label := range *l {
		if label.Name == name {
			(*l)[idx].Value = value
			break
		}
	}

	*l = append(*l, Label{Name: name, Value: value})
}

// String satisfies the fmt.Stringer interface.
// nolint:gosec
func (l Labels) String() string {
	b := bytes.NewBuffer(nil)
	b.WriteByte('{')

	for idx, label := range l {
		if idx > 0 {
			b.WriteByte(',')
		}

		b.WriteString(label.String())
	}

	b.WriteByte('}')

	return b.String()
}

// Swap satisfies the sort.Interface interface.
func (l Labels) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Validate checks whether or not the labels list is valid and returns an error
// accordingly.
//
// It assumes that the labels list is sorted for duplicated names detection.
func (l Labels) Validate() error {
	var (
		last string
		err  error
	)

	for _, label := range l {
		if label.Name == last {
			return fmt.Errorf("duplicate label name: %s", label.Name)
		}

		err = label.Validate()
		if err != nil {
			return err
		}

		last = label.Name
	}

	return nil
}

// Label is a name and value pair.
type Label struct {
	Name  string
	Value string
}

// String satisfies the fmt.Stringer interface.
func (l Label) String() string {
	return fmt.Sprintf("%s=%q", l.Name, l.Value)
}

// Validate checks whether or not the name and value pair is valid and returns
// an error accordingly.
func (l Label) Validate() error {
	switch {
	case !NameValid(l.Name):
		if l.Name == "" {
			return errors.New("empty label name")
		}

		return fmt.Errorf("invalid label name: %s", l.Name)

	case !ValueValid(l.Value):
		if l.Value == "" {
			return errors.New("empty label value")
		}

		return fmt.Errorf("invalid label value: %s", l.Value)
	}

	return nil
}

// NameValid returns whether or not the given label name is valid.
//
// A label name mustn't be empty, start by a letter and may contain letters,
// digits and underscores.
func NameValid(name string) bool {
	if name == "" {
		return false
	}

	for idx, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9' && idx > 0) || c == '_') {
			return false
		}
	}

	return true
}

// ValueValid returns whether or not the given label value is valid.
//
// A label value mustn't be empty and be a valid UTF-8 string.
func ValueValid(value string) bool {
	return value != "" && utf8.ValidString(value)
}
