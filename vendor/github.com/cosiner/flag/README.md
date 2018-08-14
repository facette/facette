# Flag
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/cosiner/flag) 
[![Build Status](https://travis-ci.org/cosiner/flag.svg?branch=master&style=flat)](https://travis-ci.org/cosiner/flag)
[![Coverage Status](https://coveralls.io/repos/github/cosiner/flag/badge.svg?style=flat)](https://coveralls.io/github/cosiner/flag)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosiner/flag?style=flat)](https://goreportcard.com/report/github.com/cosiner/flag)

Flag is a simple but powerful commandline flag parsing library for [Go](https://golang.org).

# Documentation
Documentation can be found at [Godoc](https://godoc.org/github.com/cosiner/flag)

# Features
* Support types: bool, string, all number types(except complex), slice of bool, string, number.
* Embed structure as subcommand.
* Multiple flag names, e.g. '-z, -gz, -gzip, --gz, --gzip'
* '-' to ensure next argument must be a flag, e.g. 
* '--' to ensure next argument must be a value, e.g. 'rm -- -a.go' to delete file '-a.go'
* '-!' to stop greedy-consumption for slice flags.
* Support '=', e.g. '-a=b', '-a=true'
* Support single bool flag, e.g. '-rm' is equal to '-rm=true'
* Support multiple single flags: e.g. '-zcf a.tgz' is equal to '-z -c -f a.tgz'
* Support '-I/usr/include' like format, the character next to '-' must be a alphabet, and the next next must not be.
* Support catch non-flag values, e.g. 'tar -zcf a.tgz a.go b.go' will catch the values ['a.go', 'b.go']
* Default value
* Select values
* Environment variable
* Duplicate flag names detect

# Parsing
* Flag/FlagSet
  * Names(tag: 'names'): split by ',', fully custom: short, long, with or without '-'/'--'.
  * Arglist(tag: 'arglist'): show commandline of flag or flag set, 
    E.g., `-input INPUT -output OUTPUT... -t 'tag list'`.
  * Usage(tag: 'usage'): the short help message for this flag or flag set, 
    E.g., `build       compile packages and dependencies`.
  * Desc(tag: 'desc'): long description for this flag or flag set,  it will be split to multiple lines 
    and format with same indents.
  * Ptr(field pointer for Flag, field 'Enable bool' for FlagSet): result pointer
  
* Flag (structure field)
  * Default(tag: 'default'): default value
  * Selects(tag: 'selects'): selectable values, must be slice.
  * Env(tag: 'env'): environment variable, only used when flag not appeared in arguments.
  * ValSep(tag: 'valsep'): slice value separator for environment variable's value,
  * ShowType(tag: 'showType'): show flag type in help message
  
* FlagSet (embed structure)
  * Expand(tag: 'expand'): always expand subset info in help message.
  * Version(tag: 'version'): app version, will be split to multiple lines and format with same indents.
  * ArgsPtr(field: 'Args'): pointer to accept all the last non-flag values, 
    nil if don't need and error will be reported automatically.
  
* FlagMeta
  Structure can implement the Metadata interface to update flag metadata instead write in structure tag, 
  it's designed for long messages.
   
  
# Example
## Flags
```Go
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

```
## Help message
```
tar is a tool for manipulate tape archives.

Usage:
      flag.test [FLAG]...

Version:
      version: v1.0.0
      commit: 10adf10dc10
      date:   2017-01-01 10:00:01

Description:
      tar creates and manipulates streaming archive files.  This implementation can extract
      from tar, pax, cpio, zip, jar, ar, and ISO 9660 cdrom images and can create tar, pax,
      cpio, ar, and shar archives.

Flags:
      -z, --gz     gzip format (bool)
            use gzip format
      -j, --bz     bzip2 format (bool)
      -J, --xz     xz format (bool)
      -c           create tar file (bool)
      -x           extract tar file (bool)
      -f           output file for create or input file for extract (string)
      -C           extract directory (string)
```

## FlagSet
```Go

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

func TestSubset(t *testing.T) {
	var g GoCmd

	set := NewFlagSet(Flag{})
	set.StructFlags(&g)
	set.Help(false)
	fmt.Println()
	build, _ := set.FindSubset("build")
	build.Help(false)
}
```
##Help Message
```
Go is a tool for managing Go source code.

Usage:
      flag.test command [argument]

Sets:
      build        compile packages and dependencies
      clean        remove object files
      doc          show documentation for package or symbol
      env          print Go environment information
      bug          start a bug report
      fix          run go tool fix on packages
      fmt          run gofmt on package sources
```
```
compile packages and dependencies

Usage:
      build [-o output] [-i] [build flags] [packages]

Description:
      Build compiles the packages named by the import paths,
      along with their dependencies, but it does not install the results.
      ...
      The build flags are shared by the build, clean, get, install, list, run,
      and test commands:

Flags:
      -a                  
            force rebuilding of packages that are already up-to-date.
      -race               
            enable data race detection.
            Supported only on linux/amd64, freebsd/amd64, darwin/amd64 and windows/amd64.
      -o output           
            only allowed when compiling a single package

      -ldflags 'flag list'
            rguments to pass on each go tool link invocation.
```
# LICENSE
MIT.
