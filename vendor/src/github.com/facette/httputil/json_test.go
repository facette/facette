package httputil

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

var (
	jsonData = "{\"key1\":1,\"key2\":2}\n"
	jsonMap  = map[string]int{"key1": 1, "key2": 2}
)

func Test_BindJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/", strings.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	result := make(map[string]int)
	if err := BindJSON(req, &result); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(result, jsonMap) {
		t.Logf("\nExpected %#v\nbut got  %#v", jsonData, result)
		t.Fail()
	}
}

func Test_WriteJSON(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := testWriter{buf}

	if err := WriteJSON(w, jsonMap, http.StatusOK); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(buf.String(), jsonData) {
		t.Logf("\nExpected %q\nbut got  %q", jsonData, buf.String())
		t.Fail()
	}
}

type testWriter struct {
	io.Writer
}

func (testWriter) Header() http.Header { return http.Header{} }
func (testWriter) WriteHeader(int)     { return }
