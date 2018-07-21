package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosiner/flag"
	"github.com/mgutz/ansi"
)

type command struct {
	Address string `names:"-a, --address" usage:"Upstream socket address" default:"http://localhost:12003"`
	Help    bool   `names:"-h, --help" usage:"Display this help and exit"`
	Timeout int    `names:"-t, --timeout" usage:"Upstream connection timeout" default:"30"`
	Version bool   `names:"-V, --version" usage:"Display version information and exit"`
	Quiet   bool   `names:"-q, --quiet" usage:"Run in quiet mode"`

	Catalog catalogCommand `usage:"Manage catalog operations"`
	Library libraryCommand `usage:"Manage library operations"`
}

func (*command) Metadata() map[string]flag.Flag {
	return map[string]flag.Flag{
		"": {
			Usage: "Facette control utility",
		},
	}
}

var (
	version   string
	buildDate string
	buildHash string

	cmd     command
	flagSet *flag.FlagSet
)

func main() {
	var err error

	flagSet = flag.NewFlagSet(flag.Flag{}).ErrHandling(0)
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
	} else if cmd.Catalog.Enable {
		err = execCatalog()
	} else if cmd.Library.Enable {
		err = execLibrary()
	} else {
		flagSet.Help(false)
	}

	if err != nil {
		if err != errExecFailed {
			die("%s", err)
		}

		os.Exit(1)
	}
}

func die(format string, v ...interface{}) {
	printError(format, v...)
	os.Exit(1)
}

func printError(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, ansi.Color("Error: %s\n", "red"), fmt.Sprintf(format, v...))
}
