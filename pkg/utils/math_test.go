package utils

import (
	"testing"
)

func Test_Round(test *testing.T) {
	var (
		expected int64
		actual   int64
	)

	expected, actual = 0, Round(0.0)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = 0, Round(0.1)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = 1, Round(0.5)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = 1, Round(1.2)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = 2, Round(1.7)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = 0, Round(-0.1)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = -1, Round(-0.5)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = -1, Round(-1.2)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}

	expected, actual = -2, Round(-1.7)
	if expected != actual {
		test.Logf("\nExpected %d\nbut got  %d", expected, actual)
		test.Fail()
	}
}
