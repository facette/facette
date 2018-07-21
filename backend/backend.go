package backend

import (
	"regexp"

	"facette.io/logger"
	"facette.io/maputil"
	"facette.io/sqlstorage"
)

var (
	nameRegexp = regexp.MustCompile("(?i)^[a-z0-9](?:[a-z0-9\\-_\\.]*[a-z0-9])?$")
)

// Backend represents a back-end instance.
type Backend struct {
	config  *maputil.Map
	log     *logger.Logger
	storage *sqlstorage.Storage
}

// NewBackend creates a new back-end instance.
func NewBackend(config *maputil.Map, log *logger.Logger) (*Backend, error) {
	// Initialize storage
	storage, err := sqlstorage.NewStorage("facette", config, log)
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

	return &Backend{
		config:  config,
		log:     log,
		storage: storage,
	}, nil
}

// Close closes the back-end storage.
func (b *Backend) Close() error {
	return b.storage.Close()
}

// Storage returns the back-end storage instance.
func (b *Backend) Storage() *sqlstorage.Storage {
	return b.storage
}
