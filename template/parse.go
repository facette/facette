package template

import (
	"sort"
	"text/template/parse"

	"facette.io/facette/set"
)

// Parse parses a given string, returning the list of template keys.
func Parse(data string) ([]string, error) {
	// Parse response for template keys
	trees, err := parse.Parse("inline", data, "", "")
	if err != nil {
		return nil, ErrInvalidTemplate
	}

	keys := set.New()
	for _, node := range trees["inline"].Root.Nodes {
		if action, ok := node.(*parse.ActionNode); ok {
			if len(action.Pipe.Cmds) != 1 {
				continue
			}

			for _, arg := range action.Pipe.Cmds[0].Args {
				if field, ok := arg.(*parse.FieldNode); ok {
					if len(field.Ident) == 1 {
						// Found a new key, add it to the results list
						keys.Add(field.Ident[0])
					} else {
						return nil, ErrInvalidTemplate
					}
				}
			}
		}
	}

	result := set.StringSlice(keys)
	sort.Strings(result)

	return result, nil
}
