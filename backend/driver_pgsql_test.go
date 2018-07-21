// +build !disable_driver_pgsql

package backend

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"facette.io/maputil"
	"github.com/pkg/errors"
)

var (
	pgsqlBackend      *Backend
	pgsqlProviders    []*Provider
	pgsqlSourceGroups []*SourceGroup
	pgsqlMetricGroups []*MetricGroup
	pgsqlGraphs       []*Graph
	pgsqlCollections  []*Collection
)

func init() {
	var err error

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
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("failed to convert port to integer: %s", err))
		}
		config.Set("port", i)
	}
	if v := os.Getenv("TEST_PGSQL_USER"); v != "" {
		config.Set("user", v)
	}
	if v := os.Getenv("TEST_PGSQL_PASSWORD"); v != "" {
		config.Set("password", v)
	}

	pgsqlBackend, err = NewBackend(&config, log)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize PostgreSQL backend"))
	}

	pgsqlProviders = testProviderNew()
	pgsqlSourceGroups = testSourceGroupNew()
	pgsqlMetricGroups = testMetricGroupNew()
	pgsqlGraphs = testGraphNew()
	pgsqlCollections = testCollectionNew()
}

func Test_PgSQL_Providers_Create(t *testing.T) {
	testProviderCreate(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Create_Invalid(t *testing.T) {
	testProviderCreateInvalid(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Get(t *testing.T) {
	testProviderGet(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Get_Unknown(t *testing.T) {
	testProviderGetUnknown(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Update(t *testing.T) {
	testProviderUpdate(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Delete(t *testing.T) {
	testProviderDelete(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_List(t *testing.T) {
	testProviderList(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Count(t *testing.T) {
	testProviderCount(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_Providers_Delete_All(t *testing.T) {
	testProviderDeleteAll(pgsqlBackend, pgsqlProviders, t)
}

func Test_PgSQL_SourceGroups_Create(t *testing.T) {
	testSourceGroupCreate(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Create_Invalid(t *testing.T) {
	testSourceGroupCreateInvalid(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Get(t *testing.T) {
	testSourceGroupGet(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Get_Unknown(t *testing.T) {
	testSourceGroupGetUnknown(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Update(t *testing.T) {
	testSourceGroupUpdate(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Delete(t *testing.T) {
	testSourceGroupDelete(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_List(t *testing.T) {
	testSourceGroupList(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Count(t *testing.T) {
	testSourceGroupCount(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_SourceGroups_Delete_All(t *testing.T) {
	testSourceGroupDeleteAll(pgsqlBackend, pgsqlSourceGroups, t)
}

func Test_PgSQL_MetricGroups_Create(t *testing.T) {
	testMetricGroupCreate(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Create_Invalid(t *testing.T) {
	testMetricGroupCreateInvalid(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Get(t *testing.T) {
	testMetricGroupGet(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Get_Unknown(t *testing.T) {
	testMetricGroupGetUnknown(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Update(t *testing.T) {
	testMetricGroupUpdate(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Delete(t *testing.T) {
	testMetricGroupDelete(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_List(t *testing.T) {
	testMetricGroupList(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Count(t *testing.T) {
	testMetricGroupCount(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_MetricGroups_Delete_All(t *testing.T) {
	testMetricGroupDeleteAll(pgsqlBackend, pgsqlMetricGroups, t)
}

func Test_PgSQL_Graphs_Create(t *testing.T) {
	testGraphCreate(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Create_Invalid(t *testing.T) {
	testGraphCreateInvalid(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Get(t *testing.T) {
	testGraphGet(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Get_Unknown(t *testing.T) {
	testGraphGetUnknown(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Update(t *testing.T) {
	testGraphUpdate(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Delete(t *testing.T) {
	testGraphDelete(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_List(t *testing.T) {
	testGraphList(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Count(t *testing.T) {
	testGraphCount(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Resolve(t *testing.T) {
	testGraphResolve(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Expand(t *testing.T) {
	testGraphExpand(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Graphs_Delete_All(t *testing.T) {
	testGraphDeleteAll(pgsqlBackend, pgsqlGraphs, t)
}

func Test_PgSQL_Collections_Create(t *testing.T) {
	testCollectionCreate(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Create_Invalid(t *testing.T) {
	testCollectionCreateInvalid(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Get(t *testing.T) {
	testCollectionGet(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Get_Unknown(t *testing.T) {
	testCollectionGetUnknown(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Update(t *testing.T) {
	testCollectionUpdate(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Delete(t *testing.T) {
	testCollectionDelete(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_List(t *testing.T) {
	testCollectionList(pgsqlBackend, pgsqlCollections, pgsqlGraphs, t)
}

func Test_PgSQL_Collections_Count(t *testing.T) {
	testCollectionCount(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Resolve(t *testing.T) {
	testCollectionResolve(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Expand(t *testing.T) {
	testCollectionExpand(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Tree(t *testing.T) {
	testCollectionTree(pgsqlBackend, pgsqlCollections, t)
}

func Test_PgSQL_Collections_Delete_All(t *testing.T) {
	testCollectionDeleteAll(pgsqlBackend, pgsqlCollections, t)
}
