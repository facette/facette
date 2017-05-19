package template

import (
	"reflect"
	"testing"
)

func Test_Expand(t *testing.T) {
	testExpand("This is a sample text!", false, "This {{ .a }} a {{ .b }}!",
		map[string]interface{}{"a": "is", "b": "sample text"}, t)
}

func Test_Expand_Empty(t *testing.T) {
	testExpand("This is a !", false, "This {{ .a }} a {{ .b }}!", map[string]interface{}{"a": "is"}, t)
}

func Test_Expand_Fail_Syntax(t *testing.T) {
	testExpand("", true, "This {{ .a } a {{ .b }}!", map[string]interface{}{"a": "is", "b": "sample text"}, t)
}

func Test_Expand_Fail_Ident(t *testing.T) {
	testExpand("", true, "This {{ .a }} a {{ b }}!", map[string]interface{}{"a": "is", "b": "sample text"}, t)
}

func Test_Parse(t *testing.T) {
	testParse([]string{"a", "b"}, false, "This {{ .a }} a {{ .b }}!", t)
}

func Test_Parse_Fail_Syntax(t *testing.T) {
	testParse(nil, true, "This {{ .a } a {{ .b }}!", t)
}

func Test_Parse_Fail_Ident(t *testing.T) {
	testParse(nil, true, "This {{ .a }} a {{ b }}!", t)
}

func testExpand(expected string, expectedErr bool, data string, attrs map[string]interface{}, t *testing.T) {
	result, err := Expand(data, attrs)
	if expectedErr && err == nil || !expectedErr && err != nil {
		t.Logf("\nExpected an error\nbut got  %#v", err)
		t.Fail()
	} else if result != expected {
		t.Logf("\nExpected %q\nbut got  %q", expected, result)
		t.Fail()
	}
}

func testParse(expected []string, expectedErr bool, data string, t *testing.T) {
	result, err := Parse(data)
	if expectedErr && err == nil || !expectedErr && err != nil {
		t.Logf("\nExpected an error\nbut got  %#v", err)
		t.Fail()
	} else if !reflect.DeepEqual(result, expected) {
		t.Logf("\nExpected %q\nbut got  %q", expected, result)
		t.Fail()
	}
}
