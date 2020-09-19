// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package filter

import (
	"go.uber.org/zap"

	"facette.io/facette/pkg/catalog"
)

// New creates a new metrics filter.
func New(in <-chan catalog.Metric, rules []Rule) <-chan catalog.Metric {
	out := make(chan catalog.Metric)

	go func() {
		for metric := range in {
			for _, rule := range rules {
				log := zap.L().With(
					zap.String("action", string(rule.Action)),
					zap.String("pattern", rule.Pattern.s),
				)

				if rule.Pattern.re.MatchString(metric.Labels.Get(rule.Label)) {
					switch rule.Action {
					case ActionDiscard:
						log.Debug("metric skipped", zap.String("metric", metric.String()))
						goto skip

					case ActionRelabel:
						relabel(metric, rule)
						continue

					case ActionRewrite:
						rewrite(metric, rule)
						continue
					}
				} else if rule.Action == ActionSieve {
					log.Debug("metric skipped", zap.String("metric", metric.String()))
					goto skip
				}
			}

			out <- metric
		skip:
		}

		close(out)
	}()

	return out
}

func relabel(metric catalog.Metric, rule Rule) {
	value := metric.Labels.Get(rule.Label)
	match := rule.Pattern.re.FindStringSubmatchIndex(value)

	for k, v := range rule.Targets {
		if v == "" {
			metric.Labels.Delete(k)
		} else {
			metric.Labels.Set(k, string(rule.Pattern.re.Expand(nil, []byte(v), []byte(value), match)))
		}
	}
}

func rewrite(metric catalog.Metric, rule Rule) {
	metric.Labels.Set(
		rule.Label,
		rule.Pattern.re.ReplaceAllString(metric.Labels.Get(rule.Label), rule.Into),
	)
}
