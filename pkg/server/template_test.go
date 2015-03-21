package server

import "testing"

func Test_expandStringTemplate(test *testing.T) {
	result, err := expandStringTemplate("{{ .attr1 }} - {{ .attr2 }}", map[string]interface{}{
		"attr1": "value1",
		"attr2": "value2",
	})

	if err != nil {
		test.Fatal(err)
	}

	if result != "value1 - value2" {
		test.Logf("\nExpected %s\nbut got  %s", "value1 - value2", result)
		test.Fail()
	}
}
