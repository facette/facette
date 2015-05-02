// Package cmd provides common helper functions to command line binaries.
package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
)

// PrintUsage prettifies the output of command-line usage.
func PrintUsage(output io.Writer, usage string) {
	fmt.Fprintf(output, usage, path.Base(os.Args[0]))
	fmt.Fprint(output, "\n\nOptions:\n")

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(output, "   -%s  %s\n", f.Name, f.Usage)
	})

	os.Exit(2)
}

// PrintVersion prettifies the output of command-line usage.
func PrintVersion(version string) {
	fmt.Printf("%s version %s\nGo version: %s (%s)\n",
		path.Base(os.Args[0]),
		version,
		runtime.Version(),
		runtime.Compiler,
	)
}
