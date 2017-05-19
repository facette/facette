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
	"gopkg.in/tylerb/graceful.v1"
)

const apiPrefix = "/api/v1"

type httpWorker struct {
	sync.Mutex
	worker.CommonWorker

	service *Service
	log     *logger.Logger
	router  *httproute.Router
	server  *graceful.Server
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
		Get(w.httpHandleCatalogRoot)
	w.router.Endpoint(w.prefix + "/catalog/:type/").
		Get(w.httpHandleCatalogType)
	w.router.Endpoint(w.prefix + "/catalog/:type/:name").
		Get(w.httpHandleCatalogEntry)

	w.router.Endpoint(w.prefix + "/expand").
		Post(w.httpHandleExpand)

	w.router.Endpoint(w.prefix + "/library/").
		Get(w.httpHandleLibraryRoot)
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

	w.router.Endpoint(w.prefix + "/plots").
		Post(w.httpHandlePlots)

	providerCtx := context.WithValue(context.Background(), "type", "providers")

	w.router.EndpointWithContext(w.prefix+"/providers/", providerCtx).
		Delete(w.httpHandleBackendDeleteAll).
		Get(w.httpHandleBackendList).
		Post(w.httpHandleBackendCreate)
	w.router.EndpointWithContext(w.prefix+"/providers/:id", providerCtx).
		Delete(w.httpHandleBackendDelete).
		Get(w.httpHandleBackendGet).
		Patch(w.httpHandleBackendUpdate).
		Put(w.httpHandleBackendUpdate)

	w.router.Endpoint(w.prefix + "/providers/:id/refresh").
		Post(w.httpHandleProviderRefresh)

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

	w.Lock()
	w.server = &graceful.Server{
		Server: &http.Server{
			Addr:    netAddr,
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
