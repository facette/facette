package httproute

import (
	"reflect"
	"strings"
	"testing"
)

type matchTest struct {
	pattern string
	path    string
	match   bool
	data    map[string]interface{}
}

func TestPatternMatch(t *testing.T) {
	for _, mt := range []matchTest{
		{"/", "/", true, nil},
		{"/", "/foo", false, nil},
		{"/*", "/", true, nil},
		{"/*", "/foo", true, nil},
		{"/foo/:key", "/foo/", false, nil},
		{"/foo/:key", "/foo/a", true, map[string]interface{}{"key": "a"}},
		{"/foo/:key", "/foo/a/b", false, nil},
		{"/foo/:key", "/foo/:a", true, map[string]interface{}{"key": ":a"}},
		{"/foo/:key_a", "/foo/a", true, map[string]interface{}{"key_a": "a"}},
		{"/foo/:key/bar", "/foo/a", false, nil},
		{"/foo/:key/bar", "/foo/a/bar", true, map[string]interface{}{"key": "a"}},
		{"/foo/:key1/bar/:key2", "/foo/a/bar", false, nil},
		{"/foo/:key1/bar/:key2", "/foo/a/bar/b", true, map[string]interface{}{"key1": "a", "key2": "b"}},
		{"/foo/:key.ext", "/foo/.ext", true, map[string]interface{}{"key": ""}},
		{"/foo/:key.ext", "/foo/a.ext", true, map[string]interface{}{"key": "a"}},
		{"/foo/:key.", "/foo/a.", true, map[string]interface{}{"key": "a"}},
		{"/foo/:key1:key2", "/foo/a", true, map[string]interface{}{"key1": "a", "key2": ""}},
		{"/foo/:key1.:key2", "/foo/a.b", true, map[string]interface{}{"key1": "a", "key2": "b"}},
		{"/foo/:key1.:key2", "/foo/.b", true, map[string]interface{}{"key1": "", "key2": "b"}},
		{"/foo/:key1.:key2", "/foo/a.", true, map[string]interface{}{"key1": "a", "key2": ""}},
		{"/foo/bar:key", "/foo/bara", true, map[string]interface{}{"key": "a"}},
	} {
		mt1 := mt
		execTestPatternMatch(mt1, t)

		// Testing trailing slashes pattern variants
		if !strings.HasSuffix(mt.pattern, "/*") {
			mt2 := mt
			mt.pattern += "/"
			execTestPatternMatch(mt2, t)
		}

		// Testing trailing slashes paths variants
		mt3 := mt
		mt.path += "/"
		execTestPatternMatch(mt3, t)
	}
}

func execTestPatternMatch(mt matchTest, t *testing.T) {
	p := newPattern(mt.pattern)

	ctx, match := p.match(mt.path)
	if match != mt.match {
		t.Errorf(
			"invalid match for %q path on %q pattern: expected \"%t\" but got \"%t\"",
			mt.path,
			mt.pattern,
			mt.match,
			match,
		)
	} else if ctx != nil && mt.data != nil {
		data := map[string]interface{}{}
		for key := range mt.data {
			data[key] = ctx.Value(key)
		}

		if !reflect.DeepEqual(mt.data, data) {
			t.Errorf(
				"invalid context data for %q path on %q pattern: expected %v but got %v",
				mt.path,
				mt.pattern,
				mt.data,
				data,
			)
		}
	}
}
