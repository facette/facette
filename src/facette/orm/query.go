package orm

import (
	"database/sql"
	"reflect"
	"strings"
)

type query struct {
	columns []string
	from    string
	where   []string
	offset  int
	limit   int
	orders  []string
	args    []interface{}
	raw     string
	db      *DB
	skipLog bool
}

func (q *query) appendArg(value interface{}) string {
	q.args = append(q.args, q.db.driver.adaptValue(reflect.ValueOf(value)).Interface())
	return q.db.driver.bindVar(len(q.args))
}

func (q *query) Select(columns ...string) *query {
	q.columns = append(q.columns, columns...)
	return q
}

func (q *query) From(from string) *query {
	q.from = from
	return q
}

func (q *query) Where(condition string, args ...interface{}) *query {
	// Adapt variable bindings if needed
	for _, arg := range args {
		if value := reflect.ValueOf(arg); value.Kind() == reflect.Slice {
			if _, ok := value.Interface().([]byte); ok {
				condition = strings.Replace(condition, "?", q.appendArg(arg), 1)
				continue
			}

			tmpBinds := []string{}
			for i := 0; i < value.Len(); i++ {
				tmpBinds = append(tmpBinds, q.appendArg(value.Index(i).Interface()))
			}

			condition = strings.Replace(condition, "?", strings.Join(tmpBinds, ", "), 1)
		} else {
			condition = strings.Replace(condition, "?", q.appendArg(arg), 1)
		}
	}

	q.where = append(q.where, condition)

	return q
}

func (q *query) Count() *query {
	q.columns = []string{"count(*)"}
	q.offset = 0
	q.limit = 0
	return q
}

func (q *query) Offset(offset int) *query {
	q.offset = offset
	return q
}

func (q *query) Limit(limit int) *query {
	q.limit = limit
	return q
}

func (q *query) OrderBy(terms ...string) *query {
	q.orders = append(q.orders, terms...)
	return q
}

func (q *query) Raw(query string, args ...interface{}) *query {
	q.raw = query
	q.args = args
	return q
}

func (q *query) Row() *sql.Row {
	var row *sql.Row

	query, args := q.build()
	if !q.skipLog {
		q.db.logQuery(query, args...)
	}

	if q.db.tx != nil {
		row = q.db.tx.QueryRow(query, args...)
	} else {
		row = q.db.sqlDB.QueryRow(query, args...)
	}

	q.reset()

	return row
}

func (q *query) Rows() (*sql.Rows, error) {
	var (
		rows *sql.Rows
		err  error
	)

	query, args := q.build()
	if !q.skipLog {
		q.db.logQuery(query, args...)
	}

	if q.db.tx != nil {
		rows, err = q.db.tx.Query(query, args...)
	} else {
		rows, err = q.db.sqlDB.Query(query, args...)
	}

	q.reset()

	return rows, err
}

func (q *query) Result() (sql.Result, error) {
	var (
		result sql.Result
		err    error
	)

	query, args := q.build()
	if !q.skipLog {
		q.db.logQuery(query, args...)
	}

	if q.db.tx != nil {
		result, err = q.db.tx.Exec(query, args...)
	} else {
		result, err = q.db.sqlDB.Exec(query, args...)
	}

	q.reset()

	return result, err
}

func (q *query) quiet() {
	q.skipLog = true
}

func (q *query) build() (string, []interface{}) {
	// Return raw query if set
	if q.raw != "" {
		return q.raw, q.args
	}

	// Build query from properties
	query := "SELECT "

	if len(q.columns) > 0 {
		query += strings.Join(q.columns, ", ")
	} else {
		query += "*"
	}

	if q.from != "" {
		if strings.Contains(q.from, ".") {
			parts := strings.SplitN(q.from, ".", 2)
			query += " FROM " + q.db.driver.QuoteName(parts[0]) + "." + q.db.driver.QuoteName(parts[1])
		} else {
			query += " FROM " + q.db.driver.QuoteName(q.from)
		}
	}

	if len(q.where) > 0 {
		query += " WHERE " + strings.Join(q.where, " AND ")
	}

	if len(q.orders) > 0 {
		query += " ORDER BY " + strings.Join(q.orders, ", ")
	}

	if q.limit > 0 {
		query += " " + q.db.driver.LimitClause(q.offset, q.limit)
	}

	return query, q.args
}

func (q *query) reset() {
	*q = *newQuery(q.db)
}

func newQuery(db *DB) *query {
	return &query{
		columns: []string{},
		where:   []string{},
		args:    []interface{}{},
		db:      db,
	}
}
