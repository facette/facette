package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"facette.io/facette/backend"
	"facette.io/facette/catalog"
	"facette.io/facette/config"
	"facette.io/facette/poller"
	"facette.io/facette/version"
	"facette.io/facette/web"
	"github.com/cosiner/flag"
	"github.com/oklog/run"
	"github.com/pkg/errors"
)

type command struct {
	Config  string `names:"-c, --config" usage:"configuration file path" default:"/etc/facette/facette.yaml"`
	Help    bool   `names:"-h, --help" usage:"display this help and exit"`
	Version bool   `names:"-V, --version" usage:"display version information and exit"`
}

func (*command) Metadata() map[string]flag.Flag {
	return map[string]flag.Flag{"": {Usage: "Time series data visualization software"}}
}

var cmd command

func init() {
	flagSet := flag.NewFlagSet(flag.Flag{}).ErrHandling(0)
	flagSet.StructFlags(&cmd)

	err := flagSet.Parse(os.Args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		flagSet.Help(false)
		os.Exit(2)
	} else if cmd.Version {
		version.Print()
		os.Exit(0)
	}
}

func main() {
	var g run.Group

	config, err := config.New(cmd.Config)
	if err != nil {
		die(errors.Wrap(err, "cannot initialize configuration"))
	}

	logger, err := newLogger(config)
	if err != nil {
		die(errors.Wrap(err, "cannot initialize logger"))
	}

	// Catch panic and write its output to the logger
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic: %s\n%s", r, debug.Stack())
			os.Exit(1)
		}
	}()

	// Initialize subcomponents pre-requisites
	backend, err := backend.New(config.Backend, logger.Context("backend"))
	if err != nil {
		die(errors.Wrap(err, "cannot initialize back-end"))
	}
	defer backend.Close()

	searcher := catalog.NewSearcher()

	// Run subcomponents and wait for them to finish their job
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller := poller.New(ctx, backend, searcher, config, logger.Context("poller"))
	g.Add(func() error { return poller.Run() }, func(error) { poller.Shutdown(); cancel() })

	web := web.NewHandler(ctx, backend, searcher, poller, config, logger.Context("http"))
	g.Add(func() error { return web.Run() }, func(error) { web.Shutdown(); cancel() })

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1)

	go func() {
		for s := range sc {
			switch s {
			case syscall.SIGUSR1:
				poller.RefreshAll()

			default:
				if ctx.Err() != context.Canceled {
					logger.Notice("received shutdown signal, stopping")
					cancel()
				}
			}
		}
	}()

	g.Run()
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}
