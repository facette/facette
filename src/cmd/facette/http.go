package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"facette/worker"

	"github.com/facette/httproute"
	"github.com/facette/logger"
)

const apiPrefix = "/api/v1"

type httpWorker struct {
	sync.Mutex
	worker.CommonWorker

	service *Service
	log     *logger.Logger
	router  *httproute.Router
	server  *http.Server
	prefix  string
}

func newHTTPWorker(s *Service) *httpWorker {
	w := &httpWorker{
		service: s,
		log:     s.log.Context("http"),
		router:  httproute.NewRouter(),
		prefix:  s.config.RootPath + apiPrefix,
	}

	// Initialize HTTP router
	w.router.Use(w.httpHandleLogger)

	w.router.Endpoint(w.prefix + "/bulk").
		Post(w.httpHandleBulk)

	w.router.Endpoint(w.prefix + "/catalog/").
		Get(w.httpHandleCatalogSummary)
	w.router.Endpoint(w.prefix + "/catalog/:type/").
		Get(w.httpHandleCatalogType)
	w.router.Endpoint(w.prefix + "/catalog/:type/*").
		Get(w.httpHandleCatalogEntry)

	w.router.Endpoint(w.prefix + "/library/").
		Get(w.httpHandleLibrarySummary)
	w.router.Endpoint(w.prefix + "/library/parse").
		Post(w.httpHandleLibraryParse)
	w.router.Endpoint(w.prefix + "/library/search").
		Post(w.httpHandleLibrarySearch)
	w.router.Endpoint(w.prefix + "/library/collections/tree").
		Get(w.httpHandleLibraryCollectionTree)
	w.router.Endpoint(w.prefix + "/library/:type/").
		Delete(w.httpHandleBackendDeleteAll).
		Get(w.httpHandleBackendList).
		Post(w.httpHandleBackendCreate)
	w.router.Endpoint(w.prefix + "/library/:type/:id").
		Delete(w.httpHandleBackendDelete).
		Get(w.httpHandleBackendGet).
		Patch(w.httpHandleBackendUpdate).
		Put(w.httpHandleBackendUpdate)

	w.router.Endpoint(w.prefix + "/providers/").
		Delete(w.httpHandleProviderDeleteAll).
		Get(w.httpHandleProviderList).
		Post(w.httpHandleProviderCreate)
	w.router.Endpoint(w.prefix + "/providers/:id").
		Delete(w.httpHandleProviderDelete).
		Get(w.httpHandleProviderGet).
		Patch(w.httpHandleProviderUpdate).
		Put(w.httpHandleProviderUpdate)
	w.router.Endpoint(w.prefix + "/providers/:id/refresh").
		Post(w.httpHandleProviderRefresh)

	w.router.Endpoint(w.prefix + "/series/expand").
		Post(w.httpHandleSeriesExpand)
	w.router.Endpoint(w.prefix + "/series/points").
		Post(w.httpHandleSeriesPoints)

	w.router.Endpoint(w.prefix + "/").
		Get(w.httpHandleInfo)

	w.router.Endpoint(w.prefix + "/*").
		Any(httpHandleNotFound)

	w.router.Endpoint("/*").
		Get(w.httpHandleAsset)

	return w
}

func (w *httpWorker) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	w.log.Debug("worker started")

	// Start router
	w.log.Info("listening on %q", w.service.config.Listen)

	netProto := "tcp"
	netAddr := w.service.config.Listen

	if strings.HasPrefix(netAddr, "unix:") {
		netProto = "unix"
		netAddr = strings.TrimPrefix(netAddr, "unix:")

	}

	listener, err := net.Listen(netProto, netAddr)
	if err != nil {
		w.log.Error("failed to listen: %s", err)
		return
	}
	defer listener.Close()

	if netProto == "unix" {
		socketUID := os.Getuid()
		if w.service.config.SocketUser != "" {
			user, err := user.Lookup(w.service.config.SocketUser)
			if err != nil {
				w.log.Error("failed to change socket ownership: %s", err)
				return
			}
			socketUID, _ = strconv.Atoi(user.Uid)
		}

		socketGID := os.Getgid()
		if w.service.config.SocketGroup != "" {
			group, err := user.LookupGroup(w.service.config.SocketGroup)
			if err != nil {
				w.log.Error("failed to change socket ownership: %s", err)
				return
			}
			socketGID, _ = strconv.Atoi(group.Gid)
		}

		if err := os.Chown(netAddr, socketUID, socketGID); err != nil {
			w.log.Error("failed to change socket ownership: %s", err)
			return
		}

		if w.service.config.SocketMode != "" {
			mode, err := strconv.ParseUint(w.service.config.SocketMode, 8, 32)
			if err != nil {
				w.log.Error("failed to change socket permissions: invalid socket mode")
				return
			}

			err = os.Chmod(netAddr, os.FileMode(mode))
			if err != nil {
				w.log.Error("failed to change socket permissions: %s", err)
				return
			}
		}
	}

	if builtinAssets {
		w.log.Info("serving web assets from built-in files")
	}

	w.Lock()
	w.server = &http.Server{
		Addr:    netAddr,
		Handler: w.router,
	}
	w.Unlock()

	if err := w.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		w.log.Error("failed to serve: %s", err)
	}

	w.log.Debug("worker stopped")
}

func (w *httpWorker) Shutdown() {
	// Gracefully stop HTTP server
	if w.server != nil {
		w.Lock()
		defer w.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(),
			time.Duration(w.service.config.GracefulTimeout)*time.Second)
		defer cancel()

		w.server.Shutdown(ctx)
	}

	// Shutdown worker
	w.CommonWorker.Shutdown()
}
