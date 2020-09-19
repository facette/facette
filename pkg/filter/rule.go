// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package filter

import (
	"fmt"
	"regexp"
	"regexp/syntax"

	"facette.io/facette/pkg/errors"
)

// Rule is a metrics filter rule.
type Rule struct {
	Action  Action            `json:"action"`
	Label   string            `json:"label"`
	Pattern Pattern           `json:"pattern"`
	Into    string            `json:"into,omitempty"`
	Targets map[string]string `json:"targets,omitempty"`
}

// Action is a metrics filter action.
type Action string

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (a *Action) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		return errors.New("invalid filter action")
	}

	switch v := Action(b); v {
	case ActionDiscard, ActionRelabel, ActionRewrite, ActionSieve:
		*a = v
		return nil
	}

	return fmt.Errorf("invalid filter action: %s", b)
}

// Actions:
const (
	ActionDiscard Action = "discard"
	ActionRelabel Action = "relabel"
	ActionRewrite Action = "rewrite"
	ActionSieve   Action = "sieve"
)

// Pattern is a metrics filter pattern.
type Pattern struct {
	s  string
	re *regexp.Regexp
}

// MarshalText satisfies the encoding.TextMarshaler interface.
func (f Pattern) MarshalText() ([]byte, error) {
	return []byte(f.s), nil
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface.
func (f *Pattern) UnmarshalText(b []byte) error {
	f.s = string(b)

	var err error

	f.re, err = regexp.Compile(f.s)
	if err != nil {
		var xerr *syntax.Error

		ok := errors.As(err, &xerr)
		if ok {
			return fmt.Errorf("invalid filter pattern: %s: `%s`", xerr.Code, xerr.Expr)
		}

		return err
	}

	return nil
}
