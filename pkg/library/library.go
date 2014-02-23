package library

import (
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/utils"
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
	Catalog        *catalog.Catalog
	Groups         map[string]*Group
	Graphs         map[string]*Graph
	TemplateGraphs map[string]*Graph
	Collections    map[string]*Collection
	debugLevel     int
	idRegexp       *regexp.Regexp
}

// Update updates the current Library by browsing the filesystem for stored data.
func (library *Library) Update() error {
	var itemType int

	// Empty library maps
	library.Groups = make(map[string]*Group)
	library.Graphs = make(map[string]*Graph)
	library.TemplateGraphs = make(map[string]*Graph)
	library.Collections = make(map[string]*Collection)

	walkFunc := func(filePath string, fileInfo os.FileInfo, fileError error) error {
		mode := fileInfo.Mode() & os.ModeType
		if mode != 0 || !strings.HasSuffix(filePath, ".json") {
			return nil
		}

		_, itemID := path.Split(filePath[:len(filePath)-5])

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
		dirPath := library.getDirPath(itemType)

		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		if err := utils.WalkDir(dirPath, walkFunc); err != nil {
			log.Println("ERROR: " + err.Error())
		}
	}

	// Update collection items parent-children relations
	for _, collection := range library.Collections {
		if collection.ParentID == "" {
			continue
		}

		if _, ok := library.Collections[collection.ParentID]; !ok {
			log.Println("ERROR: unknown `%s' parent identifier", collection.ParentID)
			continue
		}

		collection.Parent = library.Collections[collection.ParentID]
		collection.Parent.Children = append(collection.Parent.Children, collection)
	}

	log.Println("INFO: library update completed")

	return nil
}

// NewLibrary creates a new instance of Library.
func NewLibrary(config *config.Config, catalog *catalog.Catalog, debugLevel int) *Library {
	// Create new Library instance
	library := &Library{Config: config, Catalog: catalog, debugLevel: debugLevel}

	// Compile ID validation regexp
	library.idRegexp = regexp.MustCompile(UUIDPattern)

	return library
}
