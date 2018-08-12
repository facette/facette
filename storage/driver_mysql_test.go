// +build !disable_driver_mysql

package storage

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"facette.io/maputil"
	"github.com/pkg/errors"
)

var (
	mysqlStorage      *Storage
	mysqlProviders    []*Provider
	mysqlSourceGroups []*SourceGroup
	mysqlMetricGroups []*MetricGroup
	mysqlGraphs       []*Graph
	mysqlCollections  []*Collection
)

func init() {
	var (
		port int64
		err  error
	)

	config := maputil.Map{"driver": "mysql"}

	if v := os.Getenv("TEST_MYSQL_DBNAME"); v != "" {
		config.Set("dbname", v)
	}
	if v := os.Getenv("TEST_MYSQL_HOST"); v != "" {
		config.Set("host", v)
	}
	if v := os.Getenv("TEST_MYSQL_PORT"); v != "" {
		port, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("failed to convert port to integer: %s", err))
		}
		config.Set("port", port)
	}
	if v := os.Getenv("TEST_MYSQL_USER"); v != "" {
		config.Set("user", v)
	}
	if v := os.Getenv("TEST_MYSQL_PASSWORD"); v != "" {
		config.Set("password", v)
	}

	mysqlStorage, err = New(&config, log)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize MySQL storage"))
	}

	mysqlProviders = testProviderNew()
	mysqlSourceGroups = testSourceGroupNew()
	mysqlMetricGroups = testMetricGroupNew()
	mysqlGraphs = testGraphNew()
	mysqlCollections = testCollectionNew()
}

func Test_MySQL_Providers_Create(t *testing.T) {
	testProviderCreate(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_Create_Invalid(t *testing.T) {
	testProviderCreateInvalid(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_Get(t *testing.T) {
	testProviderGet(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_Get_Unknown(t *testing.T) {
	testProviderGetUnknown(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_Update(t *testing.T) {
	testProviderUpdate(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_Delete(t *testing.T) {
	testProviderDelete(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_List(t *testing.T) {
	testProviderList(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_Count(t *testing.T) {
	testProviderCount(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_Providers_Delete_All(t *testing.T) {
	testProviderDeleteAll(mysqlStorage, mysqlProviders, t)
}

func Test_MySQL_SourceGroups_Create(t *testing.T) {
	testSourceGroupCreate(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Create_Invalid(t *testing.T) {
	testSourceGroupCreateInvalid(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Get(t *testing.T) {
	testSourceGroupGet(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Get_Unknown(t *testing.T) {
	testSourceGroupGetUnknown(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Update(t *testing.T) {
	testSourceGroupUpdate(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Delete(t *testing.T) {
	testSourceGroupDelete(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_List(t *testing.T) {
	testSourceGroupList(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Count(t *testing.T) {
	testSourceGroupCount(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_SourceGroups_Delete_All(t *testing.T) {
	testSourceGroupDeleteAll(mysqlStorage, mysqlSourceGroups, t)
}

func Test_MySQL_MetricGroups_Create(t *testing.T) {
	testMetricGroupCreate(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Create_Invalid(t *testing.T) {
	testMetricGroupCreateInvalid(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Get(t *testing.T) {
	testMetricGroupGet(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Get_Unknown(t *testing.T) {
	testMetricGroupGetUnknown(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Update(t *testing.T) {
	testMetricGroupUpdate(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Delete(t *testing.T) {
	testMetricGroupDelete(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_List(t *testing.T) {
	testMetricGroupList(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Count(t *testing.T) {
	testMetricGroupCount(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_MetricGroups_Delete_All(t *testing.T) {
	testMetricGroupDeleteAll(mysqlStorage, mysqlMetricGroups, t)
}

func Test_MySQL_Graphs_Create(t *testing.T) {
	testGraphCreate(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Create_Invalid(t *testing.T) {
	testGraphCreateInvalid(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Get(t *testing.T) {
	testGraphGet(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Get_Unknown(t *testing.T) {
	testGraphGetUnknown(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Update(t *testing.T) {
	testGraphUpdate(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Delete(t *testing.T) {
	testGraphDelete(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_List(t *testing.T) {
	testGraphList(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Count(t *testing.T) {
	testGraphCount(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Resolve(t *testing.T) {
	testGraphResolve(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Expand(t *testing.T) {
	testGraphExpand(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Graphs_Delete_All(t *testing.T) {
	testGraphDeleteAll(mysqlStorage, mysqlGraphs, t)
}

func Test_MySQL_Collections_Create(t *testing.T) {
	testCollectionCreate(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Create_Invalid(t *testing.T) {
	testCollectionCreateInvalid(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Get(t *testing.T) {
	testCollectionGet(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Get_Unknown(t *testing.T) {
	testCollectionGetUnknown(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Update(t *testing.T) {
	testCollectionUpdate(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Delete(t *testing.T) {
	testCollectionDelete(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_List(t *testing.T) {
	testCollectionList(mysqlStorage, mysqlCollections, mysqlGraphs, t)
}

func Test_MySQL_Collections_Count(t *testing.T) {
	testCollectionCount(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Resolve(t *testing.T) {
	testCollectionResolve(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Expand(t *testing.T) {
	testCollectionExpand(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Tree(t *testing.T) {
	testCollectionTree(mysqlStorage, mysqlCollections, t)
}

func Test_MySQL_Collections_Delete_All(t *testing.T) {
	testCollectionDeleteAll(mysqlStorage, mysqlCollections, t)
}
