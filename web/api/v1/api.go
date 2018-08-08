package v1

import (
	"net/http"

	"facette.io/facette/backend"
	"facette.io/facette/catalog"
	"facette.io/facette/config"
	"facette.io/facette/poller"
	"facette.io/httputil"
	"facette.io/logger"
	"github.com/vbatoufflet/httproute"
)

// Prefix represents the versioned API prefix.
const Prefix = "/api/v1"

// API represents an API instance.
type API struct {
	router   *httproute.Router
	backend  *backend.Backend
	searcher *catalog.Searcher
	poller   *poller.Poller
	config   *config.Config
	logger   *logger.Logger
}

// NewAPI creates a new API instance.
func NewAPI(
	router *httproute.Router,
	backend *backend.Backend,
	searcher *catalog.Searcher,
	poller *poller.Poller,
	config *config.Config,
	logger *logger.Logger,
) *API {
	api := &API{
		router:   router,
		backend:  backend,
		searcher: searcher,
		poller:   poller,
		config:   config,
		logger:   logger,
	}

	root := router.Endpoint(Prefix).
		Use(handleCache).
		Get(api.infoGet)

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
		Delete(api.backendDeleteAll).
		Get(api.backendList).
		Post(api.backendCreate)
	root.Endpoint("/library/:type/:id").
		Delete(api.backendDelete).
		Get(api.backendGet).
		Patch(api.backendUpdate).
		Put(api.backendUpdate)

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
