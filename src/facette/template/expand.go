package template

import (
	"bytes"
	"fmt"
	"text/template/parse"
)

// Expand parses a given string replacing template keys with attributes values.
func Expand(data string, attrs map[string]interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)

	// Parse response for template keys
	trees, err := parse.Parse("inline", data, "", "")
	if err != nil {
		return "", ErrInvalidTemplate
	}

	for _, node := range trees["inline"].Root.Nodes {
		if text, ok := node.(*parse.TextNode); ok {
			buf.Write(text.Text)
		} else if action, ok := node.(*parse.ActionNode); ok {
			if len(action.Pipe.Cmds) != 1 {
				continue
			}

			for _, arg := range action.Pipe.Cmds[0].Args {
				fmt.Printf(">>> %#v\n", arg)
				if field, ok := arg.(*parse.FieldNode); ok {
					if len(field.Ident) == 1 {
						// Replace template key with attribute value
						if v, ok := attrs[field.Ident[0]]; ok && v != nil {
							buf.WriteString(fmt.Sprintf("%v", v))
						}
					} else {
						return "", ErrInvalidTemplate
					}
				}
			}
		}
	}

	return buf.String(), nil
}
