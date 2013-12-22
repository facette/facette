package main

import (
	"facette/common"
	"facette/server"
	"facette/utils"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	flagConfig string
	flagDebug  int
	flagHelp   bool
)

func init() {
	flag.StringVar(&flagConfig, "c", common.DefaultConfigFile, "configuration file path")
	flag.IntVar(&flagDebug, "d", 0, "debugging level")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.Usage = func() { utils.PrintUsage(os.Stderr) }
	flag.Parse()

	if flagHelp {
		utils.PrintUsage(os.Stdout)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		utils.PrintUsage(os.Stderr)
	}
}

func main() {
	var (
		err      error
		sigChan  chan os.Signal
		instance *server.Server
	)

	// Create new server instance and load configuration
	if instance, err = server.NewServer(flagDebug); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	} else if err = instance.LoadConfig(flagConfig); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	// Reload server configuration on SIGHUP
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	go func() {
		for _ = range sigChan {
			instance.Reload()
		}
	}()

	// Run instance
	if err = instance.Run(); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
}
