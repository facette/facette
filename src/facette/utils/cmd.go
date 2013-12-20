package utils

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
)

// PrintUsage prettifies the output of command-line usage.
func PrintUsage(output io.Writer) {
	fmt.Fprintf(output, "Usage: %s [OPTIONS] -c FILE\n\nOptions:\n", path.Base(os.Args[0]))

	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(output, "   -%s  %s\n", f.Name, f.Usage)
	})

	os.Exit(2)
}
