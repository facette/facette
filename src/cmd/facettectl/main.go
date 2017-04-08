package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosiner/flag"
)

type command struct {
	Address string `names:"-a, --address" usage:"upstream socket address" default:"http://localhost:12003"`
	Help    bool   `names:"-h, --help" usage:"display this help and exit"`
	Timeout int    `names:"-t, --timeout" usage:"upstream connection timeout" default:"30"`
	Version bool   `names:"-V, --version" usage:"display version information and exit"`
	Quiet   bool   `names:"-q, --quiet" usage:"run in quiet mode"`

	Library libraryCommand
}

func (*command) Metadata() map[string]flag.Flag {
	return map[string]flag.Flag{
		"": {
			Usage: "Facette control utility",
		},
		"library": {
			Usage: "manage library operations",
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
	flagSet := flag.NewFlagSet(flag.Flag{}).ErrHandling(0)
	flagSet.StructFlags(&cmd)

	if err := flagSet.Parse(os.Args...); err != nil {
		fmt.Printf("Error: %s\n", err)
		flagSet.Help(false)
		os.Exit(1)
	}

	// Add default scheme to address if none provided
	if !strings.HasPrefix(cmd.Address, "http://") && !strings.HasPrefix(cmd.Address, "https://") {
		cmd.Address = "http://" + cmd.Address
	}

	if cmd.Version {
		execVersion()
		os.Exit(0)
	} else if cmd.Library.Dump.Enable {
		execLibraryDump()
	} else if cmd.Library.Restore.Enable {
		execLibraryRestore()
	} else if cmd.Library.Enable {
		library, _ := flagSet.FindSubset("library")
		library.Help(false)
	} else {
		flagSet.Help(false)
	}
}

func die(format string, v ...interface{}) {
	printError(format, v...)
	os.Exit(1)
}

func printError(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", fmt.Sprintf(format, v...))
}
