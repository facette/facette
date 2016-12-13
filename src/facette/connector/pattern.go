package connector

import (
	"fmt"
	"regexp"
)

const (
	// PatternKeywordSource represents the pattern keyword value for source name.
	PatternKeywordSource = "source"
	// PatternKeywordMetric represents the pattern keyword value for metric name.
	PatternKeywordMetric = "metric"
)

func compilePattern(pattern string) (*regexp.Regexp, error) {
	// Compile regexp pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// Validate pattern keywords
	groups := make(map[string]bool)
	for _, key := range re.SubexpNames() {
		if key == PatternKeywordSource || key == PatternKeywordMetric {
			groups[key] = true
		} else if key != "" {
			return nil, fmt.Errorf("invalid %q pattern keyword", key)
		}
	}

	if _, ok := groups[PatternKeywordSource]; !ok {
		return nil, ErrMissingSourcePattern
	} else if _, ok := groups[PatternKeywordMetric]; !ok {
		return nil, ErrMissingMetricPattern
	}

	return re, nil
}

func matchPattern(re *regexp.Regexp, input string) ([2]string, error) {
	var result [2]string

	m := re.FindStringSubmatch(input)
	if len(m) == 0 {
		return result, fmt.Errorf("series %q does not match pattern", input)
	}

	if re.SubexpNames()[1] == PatternKeywordSource {
		result = [2]string{m[1], m[2]}
	} else {
		result = [2]string{m[2], m[1]}
	}

	return result, nil
}
