// +build !disable_driver_pgsql

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
	pgsqlStorage      *Storage
	pgsqlProviders    []*Provider
	pgsqlSourceGroups []*SourceGroup
	pgsqlMetricGroups []*MetricGroup
	pgsqlGraphs       []*Graph
	pgsqlCollections  []*Collection
)

func init() {
	var (
		port int64
		err  error
	)

	config := maputil.Map{
		"driver": "pgsql",
	}

	if v := os.Getenv("TEST_PGSQL_DBNAME"); v != "" {
		config.Set("dbname", v)
	}
	if v := os.Getenv("TEST_PGSQL_HOST"); v != "" {
		config.Set("host", v)
	}
	if v := os.Getenv("TEST_PGSQL_PORT"); v != "" {
		port, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("failed to convert port to integer: %s", err))
		}
		config.Set("port", port)
	}
	if v := os.Getenv("TEST_PGSQL_USER"); v != "" {
		config.Set("user", v)
	}
	if v := os.Getenv("TEST_PGSQL_PASSWORD"); v != "" {
		config.Set("password", v)
	}

	pgsqlStorage, err = New(&config, log)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize PostgreSQL storage"))
	}

	pgsqlProviders = testProviderNew()
	pgsqlSourceGroups = testSourceGroupNew()
	pgsqlMetricGroups = testMetricGroupNew()
	pgsqlGraphs = testGraphNew()
	pgsqlCollections = testCollectionNew()
}

func Test_PgSQL_Providers_Create(t *testing.T) {
	testProviderCreate(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Create_Invalid(t *testing.T) {
	testProviderCreateInvalid(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Get(t *testing.T) {
	testProviderGet(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Get_Unknown(t *testing.T) {
	testProviderGetUnknown(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Update(t *testing.T) {
	testProviderUpdate(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Delete(t *testing.T) {
	testProviderDelete(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_List(t *testing.T) {
	testProviderList(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Count(t *testing.T) {
	testProviderCount(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Delete_All(t *testing.T) {
	testProviderDeleteAll(pgsqlStorage, pgsqlProviders, t)
}

func Test_PgSQL_SourceGroups_Create(t *testing.T) {
	testSourceGroupCreate(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Create_Invalid(t *testing.T) {
	testSourceGroupCreateInvalid(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Get(t *testing.T) {
	testSourceGroupGet(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Get_Unknown(t *testing.T) {
	testSourceGroupGetUnknown(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Update(t *testing.T) {
	testSourceGroupUpdate(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Delete(t *testing.T) {
	testSourceGroupDelete(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_List(t *testing.T) {
	testSourceGroupList(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Count(t *testing.T) {
	testSourceGroupCount(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Delete_All(t *testing.T) {
	testSourceGroupDeleteAll(pgsqlStorage, pgsqlSourceGroups, t)
}

func Test_PgSQL_MetricGroups_Create(t *testing.T) {
	testMetricGroupCreate(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Create_Invalid(t *testing.T) {
	testMetricGroupCreateInvalid(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Get(t *testing.T) {
	testMetricGroupGet(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Get_Unknown(t *testing.T) {
	testMetricGroupGetUnknown(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Update(t *testing.T) {
	testMetricGroupUpdate(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Delete(t *testing.T) {
	testMetricGroupDelete(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_List(t *testing.T) {
	testMetricGroupList(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Count(t *testing.T) {
	testMetricGroupCount(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Delete_All(t *testing.T) {
	testMetricGroupDeleteAll(pgsqlStorage, pgsqlMetricGroups, t)
}

func Test_PgSQL_Graphs_Create(t *testing.T) {
	testGraphCreate(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Create_Invalid(t *testing.T) {
	testGraphCreateInvalid(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Get(t *testing.T) {
	testGraphGet(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Get_Unknown(t *testing.T) {
	testGraphGetUnknown(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Update(t *testing.T) {
	testGraphUpdate(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Delete(t *testing.T) {
	testGraphDelete(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_List(t *testing.T) {
	testGraphList(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Count(t *testing.T) {
	testGraphCount(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Resolve(t *testing.T) {
	testGraphResolve(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Expand(t *testing.T) {
	testGraphExpand(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Delete_All(t *testing.T) {
	testGraphDeleteAll(pgsqlStorage, pgsqlGraphs, t)
}

func Test_PgSQL_Collections_Create(t *testing.T) {
	testCollectionCreate(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Create_Invalid(t *testing.T) {
	testCollectionCreateInvalid(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Get(t *testing.T) {
	testCollectionGet(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Get_Unknown(t *testing.T) {
	testCollectionGetUnknown(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Update(t *testing.T) {
	testCollectionUpdate(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Delete(t *testing.T) {
	testCollectionDelete(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_List(t *testing.T) {
	testCollectionList(pgsqlStorage, pgsqlCollections, pgsqlGraphs, t)
}

func Test_PgSQL_Collections_Count(t *testing.T) {
	testCollectionCount(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Resolve(t *testing.T) {
	testCollectionResolve(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Expand(t *testing.T) {
	testCollectionExpand(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Tree(t *testing.T) {
	testCollectionTree(pgsqlStorage, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Delete_All(t *testing.T) {
	testCollectionDeleteAll(pgsqlStorage, pgsqlCollections, t)
}
