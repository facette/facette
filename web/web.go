package web

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"facette.io/facette/catalog"
	"facette.io/facette/config"
	"facette.io/facette/poller"
	"facette.io/facette/storage"
	"facette.io/facette/web/api/v1"
	"facette.io/logger"
	"github.com/vbatoufflet/httproute"
)

// Handler represents a HTTP handler serving the various endpoints.
type Handler struct {
	sync.Mutex

	ctx      context.Context
	storage  *storage.Storage
	searcher *catalog.Searcher
	poller   *poller.Poller
	config   *config.Config
	logger   *logger.Logger
	server   *http.Server
	shutdown bool
}

// NewHandler creates a new HTTP handler instance.
func NewHandler(
	ctx context.Context,
	storage *storage.Storage,
	searcher *catalog.Searcher,
	poller *poller.Poller,
	config *config.Config,
	logger *logger.Logger,
) *Handler {
	return &Handler{
		ctx:      ctx,
		storage:  storage,
		searcher: searcher,
		poller:   poller,
		config:   config,
		logger:   logger,
	}
}

// Run starts serving the HTTP endpoints.
func (h *Handler) Run() error {
	h.logger.Info("started")

	// Initialize HTTP router
	r := httproute.NewRouter()
	if h.config.LogLevel == "debug" {
		r.Use(h.handleLog)
	}

	v1.NewAPI(r, h.storage, h.searcher, h.poller, h.config, h.logger)

	r.Endpoint("/*").
		Get(h.handleAsset)

	proto := "tcp"
	addr := h.config.Listen

	if strings.HasPrefix(addr, "unix:") {
		proto = "unix"
		addr = strings.TrimPrefix(addr, "unix:")
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		h.logger.Error("failed to listen: %s", err)
		return err
	}
	defer listener.Close()

	if proto == "unix" {
		err = h.initSocket(addr)
		if err != nil {
			return err
		}
	}

	h.logger.Info("listening on %q", addr)

	h.Lock()
	h.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}
	h.Unlock()

	if !h.shutdown {
		err = h.server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			h.logger.Error("failed to serve: %s", err)
			return err
		}
	}

	h.logger.Info("stopped")

	return nil
}

// Shutdown gracefully stops the HTTP endpoints serving.
func (h *Handler) Shutdown() {
	h.Lock()
	defer h.Unlock()

	if h.shutdown {
		return
	} else if h.server == nil {
		h.shutdown = true
		return
	}

	ctx, cancel := context.WithTimeout(h.ctx, time.Duration(h.config.GracefulTimeout)*time.Second)
	defer cancel()

	h.server.Shutdown(ctx)
	h.server = nil
}
