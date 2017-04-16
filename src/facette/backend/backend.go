package backend

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"facette/mapper"
	"facette/orm"

	"github.com/facette/logger"
	"github.com/facette/sliceutil"
)

const (
	// FilterGlobPrefix is the glob pattern matching filter prefix.
	FilterGlobPrefix = "glob:"
	// FilterRegexpPrefix is the regular expression matching filter prefix.
	FilterRegexpPrefix = "regexp:"
)

var (
	authorizedAliasChars *regexp.Regexp
)

// Backend represents a backend instance.
type Backend struct {
	config *mapper.Map
	log    *logger.Logger
	db     *orm.DB
}

func init() {
	authorizedAliasChars = regexp.MustCompile("^[A-Za-z0-9\\-_]+$")
}

// NewBackend creates a new instance of a backend.
func NewBackend(settings *mapper.Map, log *logger.Logger) (*Backend, error) {
	if settings == nil {
		return nil, ErrMissingBackendConfig
	}

	// Open database connection
	db, err := orm.Open(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %s", err)
	}

	// TODO: make it configurable
	db.SetLogger(log.Logger(logger.LevelDebug))

	// Initialize database schema
	if err := db.Migrate(
		Provider{},
		SourceGroup{},
		MetricGroup{},
		Graph{},
		Collection{},
	).Error(); err != nil {
		return nil, err
	}

	return &Backend{
		log: log,
		db:  db,
	}, nil
}

// Close closes the backend database connection.
func (b *Backend) Close() error {
	return b.db.Close()
}

// Add inserts or updates an item into the backend database.
func (b *Backend) Add(v interface{}) error {
	if val, ok := v.(Validator); ok {
		if err := val.Validate(b); err != nil {
			return err
		}
	}

	err := b.db.Save(v).Error()

	switch err {
	case orm.ErrConstraintUnique:
		err = ErrResourceConflict

	case orm.ErrConstraintForeignKey:
		err = ErrResourceMissingDependency

	case orm.ErrConstraintNotNull:
		err = ErrResourceMissingData
	}

	return err
}

// Delete deletes an existing item from the backend database.
func (b *Backend) Delete(v interface{}) error {
	if count := b.db.Delete(v).RowsAffected(); count == 0 {
		return ErrItemNotExist
	}
	return nil
}

// Reset deletes all resources of a given type from the backend database.
func (b *Backend) Reset(v interface{}) error {
	tx := b.db.Begin()

	_, err := tx.Raw("DELETE FROM " + b.db.Driver().QuoteName(b.db.From(v).TableName())).Result()
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return err
}

// Get fetches an item from the backend database.
func (b *Backend) Get(term string, v interface{}) error {
	// This method accepts either a backend item ID (UUID) or an 'alias' for some items types,
	// such as graphs or collections.

	rt := reflect.TypeOf(v)
	rv := reflect.New(reflect.MakeSlice(reflect.SliceOf(rt), 0, 0).Type())

	tx := b.db.Begin()
	defer tx.Commit()

	if b.IsAliasable(v) {
		tx.Where("id = ? OR alias = ?", term, term)
	} else {
		tx.Where("id = ?", term)
	}

	if err := tx.Find(rv.Interface()).Error(); err == sql.ErrNoRows {
		return ErrItemNotExist
	} else if err != nil {
		return err
	}

	ri := reflect.ValueOf(v).Elem()
	if reflect.Indirect(rv).Len() == 2 {
		// In case of a backend search returning 2 results
		// (i.e. one with ID='<UUID>' and another with Alias='<UUID>'), the record with ID='<UUID>'
		// has the precedence.

		item := reflect.Indirect(rv).Index(0)
		if reflect.Indirect(item).FieldByName("ID").String() == term {
			ri.Set(item.Elem())
		} else {
			ri.Set(reflect.Indirect(rv).Index(1).Elem())
		}
	} else {
		ri.Set(reflect.Indirect(rv).Index(0).Elem())
	}

	return nil
}

// List lists the existing items from the backend database.
func (b *Backend) List(v interface{}, filters map[string]interface{}, sort []string, offset, limit int) (int, error) {
	// Check for slice element type
	rv := reflect.ValueOf(v)
	if reflect.Indirect(rv).Kind() != reflect.Slice {
		return 0, ErrInvalidSlice
	}

	rv = reflect.New(rv.Type().Elem().Elem())

	tx := b.db.Begin()
	defer tx.Commit()

	filter, args := b.buildFilter(filters)

	// Count for total number of items
	db := tx.From(rv.Interface())
	if filter != "" {
		db.Where(filter, args...)
	}

	count := 0
	if err := db.Count(&count).Error(); err != nil {
		return 0, err
	}

	// Query items
	if filter != "" {
		db.Where(filter, args...)
	}

	for _, s := range sort {
		if strings.HasPrefix(s, "-") {
			if !sliceutil.Has(db.Columns(), s[1:]) {
				return 0, ErrUnknownColumn
			}

			s = s[1:] + " DESC"
		} else if !sliceutil.Has(db.Columns(), s) {
			return 0, ErrUnknownColumn
		}

		db.OrderBy(s)
	}

	if limit > 0 {
		db.Offset(offset).Limit(limit)
	}

	err := db.Find(v).Error()
	if err == sql.ErrNoRows {
		err = nil
	}

	return count, err
}

// Search searches for items in the backend database given a list of filters.
func (b *Backend) Search(types []interface{}, filters map[string]interface{}, sort []string,
	offset, limit int) ([]TypedItem, int, error) {

	// Fecth common columns names
	columns := []string{"? AS type"}
	columns = append(columns, b.db.From(Item{}).Columns()...)

	// Build sub-queries list
	qSelect := "SELECT " + strings.Join(columns, ", ")
	qWhere := ""
	qArgs := []interface{}{}

	filter, args := b.buildFilter(filters)
	if filter != "" {
		qWhere = " WHERE " + filter
	}

	queries := []string{}
	for _, typ := range types {
		table := b.db.From(typ).TableName()
		queries = append(queries, qSelect+" FROM "+table+qWhere)
		qArgs = append(qArgs, table)
		qArgs = append(qArgs, args...)
	}

	// Prepare query
	query := strings.Join(queries, " UNION ALL ")

	qOrder := ""
	for i, s := range sort {
		if i == 0 {
			qOrder += " ORDER BY "
		} else {
			qOrder += ", "
		}

		if strings.HasPrefix(s, "-") {
			qOrder += s[1:] + " DESC"
		} else {
			qOrder += s
		}
	}

	qLimit := ""
	if limit > 0 {
		qLimit = " " + b.db.Driver().LimitClause(offset, limit)
	}

	// Fetch search results
	tx := b.db.Begin()
	defer tx.Commit()

	q := tx.Raw(query+qOrder+qLimit, qArgs...)

	rows, err := q.Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := []TypedItem{}
	for rows.Next() {
		var item TypedItem
		if err := q.Scan(rows, &item).Error(); err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}

	// Get total count
	count := 0
	if err := tx.Raw("SELECT COUNT(*) FROM ("+query+")", qArgs...).Count(&count).Error(); err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Count returns the existing items count from the backend database.
func (b *Backend) Count(v interface{}, filters map[string]interface{}) (int, error) {
	count := 0
	return count, b.db.From(v).Count(&count).Error()
}

// IsAliasable returns whether or not the argument is an aliasable backend item.
func (b *Backend) IsAliasable(v interface{}) bool {
	return reflect.Indirect(reflect.ValueOf(v)).FieldByName("Alias").IsValid()
}

func (b *Backend) buildFilter(filters map[string]interface{}) (string, []interface{}) {
	parts := []string{}
	args := []interface{}{}

	for k, v := range filters {
		if s, ok := v.(string); ok {
			if strings.HasPrefix(s, FilterGlobPrefix) {
				v = orm.GlobModifier(strings.TrimPrefix(s, FilterGlobPrefix))
			} else if strings.HasPrefix(s, FilterRegexpPrefix) {
				v = orm.RegexpModifier(strings.TrimPrefix(s, FilterRegexpPrefix))
			}
		}

		clause, value := b.db.Driver().WhereClause(k, v)
		parts = append(parts, clause)
		args = append(args, value)
	}

	if len(parts) == 0 {
		return "", args
	}

	return strings.Join(parts, " AND "), args
}
