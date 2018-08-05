package main

import (
	"fmt"
	"os"
	"strings"

	"facette.io/facette/version"
	"github.com/cosiner/flag"
	"github.com/mgutz/ansi"
)

type command struct {
	Address string `names:"-a, --address" usage:"Upstream socket address" default:"http://localhost:12003"`
	Help    bool   `names:"-h, --help" usage:"Display this help and exit"`
	Quiet   bool   `names:"-q, --quiet" usage:"Run in quiet mode"`
	Timeout int    `names:"-t, --timeout" usage:"Upstream connection timeout" default:"30"`
	Version bool   `names:"-V, --version" usage:"Display version information and exit"`

	Catalog catalogCommand `usage:"Manage catalog operations"`
	Library libraryCommand `usage:"Manage library operations"`
}

func (*command) Metadata() map[string]flag.Flag {
	return map[string]flag.Flag{"": {Usage: "Facette control utility"}}
}

var (
	cmd     command
	flagSet *flag.FlagSet
)

func init() {
	flagSet = flag.NewFlagSet(flag.Flag{}).ErrHandling(0)
	flagSet.StructFlags(&cmd)

	err := flagSet.Parse(os.Args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		flagSet.Help(false)
		os.Exit(2)
	} else if cmd.Version {
		version.Print()
		os.Exit(0)
	}
}

func main() {
	var err error

	// Add default scheme to address if none provided
	if !strings.HasPrefix(cmd.Address, "http://") && !strings.HasPrefix(cmd.Address, "https://") {
		cmd.Address = "http://" + cmd.Address
	}

	if cmd.Catalog.Enable {
		err = execCatalog()
	} else if cmd.Library.Enable {
		err = execLibrary()
	} else {
		flagSet.Help(false)
	}

	if err != nil {
		if err != errExecFailed {
			die(err)
		}

		os.Exit(1)
	}
}

func die(err error) {
	printError("%s", err)
	os.Exit(1)
}

func printError(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, ansi.Color("Error: %s\n", "red"), fmt.Sprintf(format, v...))
}
