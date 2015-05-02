package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/facette/facette/pkg/cmd"
	"github.com/facette/facette/pkg/config"
)

const (
	cmdUsage = `Usage: %s [OPTIONS] COMMAND

Commands:
   refresh  refresh server catalog and library`

	defaultConfigFile string = "/etc/facette/facette.json"
)

var (
	version     string
	buildDate   string
	flagConfig  string
	flagHelp    bool
	flagVersion bool
)

func init() {
	flag.StringVar(&flagConfig, "c", defaultConfigFile, "configuration file path")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.BoolVar(&flagVersion, "V", false, "display software version and exit")
	flag.Usage = func() { cmd.PrintUsage(os.Stderr, cmdUsage) }
	flag.Parse()

	if flagHelp {
		cmd.PrintUsage(os.Stdout, cmdUsage)
	} else if flagVersion {
		cmd.PrintVersion(version, buildDate)
		os.Exit(0)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		cmd.PrintUsage(os.Stderr, cmdUsage)
	}
}

func main() {
	var handler func(*config.Config, []string) error

	cfg := &config.Config{}

	if err := cfg.Load(flagConfig); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	if len(flag.Args()) == 0 {
		cmd.PrintUsage(os.Stderr, cmdUsage)
	}

	switch flag.Args()[0] {
	case "refresh":
		handler = handleService
	default:
		cmd.PrintUsage(os.Stderr, cmdUsage)
		os.Exit(1)
	}

	err := handler(cfg, flag.Args())
	if err == os.ErrInvalid {
		cmd.PrintUsage(os.Stderr, cmdUsage)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
	}
}
