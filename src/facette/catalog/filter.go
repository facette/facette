package catalog

import (
	"fmt"
	"regexp"

	"facette/backend"

	"github.com/fatih/set"
)

const (
	// ActionDiscard represents the discard rule action keyword.
	ActionDiscard = "discard"
	// ActionRewrite represents the rewrite rule action keyword.
	ActionRewrite = "rewrite"
	// ActionSieve represents the sieve rule action keyword.
	ActionSieve = "sieve"

	// TargetAny represents the global target matching keyword.
	TargetAny = "any"
	// TargetOrigin represents the origin target matching keyword.
	TargetOrigin = "origin"
	// TargetSource represents the source target matching keyword.
	TargetSource = "source"
	// TargetMetric represents the metric target matching keyword.
	TargetMetric = "metric"
)

var (
	actions = set.New(
		ActionDiscard,
		ActionRewrite,
		ActionSieve,
	)

	targets = set.New(
		TargetAny,
		TargetOrigin,
		TargetSource,
		TargetMetric,
	)
)

// FilterChain represents a catalog filtering chain instance.
type FilterChain struct {
	Input    chan *Record
	Output   chan *Record
	Messages chan string
	rules    []filterRule
}

// NewFilterChain creates a new catalog filtering chain instance.
func NewFilterChain(rules *backend.ProviderFilters) *FilterChain {
	fc := &FilterChain{
		Input:    make(chan *Record),
		Output:   make(chan *Record),
		Messages: make(chan string),
		rules:    []filterRule{},
	}

	// Parse filter chain rules
	for _, r := range *rules {
		if r.Target == "" {
			r.Target = TargetAny
		}

		if !actions.Has(r.Action) {
			fc.Messages <- fmt.Sprintf("unknown %q filter action, discarding", r.Action)
			continue
		} else if !targets.Has(r.Target) {
			fc.Messages <- fmt.Sprintf("unknown %q filter target, discarding", r.Target)
			continue
		}

		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			fc.Messages <- fmt.Sprintf("unable to compile filter pattern: %s, discarding", err)
			continue
		}

		fc.rules = append(fc.rules, filterRule{ProviderFilter: r, re: re})
	}

	// Start filtering routine
	go func() {
		for record := range fc.Input {
			// Keep a copy of original names
			record.OriginalOrigin = record.Origin
			record.OriginalSource = record.Source
			record.OriginalMetric = record.Metric

			// Forward record if no rule defined
			if len(fc.rules) == 0 {
				fc.Output <- record
				continue
			}

			for _, r := range fc.rules {
				if r.Target == TargetOrigin || r.Target == TargetAny {
					if skip := fc.applyAction(r, record, &record.Origin); skip {
						goto nextRecord
					}
				}

				if r.Target == TargetSource || r.Target == TargetAny {
					if skip := fc.applyAction(r, record, &record.Source); skip {
						goto nextRecord
					}
				}

				if r.Target == TargetMetric || r.Target == TargetAny {
					if skip := fc.applyAction(r, record, &record.Metric); skip {
						goto nextRecord
					}
				}
			}

			fc.Output <- record
		nextRecord:
		}
	}()

	return fc
}

// applyAction applies a filtering chain rule, checking if record should be skipped or not.
func (fc *FilterChain) applyAction(rule filterRule, record *Record, value *string) bool {
	if rule.re.MatchString(*value) {
		switch rule.Action {
		case ActionDiscard:
			fc.Messages <- fmt.Sprintf("matches %q pattern, discarding: %s", rule.Pattern, record)
			return true

		case ActionRewrite:
			*value = rule.re.ReplaceAllString(*value, rule.Into)
		}
	} else {
		switch rule.Action {
		case ActionSieve:
			fc.Messages <- fmt.Sprintf("does not match %q sieve pattern, discarding: %s", rule.Pattern, record)
			return true
		}
	}

	return false
}

type filterRule struct {
	backend.ProviderFilter
	re *regexp.Regexp
}
