package library

import (
	"facette/utils"
	"fmt"
	"github.com/fatih/goset"
	"github.com/nu7hatch/gouuid"
	"os"
	"path"
	"syscall"
	"time"
)

// Item represents the base structure of a Library item.
type Item struct {
	path        string
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Modified    time.Time `json:"-"`
}

// GetItem returns the base structure of a Library item.
func (item *Item) GetItem() *Item {
	return item
}

func (library *Library) getDirPath(itemType int) string {
	var (
		dirName string
	)

	switch itemType {
	case LibraryItemSourceGroup:
		dirName = "sourcegroups"
		break

	case LibraryItemMetricGroup:
		dirName = "metricgroups"
		break

	case LibraryItemGraph:
		dirName = "graphs"
		break

	case LibraryItemCollection:
		dirName = "collections"
		break
	}

	return path.Join(library.Config.DataDir, dirName)
}

func (library *Library) getFilePath(id string, itemType int) string {
	return path.Join(library.getDirPath(itemType), id[0:2], id[2:4], id+".json")
}

// DeleteItem removes an existing item from the library.
func (library *Library) DeleteItem(id string, itemType int) error {
	var (
		err error
	)

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
		if err = syscall.Unlink(library.getFilePath(id, itemType)); err != nil {
			return err
		}
	}

	// Delete item from library
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		delete(library.Groups, id)
		break

	case LibraryItemGraph:
		delete(library.Graphs, id)
		break

	case LibraryItemCollection:
		delete(library.Collections, id)
		break
	}

	return nil
}

// GetItem gets an item from the library by its identifier.
func (library *Library) GetItem(id string, itemType int) (interface{}, error) {
	var (
		item interface{}
	)

	if !library.ItemExists(id, itemType) {
		return nil, os.ErrNotExist
	}

	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		return library.Groups[id], nil

	case LibraryItemGraph:
		item = library.Graphs[id]

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

		break

	case LibraryItemGraph:
		for _, item := range library.Graphs {
			if item.Name != name {
				continue
			}

			return item, nil
		}

		break

	case LibraryItemCollection:
		for _, item := range library.Collections {
			if item.Name != name {
				continue
			}

			return item, nil
		}

		break
	}

	return nil, os.ErrNotExist
}

// ItemExists returns whether an item existsn the library or not.
func (library *Library) ItemExists(id string, itemType int) bool {
	var (
		exists bool
	)

	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		if _, ok := library.Groups[id]; ok && library.Groups[id].Type == itemType {
			exists = true
		}
		break

	case LibraryItemGraph:
		_, exists = library.Graphs[id]
		break

	case LibraryItemGraphTemplate:
		_, exists = library.TemplateGraphs[id]
		break

	case LibraryItemCollection:
		_, exists = library.Collections[id]
		break
	}

	return exists
}

// LoadItem loads an item from the filesystem by its identifier.
func (library *Library) LoadItem(id string, itemType int) error {
	var (
		item          interface{}
		err           error
		fileInfo      os.FileInfo
		tmpCollection *struct {
			*Collection
			Parent string `json:"parent"`
		}
		tmpGraph *Graph
		tmpGroup *Group
	)

	// Load item from file
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		tmpGroup = &Group{}
		if fileInfo, err = utils.JSONLoad(library.getFilePath(id, itemType), &tmpGroup); err != nil {
			return err
		}

		library.Groups[id] = tmpGroup
		library.Groups[id].Type = itemType
		library.Groups[id].Modified = fileInfo.ModTime()
		break

	case LibraryItemGraph:
		tmpGraph = &Graph{}
		if fileInfo, err = utils.JSONLoad(library.getFilePath(id, itemType), &tmpGraph); err != nil {
			return err
		}

		library.Graphs[id] = tmpGraph
		library.Graphs[id].Volatile = false
		library.Graphs[id].Modified = fileInfo.ModTime()
		break

	case LibraryItemCollection:
		if fileInfo, err = utils.JSONLoad(library.getFilePath(id, LibraryItemCollection), &tmpCollection); err != nil {
			return err
		}

		if !library.ItemExists(id, LibraryItemCollection) {
			library.Collections[id] = &Collection{}
		}

		*library.Collections[id] = *tmpCollection.Collection

		if tmpCollection.Parent != "" {
			if item, err = library.GetItem(tmpCollection.Parent, LibraryItemCollection); err == nil {
				library.Collections[id].Parent = item.(*Collection)
				library.Collections[id].Parent.Children = append(library.Collections[id].Parent.Children,
					library.Collections[id])
			}
		}

		library.Collections[id].Modified = fileInfo.ModTime()

		break
	}

	return nil
}

// StoreItem stores an item into the library.
func (library *Library) StoreItem(item interface{}, itemType int) error {
	var (
		err        error
		groupSet   *goset.Set
		itemStruct *Item
		itemTemp   interface{}
		serieSet   *goset.Set
		stackSet   *goset.Set
		uuidTemp   *uuid.UUID
	)

	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		itemStruct = item.(*Group).GetItem()
		break

	case LibraryItemGraph:
		itemStruct = item.(*Graph).GetItem()
		break

	case LibraryItemCollection:
		itemStruct = item.(*Collection).GetItem()
		break
	}

	if itemStruct.ID == "" {
		if uuidTemp, err = uuid.NewV4(); err != nil {
			return err
		}

		itemStruct.ID = uuidTemp.String()
	} else if !library.ItemExists(itemStruct.ID, itemType) {
		return os.ErrNotExist
	}

	// Check for name field presence/duplicates
	if itemStruct.Name == "" {
		return os.ErrInvalid
	} else if itemTemp, err = library.GetItemByName(itemStruct.Name, itemType); err == nil {
		switch itemType {
		case LibraryItemSourceGroup, LibraryItemMetricGroup:
			if itemTemp.(*Group).ID != itemStruct.ID {
				return os.ErrExist
			}

			break

		case LibraryItemGraph:
			if !item.(*Graph).Volatile && itemTemp.(*Graph).ID != itemStruct.ID {
				return os.ErrExist
			}

			break

		case LibraryItemCollection:
			if itemTemp.(*Collection).ID != itemStruct.ID {
				return os.ErrExist
			}

			break
		}
	}

	// Store item into library
	switch itemType {
	case LibraryItemSourceGroup, LibraryItemMetricGroup:
		library.Groups[itemStruct.ID] = item.(*Group)
		library.Groups[itemStruct.ID].ID = itemStruct.ID
		break

	case LibraryItemGraph:
		// Check for definition names duplicates
		stackSet = goset.New()
		groupSet = goset.New()
		serieSet = goset.New()

		for _, stack := range item.(*Graph).Stacks {
			if stackSet.Has(stack.Name) {
				return os.ErrExist
			}

			stackSet.Add(stack.Name)

			for _, group := range stack.Groups {
				if groupSet.Has(group.Name) {
					return os.ErrExist
				}

				groupSet.Add(group.Name)

				for _, serie := range group.Series {
					if serieSet.Has(serie.Name) {
						return os.ErrExist
					}

					serieSet.Add(serie.Name)
				}
			}
		}

		library.Graphs[itemStruct.ID] = item.(*Graph)
		library.Graphs[itemStruct.ID].ID = itemStruct.ID
		break

	case LibraryItemCollection:
		library.Collections[itemStruct.ID] = item.(*Collection)
		library.Collections[itemStruct.ID].ID = itemStruct.ID
		break
	}

	// Store JSON data
	if itemType != LibraryItemGraph || itemType == LibraryItemGraph && !item.(*Graph).Volatile {
		if err = utils.JSONDump(library.getFilePath(itemStruct.ID, itemType), item, itemStruct.Modified); err != nil {
			return err
		}
	}

	return nil
}
