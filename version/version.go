package version

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// Version represents the program version.
	Version string
	// Branch represents the repository branch name.
	Branch string
	// Revision represents the repository revision hash.
	Revision string
	// BuildDate represents the program build date.
	BuildDate string
	// Compiler represents the program compiler information.
	Compiler = fmt.Sprintf("%s (%s)", runtime.Version(), runtime.Compiler)
)

// Fprint writes the version information into a writer.
func Fprint(w io.Writer) {
	program := filepath.Base(os.Args[0])

	fmt.Fprintf(w, `%s
   Version:     %s
   Branch:      %s
   Revision:    %s
   Compiler:    %s
   Build date:  %s
`,
		program,
		Version,
		Branch,
		Revision,
		Compiler,
		BuildDate,
	)
}

// Print writes the version information to the standard output.
func Print() {
	Fprint(os.Stdout)
}
