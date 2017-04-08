package flag

import "fmt"

type Tar struct {
	GZ          bool     `names:"-z, --gz" usage:"gzip format"`
	BZ          bool     `names:"-j, --bz" usage:"bzip2 format"`
	XZ          bool     `names:"-J, --xz" usage:"xz format"`
	Create      bool     `names:"-c" usage:"create tar file"`
	Extract     bool     `names:"-x" usage:"extract tar file"`
	File        string   `names:"-f" usage:"output file for create or input file for extract"`
	Directory   string   `names:"-C" usage:"extract directory"`
	SourceFiles []string `args:"true"`
}

func (t *Tar) Metadata() map[string]Flag {
	const (
		usage   = "tar is a tool for manipulate tape archives."
		version = `
			version: v1.0.0
			commit: 10adf10dc10
			date:   2017-01-01 10:00:01
		`
		desc = `
		tar creates and manipulates streaming archive files.  This implementation can extract
		from tar, pax, cpio, zip, jar, ar, and ISO 9660 cdrom images and can create tar, pax,
		cpio, ar, and shar archives.
		`
	)
	return map[string]Flag{
		"": {
			Usage:   usage,
			Version: version,
			Desc:    desc,
		},
		"--gz": {
			Desc: "use gzip format",
		},
	}
}

func ExampleFlagSet_ParseStruct() {
	var tar Tar

	NewFlagSet(Flag{}).ParseStruct(&tar, "tar", "-zcf", "a.tgz", "a.go", "b.go")
	fmt.Println(tar.GZ)
	fmt.Println(tar.Create)
	fmt.Println(tar.File)
	fmt.Println(tar.SourceFiles)

	// Output:
	// true
	// true
	// a.tgz
	// [a.go b.go]
}

type GoCmd struct {
	Build struct {
		Enable  bool
		Already bool   `names:"-a" important:"1" desc:"force rebuilding of packages that are already up-to-date."`
		Race    bool   `important:"1" desc:"enable data race detection.\nSupported only on linux/amd64, freebsd/amd64, darwin/amd64 and windows/amd64."`
		Output  string `names:"-o" arglist:"output" important:"1" desc:"only allowed when compiling a single package"`

		LdFlags  string   `names:"-ldflags" arglist:"'flag list'" desc:"rguments to pass on each go tool link invocation."`
		Packages []string `args:"true"`
	} `usage:"compile packages and dependencies"`
	Clean struct {
		Enable bool
	} `usage:"remove object files"`
	Doc struct {
		Enable bool
	} `usage:"show documentation for package or symbol"`
	Env struct {
		Enable bool
	} `usage:"print Go environment information"`
	Bug struct {
		Enable bool
	} `usage:"start a bug report"`
	Fix struct {
		Enable bool
	} `usage:"run go tool fix on packages"`
	Fmt struct {
		Enable bool
	} `usage:"run gofmt on package sources"`
}

func (*GoCmd) Metadata() map[string]Flag {
	return map[string]Flag{
		"": {
			Usage:   "Go is a tool for managing Go source code.",
			Arglist: "command [argument]",
		},
		"build": {
			Arglist: "[-o output] [-i] [build flags] [packages]",
			Desc: `
		Build compiles the packages named by the import paths,
		along with their dependencies, but it does not install the results.
		...
		The build flags are shared by the build, clean, get, install, list, run,
		and test commands:
			`,
		},
	}
}
