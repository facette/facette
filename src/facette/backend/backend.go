package backend

import (
	"regexp"

	"github.com/facette/logger"
	"github.com/facette/maputil"
	"github.com/facette/sqlstorage"
)

const (
	// FilterGlobPrefix is the glob pattern matching filter prefix.
	FilterGlobPrefix = "glob:"
	// FilterRegexpPrefix is the regular expression matching filter prefix.
	FilterRegexpPrefix = "regexp:"
)

var (
	nameRegexp = regexp.MustCompile("(?i)^[a-z0-9](?:[a-z0-9\\-_\\.]*[a-z0-9])*$")
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
		&CollectionEntry{},
		&Collection{},
	); err != nil {
		return nil, err
	}

	storage.
		AddForeignKey(&Graph{}, "link", "graphs(id)", "CASCADE", "CASCADE").
		AddForeignKey(&CollectionEntry{}, "collection", "collections(id)", "CASCADE", "CASCADE").
		AddForeignKey(&CollectionEntry{}, "graph", "graphs(id)", "CASCADE", "CASCADE").
		AddForeignKey(&Collection{}, "link", "collections(id)", "CASCADE", "CASCADE").
		AddForeignKey(&Collection{}, "parent", "collections(id)", "SET NULL", "SET NULL").
		Association(&Collection{}, "Entries")

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
