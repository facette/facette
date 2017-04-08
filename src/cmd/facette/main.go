package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"facette/backend"
	"facette/connector"

	"github.com/cosiner/flag"
)

type command struct {
	Config  string `names:"-c, --config" usage:"configuration file path" default:"/etc/facette/facette.yaml"`
	Help    bool   `names:"-h, --help" usage:"display this help and exit"`
	Version bool   `names:"-V, --version" usage:"display version information and exit"`
}

func (*command) Metadata() map[string]flag.Flag {
	return map[string]flag.Flag{
		"": {
			Usage: "Time series data visualization software",
		},
	}
}

var (
	version   string
	buildDate string
	buildHash string

	cmd command
)

func main() {
	var (
		config  *config
		service *Service
		sigChan chan os.Signal
		err     error
	)

	flagSet := flag.NewFlagSet(flag.Flag{}).ErrHandling(0)
	flagSet.StructFlags(&cmd)

	if err := flagSet.Parse(os.Args...); err != nil {
		fmt.Printf("Error: %s\n", err)
		flagSet.Help(false)
		os.Exit(1)
	}

	if cmd.Version {
		printVersion()
		os.Exit(0)
	}

	// Load service configuration
	config, err = initConfig(cmd.Config)
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
