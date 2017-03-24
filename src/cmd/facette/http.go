package main

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"facette/worker"

	"github.com/facette/httproute"
	"github.com/facette/logger"
	"github.com/tylerb/graceful"
)

const apiPrefix = "/api/v1"

type httpWorker struct {
	sync.Mutex
	worker.CommonWorker

	service *Service
	log     *logger.Logger
	router  *httproute.Router
	server  *graceful.Server
}

func newHTTPWorker(s *Service) *httpWorker {
	return &httpWorker{
		service: s,
		log:     s.log.Context("http"),
		router:  httproute.NewRouter(),
	}
}

func (w *httpWorker) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	w.log.Debug("worker started")

	// Initialize HTTP router
	w.router.Use(w.httpHandleLogger)

	w.router.Endpoint(apiPrefix + "/bulk").
		Post(w.httpHandleBulk)

	w.router.Endpoint(apiPrefix + "/catalog/").
		Get(w.httpHandleCatalogRoot)
	w.router.Endpoint(apiPrefix + "/catalog/:type/").
		Get(w.httpHandleCatalogType)
	w.router.Endpoint(apiPrefix + "/catalog/:type/:name").
		Get(w.httpHandleCatalogEntry)

	w.router.Endpoint(apiPrefix + "/expand").
		Post(w.httpHandleExpand)

	w.router.Endpoint(apiPrefix + "/library/").
		Get(w.httpHandleLibraryRoot)
	w.router.Endpoint(apiPrefix + "/library/parse").
		Post(w.httpHandleLibraryParse)
	w.router.Endpoint(apiPrefix + "/library/search").
		Post(w.httpHandleLibrarySearch)
	w.router.Endpoint(apiPrefix + "/library/collections/tree").
		Get(w.httpHandleLibraryCollectionTree)
	w.router.Endpoint(apiPrefix + "/library/:type/").
		Delete(w.httpHandleBackendDeleteAll).
		Get(w.httpHandleBackendList).
		Post(w.httpHandleBackendCreate)
	w.router.Endpoint(apiPrefix + "/library/:type/:id").
		Delete(w.httpHandleBackendDelete).
		Get(w.httpHandleBackendGet).
		Patch(w.httpHandleBackendUpdate).
		Put(w.httpHandleBackendUpdate)

	w.router.Endpoint(apiPrefix + "/plots").
		Post(w.httpHandlePlots)

	w.router.Endpoint(apiPrefix+"/providers/").
		SetContext("type", "providers").
		Delete(w.httpHandleBackendDeleteAll).
		Get(w.httpHandleBackendList).
		Post(w.httpHandleBackendCreate)
	w.router.Endpoint(apiPrefix+"/providers/:id").
		SetContext("type", "providers").
		Delete(w.httpHandleBackendDelete).
		Get(w.httpHandleBackendGet).
		Patch(w.httpHandleBackendUpdate).
		Put(w.httpHandleBackendUpdate)

	w.router.Endpoint(apiPrefix + "/providers/:id/refresh").
		Post(w.httpHandleProviderRefresh)

	w.router.Endpoint(apiPrefix + "/").
		Get(w.httpHandleInfo)

	w.router.Endpoint(apiPrefix + "/*").
		Get(httpHandleNotFound)

	w.router.Endpoint("/*").
		Get(w.httpHandleAsset)

	// Start router
	w.log.Info("listening on %q", w.service.config.Listen)

	netProto := "tcp"
	if strings.HasPrefix(w.service.config.Listen, ".") || strings.HasPrefix(w.service.config.Listen, "/") {
		netProto = "unix"
	}

	listener, err := net.Listen(netProto, w.service.config.Listen)
	if err != nil {
		w.log.Error("failed to listen: %s", err)
		return
	}

	w.Lock()
	w.server = &graceful.Server{
		Server: &http.Server{
			Addr:    w.service.config.Listen,
			Handler: w.router,
		},
		NoSignalHandling: true,
		Timeout:          time.Duration(w.service.config.GracefulTimeout) * time.Second,
	}
	w.Unlock()

	if err := w.server.Serve(listener); err != nil {
		w.log.Error("failed to serve: %s", err)
		return
	}

	w.log.Debug("worker stopped")
}

func (w *httpWorker) Shutdown() {
	w.Lock()
	defer w.Unlock()

	// Trigger graceful shutdown
	if w.server != nil {
		w.server.Stop(w.server.Timeout)
	}

	// Shutdown worker
	w.CommonWorker.Shutdown()
}
