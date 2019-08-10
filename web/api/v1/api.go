package v1

import (
	"net/http"
	"path/filepath"

	"facette.io/facette/catalog"
	"facette.io/facette/config"
	"facette.io/facette/poller"
	"facette.io/facette/storage"
	"facette.io/httputil"
	"facette.io/logger"
	"github.com/vbatoufflet/httprouter"
)

// Prefix represents the versioned API prefix.
const Prefix = "/api/v1"

// API represents an API instance.
type API struct {
	router   *httprouter.Router
	storage  *storage.Storage
	searcher *catalog.Searcher
	poller   *poller.Poller
	config   *config.Config
	logger   *logger.Logger
	prefix   string
}

// NewAPI creates a new API instance.
func NewAPI(
	router *httprouter.Router,
	storage *storage.Storage,
	searcher *catalog.Searcher,
	poller *poller.Poller,
	config *config.Config,
	logger *logger.Logger,
) *API {
	api := &API{
		router:   router,
		storage:  storage,
		searcher: searcher,
		poller:   poller,
		config:   config,
		logger:   logger,
		prefix:   Prefix,
	}

	if config.HTTP.BasePath != "" {
		api.prefix = filepath.Join(config.HTTP.BasePath, api.prefix)
	}

	root := router.Endpoint(api.prefix).
		Use(handleCache).
		Options(api.optionsGet)

	root.Endpoint("/bulk").
		Post(api.bulkExec)

	root.Endpoint("/catalog").
		Get(api.catalogSummary)
	root.Endpoint("/catalog/:type").
		Get(api.catalogList)
	root.Endpoint("/catalog/:type/*").
		Get(api.catalogGet)

	root.Endpoint("/library").
		Get(api.librarySummary)
	root.Endpoint("/library/parse").
		Post(api.libraryParse)
	root.Endpoint("/library/search").
		Post(api.librarySearch)
	root.Endpoint("/library/collections/tree").
		Get(api.libraryCollectionTree)
	root.Endpoint("/library/:type").
		Delete(api.storageDeleteAll).
		Get(api.storageList).
		Post(api.storageCreate)
	root.Endpoint("/library/:type/:id").
		Delete(api.storageDelete).
		Get(api.storageGet).
		Patch(api.storageUpdate).
		Put(api.storageUpdate)

	root.Endpoint("/providers").
		Delete(api.providerDeleteAll).
		Get(api.providerList).
		Post(api.providerCreate)
	root.Endpoint("/providers/:id").
		Delete(api.providerDelete).
		Get(api.providerGet).
		Patch(api.providerUpdate).
		Put(api.providerUpdate)
	root.Endpoint("/providers/:id/refresh").
		Post(api.providerRefresh)

	root.Endpoint("/series/expand").
		Post(api.seriesExpand)
	root.Endpoint("/series/points").
		Post(api.seriesPoints)

	root.Endpoint("/version").
		Get(api.versionGet)

	root.Endpoint("/*").
		Any(handleNotFound)

	return api
}

func handleCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(rw, r)
	})
}

func handleNotFound(rw http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(rw, newMessage(errUnknownEndpoint), http.StatusNotFound)
}
