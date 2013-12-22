package utils

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
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
