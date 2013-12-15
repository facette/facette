package main

import (
	"facette/server"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"syscall"
)

var (
	flagConfig string
	flagDebug  int
	flagHelp   bool
)

func printUsage(output io.Writer) {
	fmt.Fprintf(output, "Usage: %s [OPTIONS] -c FILE\n\nOptions:\n", path.Base(os.Args[0]))

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(output, "   -%s  %s\n", f.Name, f.Usage)
	})

	os.Exit(2)
}

func init() {
	flag.StringVar(&flagConfig, "c", "", "configuration file path")
	flag.IntVar(&flagDebug, "d", 0, "debugging level")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.Usage = func() { printUsage(os.Stderr) }
	flag.Parse()

	if flagHelp {
		printUsage(os.Stdout)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		printUsage(os.Stderr)
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
