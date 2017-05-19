// +build !disable_driver_mysql

package backend

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/facette/maputil"
	"github.com/pkg/errors"
)

var (
	mysqlBackend      *Backend
	mysqlProviders    []*Provider
	mysqlSourceGroups []*SourceGroup
	mysqlMetricGroups []*MetricGroup
	mysqlGraphs       []*Graph
	mysqlCollections  []*Collection
)

func init() {
	var err error

	config := maputil.Map{"driver": "mysql"}

	if v := os.Getenv("TEST_MYSQL_DBNAME"); v != "" {
		config.Set("dbname", v)
	}
	if v := os.Getenv("TEST_MYSQL_HOST"); v != "" {
		config.Set("host", v)
	}
	if v := os.Getenv("TEST_MYSQL_PORT"); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("failed to convert port to integer: %s", err))
		}
		config.Set("port", i)
	}
	if v := os.Getenv("TEST_MYSQL_USER"); v != "" {
		config.Set("user", v)
	}
	if v := os.Getenv("TEST_MYSQL_PASSWORD"); v != "" {
		config.Set("password", v)
	}

	mysqlBackend, err = NewBackend(&config, log)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize MySQL backend"))
	}

	mysqlProviders = testProviderNew()
	mysqlSourceGroups = testSourceGroupNew()
	mysqlMetricGroups = testMetricGroupNew()
	mysqlGraphs = testGraphNew()
	mysqlCollections = testCollectionNew()
}

func Test_MySQL_Providers_Create(t *testing.T) {
	testProviderCreate(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_Create_Invalid(t *testing.T) {
	testProviderCreateInvalid(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_Get(t *testing.T) {
	testProviderGet(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_Get_Unknown(t *testing.T) {
	testProviderGetUnknown(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_Update(t *testing.T) {
	testProviderUpdate(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_Delete(t *testing.T) {
	testProviderDelete(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_List(t *testing.T) {
	testProviderList(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_Count(t *testing.T) {
	testProviderCount(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_Providers_Delete_All(t *testing.T) {
	testProviderDeleteAll(mysqlBackend, mysqlProviders, t)
}

func Test_MySQL_SourceGroups_Create(t *testing.T) {
	testSourceGroupCreate(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Create_Invalid(t *testing.T) {
	testSourceGroupCreateInvalid(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Get(t *testing.T) {
	testSourceGroupGet(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Get_Unknown(t *testing.T) {
	testSourceGroupGetUnknown(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Update(t *testing.T) {
	testSourceGroupUpdate(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Delete(t *testing.T) {
	testSourceGroupDelete(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_List(t *testing.T) {
	testSourceGroupList(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Count(t *testing.T) {
	testSourceGroupCount(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Delete_All(t *testing.T) {
	testSourceGroupDeleteAll(mysqlBackend, mysqlSourceGroups, t)
}

func Test_MySQL_MetricGroups_Create(t *testing.T) {
	testMetricGroupCreate(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Create_Invalid(t *testing.T) {
	testMetricGroupCreateInvalid(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Get(t *testing.T) {
	testMetricGroupGet(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Get_Unknown(t *testing.T) {
	testMetricGroupGetUnknown(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Update(t *testing.T) {
	testMetricGroupUpdate(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Delete(t *testing.T) {
	testMetricGroupDelete(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_List(t *testing.T) {
	testMetricGroupList(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Count(t *testing.T) {
	testMetricGroupCount(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Delete_All(t *testing.T) {
	testMetricGroupDeleteAll(mysqlBackend, mysqlMetricGroups, t)
}

func Test_MySQL_Graphs_Create(t *testing.T) {
	testGraphCreate(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Create_Invalid(t *testing.T) {
	testGraphCreateInvalid(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Get(t *testing.T) {
	testGraphGet(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Get_Unknown(t *testing.T) {
	testGraphGetUnknown(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Update(t *testing.T) {
	testGraphUpdate(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Delete(t *testing.T) {
	testGraphDelete(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_List(t *testing.T) {
	testGraphList(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Count(t *testing.T) {
	testGraphCount(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Resolve(t *testing.T) {
	testGraphResolve(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Expand(t *testing.T) {
	testGraphExpand(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Delete_All(t *testing.T) {
	testGraphDeleteAll(mysqlBackend, mysqlGraphs, t)
}

func Test_MySQL_Collections_Create(t *testing.T) {
	testCollectionCreate(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Create_Invalid(t *testing.T) {
	testCollectionCreateInvalid(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Get(t *testing.T) {
	testCollectionGet(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Get_Unknown(t *testing.T) {
	testCollectionGetUnknown(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Update(t *testing.T) {
	testCollectionUpdate(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Delete(t *testing.T) {
	testCollectionDelete(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_List(t *testing.T) {
	testCollectionList(mysqlBackend, mysqlCollections, mysqlGraphs, t)
}

func Test_MySQL_Collections_Count(t *testing.T) {
	testCollectionCount(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Resolve(t *testing.T) {
	testCollectionResolve(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Expand(t *testing.T) {
	testCollectionExpand(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Tree(t *testing.T) {
	testCollectionTree(mysqlBackend, mysqlCollections, t)
}

func Test_MySQL_Collections_Delete_All(t *testing.T) {
	testCollectionDeleteAll(mysqlBackend, mysqlCollections, t)
}
