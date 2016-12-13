package backend

import (
	"bytes"
	"fmt"
	"text/template"
)

func expandStringTemplate(input string, attr map[string]interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)

	tpl, err := template.New("inline").Parse(input)
	if err != nil {
		return input, fmt.Errorf("failed to parse template: %s", err)
	} else if err = tpl.Execute(buf, attr); err != nil {
		return input, fmt.Errorf("failed to execute template: %s", err)
	}

	return buf.String(), nil
}
