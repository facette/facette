package orm

import (
	"fmt"
	"log"
	"strings"
)

// SetLogger sets the current database connection logger.
func (db *DB) SetLogger(logger *log.Logger) {
	db.logger = logger
}

func (db *DB) logQuery(query string, args ...interface{}) {
	if db.logger == nil {
		return
	}

	values := []string{}
	for _, arg := range args {
		format := "%v"

		switch arg.(type) {
		case string, fmt.Stringer:
			format = "%q"
		}

		values = append(values, fmt.Sprintf(format, arg))
	}

	mesg := fmt.Sprintf("\nquery: %s\n", query)
	if len(args) > 0 {
		mesg += fmt.Sprintf(" args: %s\n", strings.Join(values, ", "))
	}

	db.logger.Print(mesg)
}
