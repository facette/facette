// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package poller

import (
	"context"
	"time"
	"unsafe"

	"go.uber.org/zap"

	"facette.io/facette/pkg/api"
	"facette.io/facette/pkg/catalog"
	"facette.io/facette/pkg/connector"
	"facette.io/facette/pkg/filter"
)

type job struct {
	provider  *api.Provider
	catalog   *catalog.Catalog
	connector connector.Connector
	ctx       context.Context
	cancel    context.CancelFunc
	polling   bool
	log       *zap.Logger
}

func newJob(ctx context.Context, provider *api.Provider, catalog *catalog.Catalog) (*job, error) {
	config := connector.Config(provider.Connector)

	conn, err := connector.New(config.Type, provider.Name, config.Settings)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &job{
		provider:  provider,
		catalog:   catalog,
		ctx:       ctx,
		cancel:    cancel,
		connector: conn,
		log:       zap.L().Named("poller/job").With(zap.String("provider", provider.ID)),
	}, nil
}

func (j *job) Run() {
	j.log.Debug("job started")
	defer j.log.Debug("job stopped")

	if j.provider.PollInterval > 0 {
		// Provider has a polling interval, thus initializing a ticker to
		// handle continuous polling and trigger initial poll.
		t := time.NewTicker(j.provider.PollInterval)

		go func() { *(*chan time.Time)(unsafe.Pointer(&t.C)) <- time.Now() }() // nolint:gosec

	loop:
		for {
			select {
			case <-t.C:
				j.Poll()

			case <-j.ctx.Done():
				if t != nil {
					t.Stop()
				}

				break loop
			}
		}
	} else {
		// Provider doesn't have a polling interval, thus only perform initial
		// polling and wait for context cancelation.
		j.Poll()
		<-j.ctx.Done()
	}
}

func (j *job) Poll() {
	if j.polling {
		j.log.Debug("polling already in progress, skipped")
		return
	}

	j.polling = true

	defer func() { j.polling = false }()

	j.log.Debug("polling started")
	defer j.log.Debug("polling stopped")

	ch := make(chan catalog.Metric)
	errCh := make(chan error)

	go func() {
		for err := range errCh {
			if err != context.Canceled {
				j.log.Error(err.Error())
			}
		}
	}()

	go j.connector.Metrics(j.ctx, ch, errCh)

	section := catalog.NewSection(j.connector)

	for metric := range filter.New(ch, j.provider.Filters) {
		err := section.Insert(metric)
		if err != nil {
			j.log.Warn("invalid metric discarded", zap.Error(err), zap.String("metric", metric.Labels.String()))
		} else {
			j.log.Debug("metric inserted", zap.String("metric", metric.Labels.String()))
		}
	}

	j.catalog.Link(j.provider.Name, section)
	j.log.Debug("catalog section linked")
}

func (j *job) Shutdown() {
	j.catalog.Unlink(j.provider.Name)
	j.log.Debug("catalog section unlinked")

	j.cancel()
}
