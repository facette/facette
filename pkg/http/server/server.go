// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

// Package server provides the HTTP server.
package server

import (
	"context"

	"batou.dev/httprouter"
	"batou.dev/httpserver"
	"go.uber.org/zap"

	"facette.io/facette/pkg/catalog"
	"facette.io/facette/pkg/poller"
	"facette.io/facette/pkg/store"

	"facette.io/facette/pkg/http/server/internal/api"
	"facette.io/facette/pkg/http/server/internal/assets"
)

// Server is an HTTP server.
type Server struct {
	config *Config
	server httpserver.Server
}

// New creates a new HTTP server instance.
func New(config *Config, catalog *catalog.Catalog, store *store.Store, poller *poller.Poller) *Server {
	router := httprouter.New()

	if zap.L().Core().Enabled(zap.DebugLevel) {
		router.Use(debugLog)
	}

	api.Register(router, catalog, store, poller)

	assets.Register(router)

	return &Server{
		config: config,
		server: httpserver.Server{
			Addr:            config.Address,
			Handler:         router,
			ShutdownTimeout: config.ShutdownTimeout,
		},
	}
}

// Run satisfies the fantask.Task interface.
func (s *Server) Run(ctx context.Context) error {
	log := zap.L().Named("http/server")

	log.Info("server started", zap.String("address", s.config.Address))
	defer log.Info("server stopped")

	return s.server.Run(ctx)
}
