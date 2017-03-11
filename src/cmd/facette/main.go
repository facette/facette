package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"facette/backend"
	"facette/connector"
)

var (
	version   string
	buildDate string
	buildHash string

	flagConfig  string
	flagHelp    bool
	flagVersion bool
)

func main() {
	var (
		config  *config
		service *Service
		sigChan chan os.Signal
		err     error
	)

	flag.StringVar(&flagConfig, "c", "/etc/facette/facette.conf", "configuration file path")
	flag.BoolVar(&flagHelp, "h", false, "display this help")
	flag.BoolVar(&flagVersion, "V", false, "display version and support information")
	flag.Usage = func() { printUsage(os.Stderr); os.Exit(1) }
	flag.Parse()

	if flagHelp {
		printUsage(os.Stdout)
		os.Exit(0)
	} else if flagVersion {
		printVersion()
		os.Exit(0)
	}

	// Load service configuration
	config, err = initConfig(flagConfig)
	if err != nil {
		goto end
	}

	// Start service instance
	service = NewService(config)

	// Handle service signals
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				service.Shutdown()

			case syscall.SIGUSR1:
				service.Refresh()
			}
		}
	}()

	err = service.Run()

end:
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: %s [OPTIONS]", filepath.Base(os.Args[0]))
	fmt.Fprint(w, "\n\nOptions:\n")

	flag.VisitAll(func(f *flag.Flag) {
		if !strings.HasPrefix(f.Name, "httptest.") {
			fmt.Fprintf(w, "   -%s  %s\n", f.Name, f.Usage)
		}
	})
}

func printVersion() {
	fmt.Printf("Version:     %s\n", version)
	fmt.Printf("Build date:  %s\n", buildDate)
	fmt.Printf("Build hash:  %s\n", buildHash)
	fmt.Printf("Compiler:    %s (%s)\n", runtime.Version(), runtime.Compiler)

	drivers := backend.Drivers()
	if len(drivers) == 0 {
		drivers = append(drivers, "none")
	}
	fmt.Printf("Drivers:     %s\n", strings.Join(drivers, ", "))

	connectors := connector.Connectors()
	if len(connectors) == 0 {
		connectors = append(connectors, "none")
	}
	fmt.Printf("Connectors:  %s\n", strings.Join(connectors, ", "))
}
