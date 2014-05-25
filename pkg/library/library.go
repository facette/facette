// Package library implements the service library handling user-defined items (e.g. sources and metrics groups, graphs,
// collections).
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
	_ = iota
	// LibraryItemSourceGroup represents a source group item.
	LibraryItemSourceGroup
	// LibraryItemMetricGroup represents a metric group item.
	LibraryItemMetricGroup
	// LibraryItemScale represents a scale item.
	LibraryItemScale
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

// Library represents the main structure of library instance.
type Library struct {
	Config         *config.Config
	Catalog        *catalog.Catalog
	Groups         map[string]*Group
	Scales         map[string]*Scale
	Graphs         map[string]*Graph
	TemplateGraphs map[string]*Graph
	Collections    map[string]*Collection
	debugLevel     int
	idRegexp       *regexp.Regexp
}

// Refresh updates the current library by browsing the filesystem for stored data.
func (library *Library) Refresh() error {
	var itemType int

	// Empty library maps
	library.Groups = make(map[string]*Group)
	library.Scales = make(map[string]*Scale)
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

	log.Println("INFO: library refresh started")

	for _, itemType = range []int{
		LibraryItemSourceGroup,
		LibraryItemMetricGroup,
		LibraryItemScale,
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

	log.Println("INFO: library refresh completed")

	return nil
}

// NewLibrary creates a new instance of library.
func NewLibrary(config *config.Config, catalog *catalog.Catalog, debugLevel int) *Library {
	return &Library{
		Config:     config,
		Catalog:    catalog,
		debugLevel: debugLevel,
		idRegexp:   regexp.MustCompile(UUIDPattern),
	}
}
