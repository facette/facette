package template

import (
	"reflect"
	"testing"
)

func Test_Parse(t *testing.T) {
	testParse([]string{"a", "b"}, false, "This {{ .a }} a {{ .b }}!", t)
}

func Test_Parse_Fail_Syntax(t *testing.T) {
	testParse(nil, true, "This {{ .a } a {{ .b }}!", t)
}

func Test_Parse_Fail_Ident(t *testing.T) {
	testParse(nil, true, "This {{ .a }} a {{ b }}!", t)
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
