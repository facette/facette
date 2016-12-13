package httproute

import (
	"context"
	"strings"
)

// pattern represents an endpoint pattern instance.
type pattern struct {
	value       string
	hasSlash    bool
	hasWildcard bool
}

// newPattern creates a new endpoint pattern instance.
func newPattern(value string) *pattern {
	p := &pattern{}

	switch {
	case strings.HasSuffix(value, "/*"):
		p.value = strings.TrimSuffix(value, "/*")
		p.hasWildcard = true

	case strings.HasSuffix(value, "/"):
		p.value = strings.TrimSuffix(value, "/")
		p.hasSlash = true

	default:
		p.value = value
	}

	if p.value == "" {
		p.value = "/"
		p.hasSlash = true
	}

	return p
}

// match checks whether or not a path matches the pattern, returning the associated context.
func (p *pattern) match(path string) (context.Context, bool) {
	var i, j int

	// Initialize context
	ctx := context.Background()

	// Remove trailing slash on patn for future comparison
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	// Check for exact path match
	if path == p.value && !strings.Contains(p.value, ":") {
		return ctx, true
	}

	// Try to match path on pattern
	vLength := len(p.value)
	pLength := len(path)

	for i < pLength {
		switch {
		case j >= vLength:
			// Path has remainder, so check for wildcard in pattern
			if p.hasWildcard {
				return ctx, true
			}

			return nil, false

		case p.value[j] == ':':
			var (
				key, value string
				next       byte
			)

			// Append new value to the pattern context
			key, next, j = matchNext(p.value, matchKeyStop, j+1)
			value, _, i = matchNext(path, matchByte(next), i)

			// Stop if a sub-level has been found in value
			if strings.Contains(value, "/") {
				return nil, false
			}

			ctx = context.WithValue(ctx, key, value)

		case path[i] == p.value[j]:
			i++
			j++

		default:
			return nil, false
		}
	}

	if j != vLength {
		// Pattern value has a remainder, check if ending with a key giving it an empty value
		if p.value[j] == ':' {
			if key, _, idx := matchNext(p.value, matchKeyStop, j+1); idx == vLength {
				ctx = context.WithValue(ctx, key, "")
				return ctx, true
			}
		}

		return nil, false
	}

	return ctx, true
}

func matchNext(s string, f func(r rune) bool, i int) (string, byte, int) {
	idx := strings.IndexFunc(s[i:], f)
	if idx == -1 {
		return s[i:], 0, len(s)
	}

	idx += i

	return s[i:idx], s[idx], idx
}

func matchByte(b byte) func(c rune) bool {
	c1 := rune(b)

	return func(c2 rune) bool {
		return c2 == c1
	}
}

func matchKeyStop(c rune) bool {
	return (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_'
}
