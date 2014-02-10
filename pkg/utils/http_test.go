package utils

import (
	"net/http"
	"testing"
)

func Test_RequestGetContentType(test *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		test.Fatal(err.Error())
	}

	if result := RequestGetContentType(request); result != "" {
		test.Logf("\nExpected `%s'\nbut got  `%s'", "", result)
		test.Fail()
	}

	request.Header.Add("Content-Type", "application/json")

	if result := RequestGetContentType(request); result != "application/json" {
		test.Logf("\nExpected `%s'\nbut got  `%s'", "application/json", result)
		test.Fail()
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	if result := RequestGetContentType(request); result != "application/json" {
		test.Logf("\nExpected `%s'\nbut got  `%s'", "application/json", result)
		test.Fail()
	}
}
