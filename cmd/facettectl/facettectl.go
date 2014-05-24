package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/utils"
)

const (
	cmdUsage = `Usage: %s [OPTIONS] reload

Commands:
   reload    send reload signal to server`
)

var (
	flagConfig string
	flagDebug  int
	flagHelp   bool
)

func init() {
	flag.StringVar(&flagConfig, "c", config.DefaultConfigFile, "configuration file path")
	flag.IntVar(&flagDebug, "d", 0, "debugging level")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.Usage = func() { utils.PrintUsage(os.Stderr, cmdUsage) }
	flag.Parse()

	if flagHelp {
		utils.PrintUsage(os.Stdout, cmdUsage)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		utils.PrintUsage(os.Stderr, cmdUsage)
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
		utils.PrintUsage(os.Stderr, cmdUsage)
	}

	switch flag.Args()[0] {
	case "reload":
		handler = handleServer
	default:
		utils.PrintUsage(os.Stderr, cmdUsage)
		os.Exit(1)
	}

	err := handler(cfg, flag.Args())
	if err == os.ErrInvalid {
		utils.PrintUsage(os.Stderr, cmdUsage)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
	}
}
