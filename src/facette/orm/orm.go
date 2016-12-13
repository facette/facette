package orm

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
)

// TimeFormat represents the time format used for time values in the database.
const TimeFormat = "2006-01-02 15:04:05"

// DB represents a database connection.
type DB struct {
	logger   *log.Logger
	driver   SQLDriver
	sqlDB    *sql.DB
	models   map[reflect.Type]*model
	migrated map[string]struct{}
	query    *query
	tx       *sql.Tx

	value  interface{}
	model  *model
	result sql.Result
	err    error
}

// Open initializes a new database connection.
func Open(driver, dsn string) (*DB, error) {
	// Initialize backend driver
	drv := newDriver(driver)
	if drv == nil {
		return nil, ErrUnsupportedDriver
	}

	// Open database connection
	sqlDB, err := sql.Open(drv.name(), dsn)
	if err != nil {
		return nil, err
	} else if err = sqlDB.Ping(); err != nil {
		return nil, err
	}

	// Create new database
	db := &DB{
		driver:   drv,
		sqlDB:    sqlDB,
		models:   make(map[reflect.Type]*model),
		migrated: make(map[string]struct{}),
	}

	db.query = newQuery(db)

	// Initialize database driver
	drv.setDB(db)
	if err := drv.init(); err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the current database connection.
func (db *DB) Close() error {
	return db.sqlDB.Close()
}

// Error returns the last encountered error.
func (db *DB) Error() error {
	return db.driver.normalizeError(db.err)
}

// HasTable returns whether or not a table exists in the database.
func (db *DB) HasTable(tableName string) bool {
	return db.driver.hasTable(tableName)
}

// HasColumn returns whether or not a table column exists in the database.
func (db *DB) HasColumn(tableName, columnName string) bool {
	return db.driver.hasColumn(tableName, columnName)
}

// Driver returns the current database driver instance.
func (db *DB) Driver() SQLDriver {
	return db.driver
}

// TableName returns the current model table name.
func (db *DB) TableName() string {
	if db.model != nil {
		return db.model.name
	}
	return ""
}

// Columns returns the current model column names.
func (db *DB) Columns() []string {
	columns := []string{}

	if db.model != nil {
		for _, field := range db.model.fields {
			columns = append(columns, field.name)
		}
	}

	return columns
}

// Begin starts a new database transaction.
func (db *DB) Begin() *DB {
	tx, err := db.sqlDB.Begin()
	if err != nil {
		return db.setError(err)
	}

	dbClone := db.clone()
	dbClone.tx = tx

	return dbClone
}

// Commit commits a database transaction.
func (db *DB) Commit() *DB {
	if db.tx == nil {
		return db.setError(ErrMissingTransaction)
	}

	err := db.tx.Commit()
	if err != nil {
		return db.setError(err)
	}

	return db
}

// Rollback rollbacks a database transaction.
func (db *DB) Rollback() *DB {
	if db.tx == nil {
		return db.setError(ErrMissingTransaction)
	}

	err := db.tx.Rollback()
	if err != nil {
		return db.setError(err)
	}

	return db
}

// Raw sets the current query raw SQL statement.
func (db *DB) Raw(query string, args ...interface{}) *DB {
	db.query.Raw(query, args...)
	return db
}

// Row executes the query returning a single SQL row.
func (db *DB) Row() *sql.Row {
	return db.query.Row()
}

// Rows executes the query returning multiple rows.
func (db *DB) Rows() (*sql.Rows, error) {
	return db.query.Rows()
}

// Result exexutes the query without returning any row.
func (db *DB) Result() (sql.Result, error) {
	return db.query.Result()
}

// RowsAffected returns the number of rows affected by the latest query.
func (db *DB) RowsAffected() int64 {
	if db.result == nil {
		return 0
	}

	count, _ := db.result.RowsAffected()
	return count
}

// Count sets the current query to number of rows count.
func (db *DB) Count(value interface{}) *DB {
	if err := db.query.Count().Row().Scan(value); err != nil {
		return db.setError(err)
	}

	return db
}

// Find finds records in the database matching the current query conditions.
func (db *DB) Find(value interface{}) *DB {
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	dbClone := db.From(value)
	if dbClone.Error() != nil {
		return dbClone
	}

	rows, err := dbClone.Rows()
	if err != nil {
		return dbClone.setError(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		dbClone.Scan(rows, dbClone.value)
		count++
	}

	if count == 0 {
		dbClone.setError(sql.ErrNoRows)
	}

	return dbClone
}

// Select sets the current query columns list.
func (db *DB) Select(columns ...string) *DB {
	db.query.Select(columns...)
	return db
}

// From sets the model or table associated to the current query.
func (db *DB) From(value interface{}) *DB {
	var err error

	dbClone := db.clone()
	dbClone.value = value

	if v, ok := value.(string); ok {
		dbClone.query.From(v)
	} else {
		if dbClone.model, err = newModel(value, dbClone); err != nil {
			return dbClone.setError(err)
		}

		dbClone.query.From(dbClone.model.name)
	}

	return dbClone
}

// Where registers a new condition in the current query.
func (db *DB) Where(condition string, args ...interface{}) *DB {
	db.query.Where(condition, args...)
	return db
}

// Offset sets the offset of the current query.
func (db *DB) Offset(offset int) *DB {
	db.query.Offset(offset)
	return db
}

// Limit sets the maximum number fo rows to return for the current query.
func (db *DB) Limit(limit int) *DB {
	db.query.Limit(limit)
	return db
}

// OrderBy sets the current query rows ordering terms.
func (db *DB) OrderBy(terms ...string) *DB {
	db.query.OrderBy(terms...)
	return db
}

// Save inserts or updates a record into the database.
func (db *DB) Save(value interface{}) *DB {
	var update bool

	dbClone := db.From(value)
	if dbClone.Error() != nil {
		return dbClone
	}

	primaryField := dbClone.model.primaryField()
	if primaryField != nil {
		// Update existing value from database if already exists
		if !IsDefault(primaryField.value) {
			var count int

			dbClone.Where(fmt.Sprintf("%s = ?", dbClone.driver.QuoteName(primaryField.name)),
				primaryField.value.Interface()).Count(&count)

			update = count > 0
		}
	}

	// Insert or update value in database
	if update {
		return dbClone.update()
	}

	return dbClone.insert()
}

// Delete deletes a record from the database.
func (db *DB) Delete(value interface{}) *DB {
	dbClone := db.From(value)
	if dbClone.Error() != nil {
		return dbClone
	}

	return dbClone.delete()
}

// Migrate handles automatic tables and columns migration for given models. To be safe, it won't delete or alter
// existing columns.
func (db *DB) Migrate(values ...interface{}) *DB {
	// Register new models for migration
	for _, value := range values {
		if err := db.From(value).migrate().Error(); err != nil {
			return db.setError(err)
		}
	}

	return db
}

// DropTable drops tables from the database for given models or table names.
func (db *DB) DropTable(values ...interface{}) *DB {
	// Register new models for migration
	for _, value := range values {
		if err := db.From(value).dropTable().Error(); err != nil {
			db.setError(err)
			break
		}
	}

	return db
}

func (db *DB) quiet() *DB {
	db.query.quiet()
	return db
}

func (db *DB) setError(err error) *DB {
	if err != nil {
		db.err = err
	}
	return db
}

func (db *DB) clone() *DB {
	dbClone := &DB{
		logger:   db.logger,
		driver:   db.driver,
		sqlDB:    db.sqlDB,
		models:   db.models,
		migrated: db.migrated,
		query:    &query{},
		tx:       db.tx,
	}

	*dbClone.query = *db.query
	dbClone.query.db = dbClone

	return dbClone
}
