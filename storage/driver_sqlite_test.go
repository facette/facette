// +build !disable_driver_sqlite

package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"facette.io/logger"
	"facette.io/maputil"
	"github.com/pkg/errors"
)

var (
	sqliteStorage      *Storage
	sqliteProviders    []*Provider
	sqliteSourceGroups []*SourceGroup
	sqliteMetricGroups []*MetricGroup
	sqliteGraphs       []*Graph
	sqliteCollections  []*Collection
	sqliteTempFile     string
)

func init() {
	var (
		tmpFile *os.File
		err     error
	)

	config := maputil.Map{"driver": "sqlite"}

	if v := os.Getenv("TEST_SQLITE_PATH"); v != "" {
		config.Set("path", v)
	} else {
		tmpFile, err = ioutil.TempFile("", "facette_")
		if err != nil {
			panic(fmt.Sprintf("failed to create temporary file: %s", err))
		}
		sqliteTempFile = tmpFile.Name()

		config.Set("path", sqliteTempFile)
	}

	logger, _ := logger.NewLogger()
	sqliteStorage, err = New(&config, logger)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize SQLite storage"))
	}

	sqliteProviders = testProviderNew()
	sqliteSourceGroups = testSourceGroupNew()
	sqliteMetricGroups = testMetricGroupNew()
	sqliteGraphs = testGraphNew()
	sqliteCollections = testCollectionNew()
}

func Test_SQLite_Providers_Create(t *testing.T) {
	testProviderCreate(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_Create_Invalid(t *testing.T) {
	testProviderCreateInvalid(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_Get(t *testing.T) {
	testProviderGet(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_Get_Unknown(t *testing.T) {
	testProviderGetUnknown(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_Update(t *testing.T) {
	testProviderUpdate(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_Delete(t *testing.T) {
	testProviderDelete(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_List(t *testing.T) {
	testProviderList(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_Count(t *testing.T) {
	testProviderCount(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_Providers_Delete_All(t *testing.T) {
	testProviderDeleteAll(sqliteStorage, sqliteProviders, t)
}

func Test_SQLite_SourceGroups_Create(t *testing.T) {
	testSourceGroupCreate(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Create_Invalid(t *testing.T) {
	testSourceGroupCreateInvalid(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Get(t *testing.T) {
	testSourceGroupGet(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Get_Unknown(t *testing.T) {
	testSourceGroupGetUnknown(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Update(t *testing.T) {
	testSourceGroupUpdate(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Delete(t *testing.T) {
	testSourceGroupDelete(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_List(t *testing.T) {
	testSourceGroupList(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Count(t *testing.T) {
	testSourceGroupCount(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Delete_All(t *testing.T) {
	testSourceGroupDeleteAll(sqliteStorage, sqliteSourceGroups, t)
}

func Test_SQLite_MetricGroups_Create(t *testing.T) {
	testMetricGroupCreate(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Create_Invalid(t *testing.T) {
	testMetricGroupCreateInvalid(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Get(t *testing.T) {
	testMetricGroupGet(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Get_Unknown(t *testing.T) {
	testMetricGroupGetUnknown(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Update(t *testing.T) {
	testMetricGroupUpdate(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Delete(t *testing.T) {
	testMetricGroupDelete(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_List(t *testing.T) {
	testMetricGroupList(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Count(t *testing.T) {
	testMetricGroupCount(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Delete_All(t *testing.T) {
	testMetricGroupDeleteAll(sqliteStorage, sqliteMetricGroups, t)
}

func Test_SQLite_Graphs_Create(t *testing.T) {
	testGraphCreate(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Create_Invalid(t *testing.T) {
	testGraphCreateInvalid(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Get(t *testing.T) {
	testGraphGet(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Get_Unknown(t *testing.T) {
	testGraphGetUnknown(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Update(t *testing.T) {
	testGraphUpdate(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Delete(t *testing.T) {
	testGraphDelete(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_List(t *testing.T) {
	testGraphList(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Count(t *testing.T) {
	testGraphCount(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Resolve(t *testing.T) {
	testGraphResolve(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Expand(t *testing.T) {
	testGraphExpand(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Delete_All(t *testing.T) {
	testGraphDeleteAll(sqliteStorage, sqliteGraphs, t)
}

func Test_SQLite_Collections_Create(t *testing.T) {
	testCollectionCreate(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Create_Invalid(t *testing.T) {
	testCollectionCreateInvalid(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Get(t *testing.T) {
	testCollectionGet(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Get_Unknown(t *testing.T) {
	testCollectionGetUnknown(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Update(t *testing.T) {
	testCollectionUpdate(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Delete(t *testing.T) {
	testCollectionDelete(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_List(t *testing.T) {
	testCollectionList(sqliteStorage, sqliteCollections, sqliteGraphs, t)
}

func Test_SQLite_Collections_Count(t *testing.T) {
	testCollectionCount(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Resolve(t *testing.T) {
	testCollectionResolve(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Expand(t *testing.T) {
	testCollectionExpand(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Tree(t *testing.T) {
	testCollectionTree(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Collections_Delete_All(t *testing.T) {
	testCollectionDeleteAll(sqliteStorage, sqliteCollections, t)
}

func Test_SQLite_Cleanup(t *testing.T) {
	os.Remove(sqliteTempFile)
}
