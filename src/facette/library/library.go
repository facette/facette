package library

import (
	"facette/backend"
	"facette/config"
	"facette/utils"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	// LibraryItemSourceGroup represents a source group item.
	LibraryItemSourceGroup = iota
	// LibraryItemMetricGroup represents a metric group item.
	LibraryItemMetricGroup
	// LibraryItemGraph represents a graph item.
	LibraryItemGraph
	// LibraryItemGraphTemplate represents a graph template item.
	LibraryItemGraphTemplate
	// LibraryItemCollection represents a collection item.
	LibraryItemCollection
)

const (
	// UUIDPattern represents an UUID validation pattern.
	UUIDPattern = "^\\d{8}-(?:\\d{4}-){3}\\d{12}$"
)

// Library represents the main structure of running Facette's instance library (e.g. sources and metrics groups,
// graphs, collections).
type Library struct {
	Config         *config.Config
	Catalog        *backend.Catalog
	Groups         map[string]*Group
	Graphs         map[string]*Graph
	TemplateGraphs map[string]*Graph
	Collections    map[string]*Collection
	debugLevel     int
	idRegexp       *regexp.Regexp
}

// Update updates the current Library by browsing the filesystem for stored data.
func (library *Library) Update() error {
	var (
		err      error
		itemType int
		walkFunc func(filePath string, fileInfo os.FileInfo, err error) error
	)

	// Empty library maps
	library.Groups = make(map[string]*Group)
	library.Graphs = make(map[string]*Graph)
	library.TemplateGraphs = make(map[string]*Graph)
	library.Collections = make(map[string]*Collection)

	walkFunc = func(filePath string, fileInfo os.FileInfo, fileError error) error {
		var (
			itemID string
			mode   os.FileMode
		)

		mode = fileInfo.Mode() & os.ModeType
		if mode != 0 || !strings.HasSuffix(filePath, ".json") {
			return nil
		}

		_, itemID = path.Split(filePath[:len(filePath)-5])

		if library.debugLevel > 1 {
			log.Printf("DEBUG: loading `%s' item from `%s' file...", itemID, filePath)
		}

		return library.LoadItem(itemID, itemType)
	}

	log.Println("INFO: library update started")

	for _, itemType = range []int{
		LibraryItemSourceGroup,
		LibraryItemMetricGroup,
		LibraryItemGraph,
		LibraryItemCollection,
	} {
		if err = utils.WalkDir(library.getDirPath(itemType), walkFunc); err != nil {
			log.Println("ERROR: " + err.Error())
		}
	}

	log.Println("INFO: library update completed")

	return nil
}

// NewLibrary creates a new instance of Library.
func NewLibrary(config *config.Config, catalog *backend.Catalog, debugLevel int) *Library {
	var (
		library *Library
	)

	// Create new Library instance
	library = &Library{Config: config, Catalog: catalog, debugLevel: debugLevel}

	// Compile ID validation regexp
	library.idRegexp = regexp.MustCompile(UUIDPattern)

	return library
}
