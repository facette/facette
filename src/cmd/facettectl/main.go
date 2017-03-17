package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
)

const (
	defaultAddress = "http://localhost:12003"
	defaultTimeout = "30"
)

var (
	version   string
	buildDate string
	buildHash string

	upstreamAddress string
	upstreamTimeout int

	verbose bool
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "Facette control utility.")
	app.HelpFlag.Short('h')

	// Global
	flagAddress := app.Flag("address", "Set upstream socket address.").Short('a').Default(defaultAddress).String()
	flagTimeout := app.Flag("timeout", "Set upstream connection timeout.").Short('t').Default(defaultTimeout).Int()
	flagVerbose := app.Flag("verbose", "Run in verbose mode.").Short('v').Bool()

	// Version
	version := app.Command("version", "Display version and support information.")

	// Backend
	library := app.Command("library", "Manage library operations.")

	libraryDump := library.Command("dump", "Dump data from library.")
	libraryDumpOutput := libraryDump.Flag("output", "Set dump output file path.").Short('o').String()

	libraryRestore := library.Command("restore", "Restore data from dump into library.")
	libraryRestoreInput := libraryRestore.Flag("input", "Set dump input file path.").Short('i').Required().String()
	libraryRestoreMerge := libraryRestore.Flag("merge", "Merge data with existing library.").Short('m').Bool()

	// Parse command-line
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	upstreamAddress = *flagAddress
	upstreamTimeout = *flagTimeout

	verbose = *flagVerbose

	switch command {
	case version.FullCommand():
		execVersion()

	case libraryDump.FullCommand():
		execBackupDump(*libraryDumpOutput)

	case libraryRestore.FullCommand():
		execBackupRestore(*libraryRestoreInput, *libraryRestoreMerge)
	}
}

func die(format string, v ...interface{}) {
	printError(format, v...)
	os.Exit(1)
}

func printError(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", fmt.Sprintf(format, v...))
}
