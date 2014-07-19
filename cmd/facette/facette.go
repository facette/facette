package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/logger"
	"github.com/facette/facette/pkg/server"
	"github.com/facette/facette/pkg/utils"
)

const (
	cmdUsage string = "Usage: %s [OPTIONS]"
)

var (
	version      string
	flagConfig   string
	flagHelp     bool
	flagLog      string
	flagLogLevel string
	flagVersion  bool
	logLevel     int
	err          error
)

func init() {
	flag.StringVar(&flagConfig, "c", config.DefaultConfigFile, "configuration file path")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.StringVar(&flagLog, "l", config.DefaultLogFile, "log file path")
	flag.StringVar(&flagLogLevel, "L", config.DefaultLogLevel, "logging level (error, warning, notice, info, debug)")
	flag.BoolVar(&flagVersion, "V", false, "display software version and exit")
	flag.Usage = func() { utils.PrintUsage(os.Stderr, cmdUsage) }
	flag.Parse()

	if flagHelp {
		utils.PrintUsage(os.Stdout, cmdUsage)
	} else if flagVersion {
		utils.PrintVersion(version)
		os.Exit(0)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		utils.PrintUsage(os.Stderr, cmdUsage)
	}

	if logLevel, err = logger.GetLevelByName(flagLogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid log level `%s'\n", flagLogLevel)
		os.Exit(1)
	}
}

func main() {
	// Create new server instance and load configuration
	instance := server.NewServer(flagConfig, flagLog, logLevel)

	// Reload server configuration on SIGHUP
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGHUP:
				instance.Reload(true)
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
