// +build !disable_driver_sqlite

package backend

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"facette.io/maputil"
	"github.com/pkg/errors"
)

var (
	sqliteBackend      *Backend
	sqliteProviders    []*Provider
	sqliteSourceGroups []*SourceGroup
	sqliteMetricGroups []*MetricGroup
	sqliteGraphs       []*Graph
	sqliteCollections  []*Collection
	sqliteTempFile     string
)

func init() {
	var err error

	config := maputil.Map{"driver": "sqlite"}

	if v := os.Getenv("TEST_SQLITE_PATH"); v != "" {
		config.Set("path", v)
	} else {
		tmpFile, err := ioutil.TempFile("", "facette")
		if err != nil {
			panic(fmt.Sprintf("failed to create temporary file: %s", err))
		}
		sqliteTempFile = tmpFile.Name()

		config.Set("path", sqliteTempFile)
	}

	sqliteBackend, err = NewBackend(&config, log)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize SQLite backend"))
	}

	sqliteProviders = testProviderNew()
	sqliteSourceGroups = testSourceGroupNew()
	sqliteMetricGroups = testMetricGroupNew()
	sqliteGraphs = testGraphNew()
	sqliteCollections = testCollectionNew()
}

func Test_SQLite_Providers_Create(t *testing.T) {
	testProviderCreate(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_Create_Invalid(t *testing.T) {
	testProviderCreateInvalid(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_Get(t *testing.T) {
	testProviderGet(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_Get_Unknown(t *testing.T) {
	testProviderGetUnknown(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_Update(t *testing.T) {
	testProviderUpdate(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_Delete(t *testing.T) {
	testProviderDelete(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_List(t *testing.T) {
	testProviderList(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_Count(t *testing.T) {
	testProviderCount(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_Providers_Delete_All(t *testing.T) {
	testProviderDeleteAll(sqliteBackend, sqliteProviders, t)
}

func Test_SQLite_SourceGroups_Create(t *testing.T) {
	testSourceGroupCreate(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Create_Invalid(t *testing.T) {
	testSourceGroupCreateInvalid(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Get(t *testing.T) {
	testSourceGroupGet(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Get_Unknown(t *testing.T) {
	testSourceGroupGetUnknown(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Update(t *testing.T) {
	testSourceGroupUpdate(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Delete(t *testing.T) {
	testSourceGroupDelete(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_List(t *testing.T) {
	testSourceGroupList(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Count(t *testing.T) {
	testSourceGroupCount(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_SourceGroups_Delete_All(t *testing.T) {
	testSourceGroupDeleteAll(sqliteBackend, sqliteSourceGroups, t)
}

func Test_SQLite_MetricGroups_Create(t *testing.T) {
	testMetricGroupCreate(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Create_Invalid(t *testing.T) {
	testMetricGroupCreateInvalid(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Get(t *testing.T) {
	testMetricGroupGet(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Get_Unknown(t *testing.T) {
	testMetricGroupGetUnknown(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Update(t *testing.T) {
	testMetricGroupUpdate(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Delete(t *testing.T) {
	testMetricGroupDelete(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_List(t *testing.T) {
	testMetricGroupList(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Count(t *testing.T) {
	testMetricGroupCount(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_MetricGroups_Delete_All(t *testing.T) {
	testMetricGroupDeleteAll(sqliteBackend, sqliteMetricGroups, t)
}

func Test_SQLite_Graphs_Create(t *testing.T) {
	testGraphCreate(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Create_Invalid(t *testing.T) {
	testGraphCreateInvalid(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Get(t *testing.T) {
	testGraphGet(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Get_Unknown(t *testing.T) {
	testGraphGetUnknown(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Update(t *testing.T) {
	testGraphUpdate(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Delete(t *testing.T) {
	testGraphDelete(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_List(t *testing.T) {
	testGraphList(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Count(t *testing.T) {
	testGraphCount(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Resolve(t *testing.T) {
	testGraphResolve(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Expand(t *testing.T) {
	testGraphExpand(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Graphs_Delete_All(t *testing.T) {
	testGraphDeleteAll(sqliteBackend, sqliteGraphs, t)
}

func Test_SQLite_Collections_Create(t *testing.T) {
	testCollectionCreate(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Create_Invalid(t *testing.T) {
	testCollectionCreateInvalid(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Get(t *testing.T) {
	testCollectionGet(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Get_Unknown(t *testing.T) {
	testCollectionGetUnknown(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Update(t *testing.T) {
	testCollectionUpdate(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Delete(t *testing.T) {
	testCollectionDelete(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_List(t *testing.T) {
	testCollectionList(sqliteBackend, sqliteCollections, sqliteGraphs, t)
}

func Test_SQLite_Collections_Count(t *testing.T) {
	testCollectionCount(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Resolve(t *testing.T) {
	testCollectionResolve(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Expand(t *testing.T) {
	testCollectionExpand(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Tree(t *testing.T) {
	testCollectionTree(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Collections_Delete_All(t *testing.T) {
	testCollectionDeleteAll(sqliteBackend, sqliteCollections, t)
}

func Test_SQLite_Cleanup(t *testing.T) {
	os.Remove(sqliteTempFile)
}
