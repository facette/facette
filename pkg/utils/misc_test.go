package utils

import (
	"reflect"
	"testing"
)

func Test_Clone(test *testing.T) {
	type testStruct struct {
		FieldA string
		FieldB int
	}

	srcStruct := testStruct{"test", 42}
	dstStruct := testStruct{}

	Clone(srcStruct, &dstStruct)

	if !reflect.DeepEqual(srcStruct, dstStruct) {
		test.Logf("\nExpected %#v\nbut got  %#v", srcStruct, dstStruct)
		test.Fail()
	}
}
