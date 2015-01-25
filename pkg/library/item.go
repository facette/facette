package library

import (
	"fmt"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/fatih/set"
	uuid "github.com/facette/facette/thirdparty/github.com/nu7hatch/gouuid"
)

// Item represents the base structure of a library item.
type Item struct {
	path        string
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Modified    time.Time `json:"-"`
}

func (item *Item) String() string {
	return fmt.Sprintf(
		"Item{path:%q ID:%q Name:%q Description:%q Modified:%s}",
		item.path,
		item.ID,
		item.Name,
		item.Description,
		item.Modified,
	)
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
	if err := syscall.Unlink(library.getFilePath(id, itemType)); err != nil {
		return err
	}

	// Delete item from library
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		delete(library.Groups, id)

	case LibraryItemScale:
		delete(library.Scales, id)

	case LibraryItemUnit:
		delete(library.Units, id)

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

	case LibraryItemScale:
		return library.Scales[id], nil

	case LibraryItemUnit:
		return library.Units[id], nil

	case LibraryItemGraph:
		return library.Graphs[id], nil

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

	case LibraryItemScale:
		for _, item := range library.Scales {
			if item.Name != name {
				continue
			}

			return item, nil
		}

	case LibraryItemUnit:
		for _, item := range library.Units {
			if item.Name != name {
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

	case LibraryItemScale:
		_, exists = library.Scales[id]

	case LibraryItemUnit:
		_, exists = library.Units[id]

	case LibraryItemGraph:
		_, exists = library.Graphs[id]

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
			return fmt.Errorf("in %s, %s", filePath, err)
		}

		library.Groups[id] = tmpGroup
		library.Groups[id].Type = itemType
		library.Groups[id].Modified = fileInfo.ModTime()

	case LibraryItemScale:
		tmpScale := &Scale{}

		filePath := library.getFilePath(id, itemType)

		fileInfo, err := utils.JSONLoad(filePath, &tmpScale)
		if err != nil {
			return fmt.Errorf("in %s, %s", filePath, err)
		}

		library.Scales[id] = tmpScale
		library.Scales[id].Modified = fileInfo.ModTime()

	case LibraryItemUnit:
		tmpUnit := &Unit{}

		filePath := library.getFilePath(id, itemType)

		fileInfo, err := utils.JSONLoad(filePath, &tmpUnit)
		if err != nil {
			return fmt.Errorf("in %s, %s", filePath, err)
		}

		library.Units[id] = tmpUnit
		library.Units[id].Modified = fileInfo.ModTime()

	case LibraryItemGraph:
		tmpGraph := &Graph{}

		filePath := library.getFilePath(id, itemType)

		fileInfo, err := utils.JSONLoad(filePath, &tmpGraph)
		if err != nil {
			return fmt.Errorf("in %s, %s", filePath, err)
		}

		library.Graphs[id] = tmpGraph
		library.Graphs[id].Modified = fileInfo.ModTime()

	case LibraryItemCollection:
		var tmpCollection *struct {
			*Collection
			Parent string `json:"parent"`
		}

		filePath := library.getFilePath(id, LibraryItemCollection)

		fileInfo, err := utils.JSONLoad(filePath, &tmpCollection)
		if err != nil {
			return fmt.Errorf("in %s, %s", filePath, err)
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

	case LibraryItemScale:
		itemStruct = item.(*Scale).GetItem()

	case LibraryItemUnit:
		itemStruct = item.(*Unit).GetItem()

	case LibraryItemGraph:
		itemStruct = item.(*Graph).GetItem()

	case LibraryItemCollection:
		itemStruct = item.(*Collection).GetItem()

	default:
		return os.ErrInvalid
	}

	// If item has no ID specified, generate one
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
	if itemStruct.Name == "" {
		logger.Log(logger.LevelError, "library", "item missing `name' field")
		return os.ErrInvalid
	}

	itemTemp, err := library.GetItemByName(itemStruct.Name, itemType)

	// Item exists, check for duplicates
	if err == nil {
		switch itemType {
		case LibraryItemSourceGroup, LibraryItemMetricGroup:
			if itemTemp.(*Group).ID != itemStruct.ID {
				logger.Log(logger.LevelError, "library", "duplicate group identifier `%s'", itemStruct.ID)
				return os.ErrExist
			}

		case LibraryItemScale:
			if itemTemp.(*Scale).ID != itemStruct.ID {
				logger.Log(logger.LevelError, "library", "duplicate scale identifier `%s'", itemStruct.ID)
				return os.ErrExist
			}

		case LibraryItemUnit:
			if itemTemp.(*Unit).ID != itemStruct.ID {
				logger.Log(logger.LevelError, "library", "duplicate unit identifier `%s'", itemStruct.ID)
				return os.ErrExist
			}

		case LibraryItemGraph:
			if itemTemp.(*Graph).ID != itemStruct.ID {
				logger.Log(logger.LevelError, "library", "duplicate graph identifier `%s'", itemStruct.ID)
				return os.ErrExist
			}

		case LibraryItemCollection:
			if itemTemp.(*Collection).ID != itemStruct.ID {
				logger.Log(logger.LevelError, "library", "duplicate collection identifier `%s'", itemStruct.ID)
				return os.ErrExist
			}
		}
	}

	// Item does not exist, store it into library
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		library.Groups[itemStruct.ID] = item.(*Group)
		library.Groups[itemStruct.ID].ID = itemStruct.ID

	case LibraryItemScale:
		library.Scales[itemStruct.ID] = item.(*Scale)
		library.Scales[itemStruct.ID].ID = itemStruct.ID

	case LibraryItemUnit:
		library.Units[itemStruct.ID] = item.(*Unit)
		library.Units[itemStruct.ID].ID = itemStruct.ID

	case LibraryItemGraph:
		// Check for definition names duplicates
		groupSet := set.New(set.ThreadSafe)
		seriesSet := set.New(set.ThreadSafe)

		for _, group := range item.(*Graph).Groups {
			if group == nil {
				logger.Log(logger.LevelError, "library", "found null group")
				return os.ErrInvalid
			} else if groupSet.Has(group.Name) {
				logger.Log(logger.LevelError, "library", "duplicate group name `%s'", group.Name)
				return os.ErrExist
			}

			groupSet.Add(group.Name)

			for _, series := range group.Series {
				if series == nil {
					logger.Log(logger.LevelError, "library", "found null series in group `%s'", group.Name)
					return os.ErrInvalid
				} else if seriesSet.Has(series.Name) {
					logger.Log(logger.LevelError, "library", "duplicate series name `%s'", series.Name)
					return os.ErrExist
				}

				seriesSet.Add(series.Name)
			}
		}

		library.Graphs[itemStruct.ID] = item.(*Graph)
		library.Graphs[itemStruct.ID].ID = itemStruct.ID

	case LibraryItemCollection:
		library.Collections[itemStruct.ID] = item.(*Collection)
		library.Collections[itemStruct.ID].ID = itemStruct.ID
	}

	itemStruct.Modified = time.Now()

	// Store JSON data
	if err := utils.JSONDump(library.getFilePath(itemStruct.ID, itemType), item, itemStruct.Modified); err != nil {
		return err
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

	case LibraryItemScale:
		dirName = "scales"

	case LibraryItemUnit:
		dirName = "units"

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
