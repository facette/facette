package storage

import (
	"regexp"

	"facette.io/logger"
	"facette.io/maputil"
	"facette.io/sqlstorage"
)

var (
	nameRegexp = regexp.MustCompile(`(?i)^[a-z0-9](?:[a-z0-9\-_\.]*[a-z0-9])?$`)
)

// Storage represents a storage instance.
type Storage struct {
	config  *maputil.Map
	logger  *logger.Logger
	storage *sqlstorage.Storage
}

// New creates a new storage instance.
func New(config *maputil.Map, logger *logger.Logger) (*Storage, error) {
	// Initialize storage
	storage, err := sqlstorage.NewStorage("facette", config, logger)
	if err != nil {
		return nil, err
	}

	// Initialize database schema
	if err := storage.Migrate(
		&Provider{},
		&SourceGroup{},
		&MetricGroup{},
		&Graph{},
		&Collection{},
		&CollectionEntry{},
	); err != nil {
		return nil, err
	}

	// If driver is 'mysql', handle foreign separately as MySQL parses but ignores inlined in column definitions
	// (see https://dev.mysql.com/doc/refman/5.7/en/create-table-foreign-keys.html)
	if driver, err := config.GetString("driver", ""); err == nil && driver == "mysql" {
		storage.
			AddForeignKey(&Graph{}, "link", "graphs(id)", "CASCADE", "CASCADE").
			AddForeignKey(&CollectionEntry{}, "collection", "collections(id)", "CASCADE", "CASCADE").
			AddForeignKey(&CollectionEntry{}, "graph", "graphs(id)", "CASCADE", "CASCADE").
			AddForeignKey(&Collection{}, "link", "collections(id)", "CASCADE", "CASCADE").
			AddForeignKey(&Collection{}, "parent", "collections(id)", "SET NULL", "SET NULL")
	}

	storage.Association(&Collection{}, "Entries")

	return &Storage{
		config:  config,
		logger:  logger,
		storage: storage,
	}, nil
}

// Close closes the storage.
func (s *Storage) Close() error {
	return s.storage.Close()
}

// SQL returns the storage underlying SQL storage instance.
func (s *Storage) SQL() *sqlstorage.Storage {
	return s.storage
}
