package sqlstorage

import (
	"fmt"
	"log"
	"strings"
)

type ormLogger struct {
	sql *log.Logger
}

func (l *ormLogger) Print(v ...interface{}) {
	if v[0] == "sql" {
		args := []string{}
		for _, arg := range v[4].([]interface{}) {
			format := "%v"

			switch arg.(type) {
			case string, fmt.Stringer:
				format = "%q"
			}

			args = append(args, fmt.Sprintf(format, arg))
		}

		l.sql.Print(fmt.Sprintf("%s [%s] in %s\n", v[3], strings.Join(args, ", "), v[2]))
	}
}
