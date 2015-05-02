package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/facette/facette/pkg/cmd"
	"github.com/facette/facette/pkg/connector"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/server"
)

const (
	cmdUsage string = "Usage: %s [OPTIONS]"

	defaultConfigFile string = "/etc/facette/facette.json"
	defaultLogPath    string = ""
	defaultLogLevel   string = "warning"
)

var (
	version      string
	buildDate    string
	flagConfig   string
	flagHelp     bool
	flagLogPath  string
	flagLogLevel string
	flagVersion  bool
	logLevel     int
	err          error
)

func init() {
	flag.StringVar(&flagConfig, "c", defaultConfigFile, "configuration file path")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.StringVar(&flagLogPath, "l", defaultLogPath, "log file path")
	flag.StringVar(&flagLogLevel, "L", defaultLogLevel, "logging level (error, warning, notice, info, debug)")
	flag.BoolVar(&flagVersion, "V", false, "display software version and exit")
	flag.Usage = func() { cmd.PrintUsage(os.Stderr, cmdUsage) }
	flag.Parse()

	if flagHelp {
		cmd.PrintUsage(os.Stdout, cmdUsage)
	} else if flagVersion {
		cmd.PrintVersion(version, buildDate)

		connectors := []string{}
		for connector := range connector.Connectors {
			connectors = append(connectors, connector)
		}

		sort.Strings(connectors)

		fmt.Printf("\nAvailable connectors:\n")
		for _, connector := range connectors {
			fmt.Printf("   %s\n", connector)
		}

		os.Exit(0)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		cmd.PrintUsage(os.Stderr, cmdUsage)
	}

	if logLevel, err = logger.GetLevelByName(flagLogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid log level `%s'\n", flagLogLevel)
		os.Exit(1)
	}
}

func main() {
	// Create new server instance and load configuration
	instance := server.NewServer(flagConfig, flagLogPath, logLevel)

	// Handle server signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGUSR1:
				instance.Refresh()
				break

			case syscall.SIGINT, syscall.SIGTERM:
				instance.Stop()
				break
			}
		}
	}()

	// Run instance
	if err := instance.Run(); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
