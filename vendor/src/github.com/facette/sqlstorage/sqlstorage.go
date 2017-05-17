package sqlstorage

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/facette/logger"
	"github.com/facette/maputil"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Storage represents a storage instance.
type Storage struct {
	name         string
	config       *maputil.Map
	log          *logger.Logger
	driver       sqlDriver
	db           *gorm.DB
	associations map[reflect.Type][]string
}

// NewStorage creates a new storage instance.
func NewStorage(name string, config *maputil.Map, log *logger.Logger) (*Storage, error) {
	// Open database connection
	driver, err := newSQLDriver(name, config)
	if err != nil {
		return nil, err
	}

	sqlDB, err := driver.Open()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	db, err := gorm.Open(driver.Name(), sqlDB)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	// Set ORM logger and debugging mode
	db.SetLogger(&ormLogger{
		log.Context(log.CurrentContext() + "[sql]").Logger(logger.LevelDebug),
	})

	if debug, _ := config.GetBool("debug", false); debug {
		db.LogMode(true)
	}

	// Execute driver-specific commands
	if err := driver.Init(db); err != nil {
		return nil, errors.Wrap(err, "failed to initialize driver")
	}

	return &Storage{
		name:         name,
		config:       config,
		log:          log,
		driver:       driver,
		db:           db,
		associations: make(map[reflect.Type][]string),
	}, nil
}

// Close closes the storage database connection.
func (s *Storage) Close() error {
	return s.db.Close()
}

// Migrate handles automatic migration of given item database models.
func (s *Storage) Migrate(v ...interface{}) error {
	return s.db.AutoMigrate(v...).Error
}

// AddForeignKey defines a new foreign key for a given item.
func (s *Storage) AddForeignKey(v interface{}, field, dest, onDelete, onUpdate string) *Storage {
	s.db.Model(v).AddForeignKey(field, dest, onDelete, onUpdate)
	return s
}

// Association registers a new item association field.
func (s *Storage) Association(v interface{}, fields ...string) *Storage {
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if _, ok := s.associations[rt]; !ok {
		s.associations[rt] = []string{}
	}
	s.associations[rt] = append(s.associations[rt], fields...)

	return s
}

// Save stores a new item or modifies an existing item into the storage.
func (s *Storage) Save(v interface{}) error {
	var err error

	tx := s.db.Begin()
	defer tx.Commit()

	// Delete previous associations
	if err := s.handleAssociations(tx, v, true); err != nil {
		return err
	}

	// Check if item with the given primary field already exists
	rv := reflect.ValueOf(v)
	scope := tx.NewScope(v)
	field := scope.PrimaryField()

	if tx.First(
		reflect.New(rv.Type().Elem()).Interface(),
		fmt.Sprintf("%v = ?", scope.Quote(field.DBName)),
		reflect.Indirect(rv).FieldByName(field.Name).Interface(),
	).RecordNotFound() {
		err = tx.Create(v).Error
	} else {
		err = tx.Save(v).Error
	}

	if err != nil {
		return s.driver.NormalizeError(err)
	}

	return nil
}

// Get retrieves an existing item from the storage.
func (s *Storage) Get(column string, values interface{}, v interface{}) error {
	tx := s.db.Begin()
	defer tx.Commit()

	// Retrieve item from database
	whereClause := tx.Dialect().Quote(column)
	if reflect.TypeOf(values).Kind() == reflect.Slice {
		whereClause += " IN (?)"
	} else {
		whereClause += " = ?"
	}

	if err := tx.Where(whereClause, values).Find(v).Error; err == gorm.ErrRecordNotFound {
		return ErrItemNotFound
	} else if err != nil {
		return s.driver.NormalizeError(err)
	}

	return s.handleAssociations(tx, v, false)
}

// Delete removes an existing item from the storage.
func (s *Storage) Delete(v interface{}) error {
	tx := s.db.Begin()
	defer tx.Commit()

	// Delete item from database
	if count := tx.Delete(v).RowsAffected; count == 0 {
		return ErrItemNotFound
	}

	return nil
}

// Count returns the count of existing items from the storage.
func (s *Storage) Count(v interface{}) (int, error) {
	tx := s.db.Begin()
	defer tx.Commit()

	count := 0
	if err := tx.Model(v).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// List retrieves a list of existing items from the storage.
func (s *Storage) List(v interface{}, filters map[string]interface{}, sort []string, offset, limit int) (int, error) {
	tx := s.db.Begin()
	defer tx.Commit()

	tx = tx.Model(v)

	for k, v := range filters {
		k, v = s.driver.WhereClause(k, v)
		tx = tx.Where(k, v)
	}

	count := 0
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}

	for _, field := range sort {
		var desc bool

		if strings.HasPrefix(field, "-") {
			field = field[1:]
			desc = true
		}

		if !tx.Dialect().HasColumn(tx.NewScope(v).TableName(), field) {
			return 0, ErrUnknownColumn
		}

		if desc {
			tx = tx.Order(field + " DESC")
		} else {
			tx = tx.Order(field)
		}
	}

	if limit > 0 {
		tx = tx.Offset(offset).Limit(limit)
	}

	if err := tx.Find(v).Error; err != nil {
		return 0, s.driver.NormalizeError(err)
	}

	// Retrieve item-specific associations
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rt := rv.Type().Elem()
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if _, ok := s.associations[rt]; ok {
		// Reset filters, orders and limits
		tx = tx.New()

		for i, n := 0, rv.Len(); i < n; i++ {
			if err := s.handleAssociations(tx, rv.Index(i).Interface(), false); err != nil {
				return 0, err
			}
		}
	}

	return count, nil
}

// Search searches for existing items given multiple types in the storage.
func (s *Storage) Search(values []interface{}, v interface{}, filters map[string]interface{}, sort []string, offset,
	limit int) (int, error) {

	tx := s.db.Begin()
	defer tx.Commit()

	// Get columns list
	scope := tx.NewScope(reflect.Indirect(reflect.New(reflect.TypeOf(v).Elem())).Interface())

	columns := []string{fmt.Sprintf("? AS type")}
	for _, field := range scope.Fields() {
		if !field.IsIgnored {
			columns = append(columns, field.DBName)
		}
	}

	tx = tx.Select(columns)

	// Generate sub-queries
	queries := []string{}
	args := []interface{}{}

	for _, v := range values {
		scope := tx.NewScope(v)
		args = append(args, scope.TableName())

		for k, v := range filters {
			if scope.Dialect().HasColumn(scope.TableName(), k) {
				k, v = s.driver.WhereClause(k, v)
				scope.Search.Where(k, v)
				args = append(args, v)
			}
		}

		queries = append(queries, fmt.Sprintf(
			"SELECT %s FROM %s %s",
			strings.Join(scope.SelectAttrs(), ", "),
			scope.QuotedTableName(),
			scope.CombinedConditionSql(),
		))
	}

	tx1 := tx.Raw(fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS a", strings.Join(queries, " UNION ALL ")), args...)

	count := 0
	if err := tx1.Count(&count).Error; err != nil {
		return 0, err
	}

	scope = tx.NewScope(nil)

	for _, field := range sort {
		var desc bool

		if strings.HasPrefix(field, "-") {
			field = field[1:]
			desc = true
		}

		if desc {
			scope.Search.Order(field + " DESC")
		} else {
			scope.Search.Order(field)
		}
	}

	if limit > 0 {
		tx = tx.Offset(offset).Limit(limit)
	}

	tx2 := tx.Raw(strings.Join(queries, " UNION ALL ")+scope.CombinedConditionSql(), args...)

	if err := tx2.Scan(v).Error; err != nil {
		return 0, s.driver.NormalizeError(err)
	}

	return count, nil
}

func (s *Storage) handleAssociations(tx *gorm.DB, v interface{}, delete bool) error {
	rv := reflect.ValueOf(v)

	rt := rv.Type()
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if fieldNames, ok := s.associations[rt]; ok {
		tx = tx.New()

		for _, name := range fieldNames {
			fv := reflect.Indirect(rv).FieldByName(name).Addr()

			if delete {
				scope := tx.NewScope(v)
				if field, ok := scope.FieldByName(name); ok {
					for idx, foreignKey := range field.Relationship.ForeignDBNames {
						assocName := field.Relationship.AssociationForeignFieldNames[idx]
						if assocField, ok := scope.FieldByName(assocName); ok {
							tx = tx.Where(fmt.Sprintf("%v = ?", scope.Quote(foreignKey)), assocField.Field.Interface())
						}
					}

					if err := tx.Delete(fv.Interface()).Error; err != nil {
						return s.driver.NormalizeError(err)
					}
				}

				continue
			}

			if err := tx.Model(v).Association(name).Find(fv.Interface()).Error; err != nil {
				return s.driver.NormalizeError(err)
			}
		}
	}

	return nil
}
