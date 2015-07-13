package provider

import (
	"regexp"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/logger"
	"github.com/fatih/set"
)

type filterChain struct {
	Input  chan *catalog.Record
	output chan *catalog.Record
	rules  []*config.ProviderFilterConfig
}

func newFilterChain(filters []*config.ProviderFilterConfig, output chan *catalog.Record) filterChain {
	chain := filterChain{
		Input:  make(chan *catalog.Record),
		output: output,
		rules:  make([]*config.ProviderFilterConfig, 0),
	}

	actionSet := set.New(set.NonThreadSafe)
	actionSet.Add("rewrite", "discard", "sieve")

	targetSet := set.New(set.NonThreadSafe)
	targetSet.Add("any", "origin", "source", "metric")

	for _, filter := range filters {
		if filter.Target == "" {
			filter.Target = "any"
		}

		if !actionSet.Has(filter.Action) {
			logger.Log(logger.LevelWarning, "provider", "unknown `%s' filter action, discarding", filter.Action)
			continue
		}

		if !targetSet.Has(filter.Target) {
			logger.Log(logger.LevelWarning, "provider", "unknown `%s' filter target, discarding", filter.Target)
			continue
		}

		re, err := regexp.Compile(filter.Pattern)
		if err != nil {
			logger.Log(logger.LevelWarning, "server", "unable to compile filter pattern: %s, discarding", err)
			continue
		}

		filter.PatternRegexp = re

		chain.rules = append(chain.rules, filter)
	}

	go func(chain filterChain) {
		for record := range chain.Input {
			// Forward record if no rule defined
			if len(chain.rules) == 0 {
				chain.output <- record
				continue
			}

			// Keep a copy of original names
			record.OriginalOrigin = record.Origin
			record.OriginalSource = record.Source
			record.OriginalMetric = record.Metric

			for _, rule := range chain.rules {
				if (rule.Target == "origin" || rule.Target == "any") && !rule.PatternRegexp.MatchString(record.Origin) {
					if rule.Action == "sieve" {
						logger.Log(
							logger.LevelDebug,
							"server",
							"discard record %s, as origin doesn't match `%s' sieve pattern",
							record,
							rule.Pattern,
						)
						goto nextRecord
					}
				}

				if (rule.Target == "origin" || rule.Target == "any") && rule.PatternRegexp.MatchString(record.Origin) {
					if rule.Action == "discard" {
						logger.Log(
							logger.LevelDebug,
							"server",
							"discard record %s, as origin matches `%s' pattern",
							record,
							rule.Pattern,
						)
						goto nextRecord
					}

					if rule.Action == "rewrite" {
						record.Origin = rule.PatternRegexp.ReplaceAllString(record.Origin, rule.Into)
					}
				}

				if (rule.Target == "source" || rule.Target == "any") && !rule.PatternRegexp.MatchString(record.Source) {
					if rule.Action == "sieve" {
						logger.Log(
							logger.LevelDebug,
							"server",
							"discard record %s, as source doesn't match `%s' sieve pattern",
							record,
							rule.Pattern,
						)
						goto nextRecord
					}
				}

				if (rule.Target == "source" || rule.Target == "any") && rule.PatternRegexp.MatchString(record.Source) {
					if rule.Action == "discard" {
						logger.Log(
							logger.LevelDebug,
							"server",
							"discard record %s, as source matches `%s' pattern",
							record,
							rule.Pattern,
						)
						goto nextRecord
					}

					if rule.Action == "rewrite" {
						record.Source = rule.PatternRegexp.ReplaceAllString(record.Source, rule.Into)
					}
				}

				if (rule.Target == "metric" || rule.Target == "any") && !rule.PatternRegexp.MatchString(record.Metric) {
					if rule.Action == "sieve" {
						logger.Log(
							logger.LevelDebug,
							"server",
							"discard record %s, as metric doesn't match `%s' sieve pattern",
							record,
							rule.Pattern,
						)
						goto nextRecord
					}
				}

				if (rule.Target == "metric" || rule.Target == "any") && rule.PatternRegexp.MatchString(record.Metric) {
					if rule.Action == "discard" {
						logger.Log(
							logger.LevelDebug,
							"server",
							"discard record %s, as metric matches `%s' pattern",
							record,
							rule.Pattern,
						)
						goto nextRecord
					}

					if rule.Action == "rewrite" {
						record.Metric = rule.PatternRegexp.ReplaceAllString(record.Metric, rule.Into)
					}
				}
			}

			chain.output <- record
		nextRecord:
		}
	}(chain)

	return chain
}
