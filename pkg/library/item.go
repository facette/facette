package library

import (
	"fmt"
	"log"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
	"github.com/facette/facette/thirdparty/github.com/nu7hatch/gouuid"
)

// Item represents the base structure of a library item.
type Item struct {
	path        string
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Modified    time.Time `json:"-"`
}

// GetItem returns the base structure of a library item.
func (item *Item) GetItem() *Item {
	return item
}

// DeleteItem removes an existing item from the library.
func (library *Library) DeleteItem(id string, itemType int) error {
	if !library.ItemExists(id, itemType) {
		return os.ErrNotExist
	}

	// Delete sub-collections
	if itemType == LibraryItemCollection {
		for _, child := range library.Collections[id].Children {
			library.DeleteItem(child.ID, LibraryItemCollection)
		}
	}

	// Remove stored JSON
	if itemType != LibraryItemGraph || itemType == LibraryItemGraph && !library.Graphs[id].Volatile {
		if err := syscall.Unlink(library.getFilePath(id, itemType)); err != nil {
			return err
		}
	}

	// Delete item from library
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		delete(library.Groups, id)

	case LibraryItemGraph:
		delete(library.Graphs, id)

	case LibraryItemCollection:
		delete(library.Collections, id)
	}

	return nil
}

// GetItem gets an item from the library by its identifier.
func (library *Library) GetItem(id string, itemType int) (interface{}, error) {
	if !library.ItemExists(id, itemType) {
		return nil, os.ErrNotExist
	}

	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		return library.Groups[id], nil

	case LibraryItemGraph:
		item := library.Graphs[id]

		if library.Graphs[id].Volatile {
			delete(library.Graphs, id)
		}

		return item, nil

	case LibraryItemCollection:
		return library.Collections[id], nil
	}

	return nil, fmt.Errorf("no item found")
}

// GetItemByName gets an item from the library by its name.
func (library *Library) GetItemByName(name string, itemType int) (interface{}, error) {
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		for _, item := range library.Groups {
			if item.Type != itemType || item.Name != name {
				continue
			}

			return item, nil
		}

	case LibraryItemGraph:
		for _, item := range library.Graphs {
			if item.Name != name {
				continue
			}

			return item, nil
		}

	case LibraryItemCollection:
		for _, item := range library.Collections {
			if item.Name != name {
				continue
			}

			return item, nil
		}
	}

	return nil, os.ErrNotExist
}

// ItemExists returns whether an item exists the library or not.
func (library *Library) ItemExists(id string, itemType int) bool {
	exists := false

	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		if _, ok := library.Groups[id]; ok && library.Groups[id].Type == itemType {
			exists = true
		}

	case LibraryItemGraph:
		_, exists = library.Graphs[id]

	case LibraryItemGraphTemplate:
		_, exists = library.TemplateGraphs[id]

	case LibraryItemCollection:
		_, exists = library.Collections[id]
	}

	return exists
}

// LoadItem loads an item by its identifier.
func (library *Library) LoadItem(id string, itemType int) error {
	// Load item from file
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		tmpGroup := &Group{}

		filePath := library.getFilePath(id, itemType)

		fileInfo, err := utils.JSONLoad(filePath, &tmpGroup)
		if err != nil {
			return fmt.Errorf("in %s, %s", filePath, err.Error())
		}

		library.Groups[id] = tmpGroup
		library.Groups[id].Type = itemType
		library.Groups[id].Modified = fileInfo.ModTime()

	case LibraryItemGraph:
		tmpGraph := &Graph{}

		filePath := library.getFilePath(id, itemType)

		fileInfo, err := utils.JSONLoad(filePath, &tmpGraph)
		if err != nil {
			return fmt.Errorf("in %s, %s", filePath, err.Error())
		}

		library.Graphs[id] = tmpGraph
		library.Graphs[id].Volatile = false
		library.Graphs[id].Modified = fileInfo.ModTime()

	case LibraryItemCollection:
		var tmpCollection *struct {
			*Collection
			Parent string `json:"parent"`
		}

		filePath := library.getFilePath(id, LibraryItemCollection)

		fileInfo, err := utils.JSONLoad(filePath, &tmpCollection)
		if err != nil {
			return fmt.Errorf("in %s, %s", filePath, err.Error())
		}

		if !library.ItemExists(id, LibraryItemCollection) {
			library.Collections[id] = &Collection{}
		}

		*library.Collections[id] = *tmpCollection.Collection

		if tmpCollection.Parent != "" {
			library.Collections[id].ParentID = tmpCollection.Parent
		}

		library.Collections[id].Modified = fileInfo.ModTime()
	}

	return nil
}

// StoreItem stores an item into the library.
func (library *Library) StoreItem(item interface{}, itemType int) error {
	var itemStruct *Item

	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		itemStruct = item.(*Group).GetItem()

	case LibraryItemGraph:
		itemStruct = item.(*Graph).GetItem()

	case LibraryItemCollection:
		itemStruct = item.(*Collection).GetItem()
	}

	if itemStruct.ID == "" {
		uuidTemp, err := uuid.NewV4()
		if err != nil {
			return err
		}

		itemStruct.ID = uuidTemp.String()
	} else if !library.ItemExists(itemStruct.ID, itemType) {
		return os.ErrNotExist
	}

	// Check for name field presence/duplicates
	if itemStruct.Name == "" && (itemType != LibraryItemGraph ||
		itemType == LibraryItemGraph && !item.(*Graph).Volatile) {
		return os.ErrInvalid
	}

	itemTemp, err := library.GetItemByName(itemStruct.Name, itemType)
	if err == nil {
		switch itemType {
		case LibraryItemSourceGroup, LibraryItemMetricGroup:
			if itemTemp.(*Group).ID != itemStruct.ID {
				log.Printf("ERROR: duplicate `%s' group identifier", itemStruct.ID)
				return os.ErrExist
			}

		case LibraryItemGraph:
			if !item.(*Graph).Volatile && itemTemp.(*Graph).ID != itemStruct.ID {
				log.Printf("ERROR: duplicate `%s' graph identifier", itemStruct.ID)
				return os.ErrExist
			}

		case LibraryItemCollection:
			if itemTemp.(*Collection).ID != itemStruct.ID {
				log.Printf("ERROR: duplicate `%s' collection identifier", itemStruct.ID)
				return os.ErrExist
			}
		}
	}

	// Store item into library
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		library.Groups[itemStruct.ID] = item.(*Group)
		library.Groups[itemStruct.ID].ID = itemStruct.ID

	case LibraryItemGraph:
		// Check for definition names duplicates
		stackSet := set.New()
		groupSet := set.New()
		serieSet := set.New()

		for _, stack := range item.(*Graph).Stacks {
			if stack == nil {
				log.Println("ERROR: stack is null")
				return os.ErrInvalid
			} else if stackSet.Has(stack.Name) {
				log.Printf("ERROR: duplicate `%s' stack name", stack.Name)
				return os.ErrExist
			}

			stackSet.Add(stack.Name)

			for _, group := range stack.Groups {
				if group == nil {
					log.Printf("ERROR: found null group in `%s' stack", stack.Name)
					return os.ErrInvalid
				} else if groupSet.Has(group.Name) {
					log.Printf("ERROR: duplicate `%s' group name", group.Name)
					return os.ErrExist
				}

				groupSet.Add(group.Name)

				for _, serie := range group.Series {
					if serie == nil {
						log.Printf("ERROR: found null serie in `%s' group", group.Name)
						return os.ErrInvalid
					} else if serieSet.Has(serie.Name) {
						log.Printf("ERROR: duplicate `%s' serie name", serie.Name)
						return os.ErrExist
					}

					serieSet.Add(serie.Name)
				}
			}
		}

		library.Graphs[itemStruct.ID] = item.(*Graph)
		library.Graphs[itemStruct.ID].ID = itemStruct.ID

	case LibraryItemCollection:
		library.Collections[itemStruct.ID] = item.(*Collection)
		library.Collections[itemStruct.ID].ID = itemStruct.ID
	}

	// Store JSON data
	if itemType != LibraryItemGraph || itemType == LibraryItemGraph && !item.(*Graph).Volatile {
		if err := utils.JSONDump(library.getFilePath(itemStruct.ID, itemType), item, itemStruct.Modified); err != nil {
			return err
		}
	}

	return nil
}

func (library *Library) getDirPath(itemType int) string {
	var dirName string

	switch itemType {
	case LibraryItemSourceGroup:
		dirName = "sourcegroups"

	case LibraryItemMetricGroup:
		dirName = "metricgroups"

	case LibraryItemGraph:
		dirName = "graphs"

	case LibraryItemCollection:
		dirName = "collections"
	}

	return path.Join(library.Config.DataDir, dirName)
}

func (library *Library) getFilePath(id string, itemType int) string {
	return path.Join(library.getDirPath(itemType), id[0:2], id[2:4], id+".json")
}
