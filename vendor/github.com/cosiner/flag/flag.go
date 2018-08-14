package flag

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// Flag represents the state of a flag
type Flag struct {
	// Common fields used for Flag and FlagSet
	Names     string      // names, split by ','
	Arglist   string      // arguments list
	Usage     string      // short usage message
	Desc      string      // long description, can be multiple lines
	descLines []string    // parsed description lines
	Important bool        // important flag, will be print before unimportant flags
	Ptr       interface{} // value pointer

	// For Flag
	Default  interface{} // default value
	Selects  interface{} // select value
	Env      string      // environment name
	ValSep   string      // environment value separator
	ShowType bool        // show flag type in help message

	// For FlagSet
	Version      string    // version, can be multiple lines
	versionLines []string  // parsed version lines
	Expand       bool      // expand subsets in help message
	ArgsPtr      *[]string // NArgs pointer
}

// Metadata can be implemented by structure to update flag metadata.
type Metadata interface {
	// Metadata return the metadata map to be updated.
	// The return value is a map of children and metadata.
	Metadata() map[string]Flag
}

// NoFlag can be implemented by structure or field to prevent from parsing
type NoFlag interface {
	// NoFlag method identify the field should not be parsed
	NoFlag()
}

// ErrorHandling is the error handling way when error occurred when register/scan/resolve.
//
// ErrorHandling can be set of basic handling way, the way sequence is ErrPanic, ErrPrint, ErrExit.
type ErrorHandling uint8

const (
	// ErrPanic panic goroutine with the error
	ErrPanic ErrorHandling = 1 << iota
	// ErrPrint print the error to stdout
	ErrPrint
	// ErrExit exit process
	ErrExit

	// DefaultErrorHandling includes ErrPrint and ErrExit
	DefaultErrorHandling = ErrPrint | ErrExit
)

func (e ErrorHandling) do(eh ErrorHandling) bool {
	return e&eh != 0
}

func (e ErrorHandling) handle(err error) error {
	if err == nil {
		return nil
	}

	if e.do(ErrPanic) {
		panic(err)
	}
	if e.do(ErrPrint) {
		fmt.Fprintln(os.Stderr, err)
	}
	if e.do(ErrExit) {
		os.Exit(2)
	}
	return err
}

// FlagSet is a set of flags and other subsets.
type FlagSet struct {
	self Flag

	flags       []Flag
	flagIndexes map[string]int

	subsets       []FlagSet
	subsetIndexes map[string]int

	errorHandling   ErrorHandling
	noHelpFlag      bool
	helpFlagDefined bool
}

// NewFlagSet create a new flagset
func NewFlagSet(flag Flag) *FlagSet {
	if flag.Names == "" {
		flag.Names = filepath.Base(os.Args[0])
	}
	return newFlagSet(flag)
}

func newFlagSet(flag Flag) *FlagSet {
	defaultReguster.cleanFlag(&flag)
	return &FlagSet{
		self:          flag,
		flagIndexes:   make(map[string]int),
		subsetIndexes: make(map[string]int),
		errorHandling: DefaultErrorHandling,
	}
}

func (f *FlagSet) searchFlag(name string) *Flag {
	index, has := f.flagIndexes[name]
	if !has {
		return nil
	}
	return &f.flags[index]
}

func (f *FlagSet) isFlag(name string) bool {
	_, has := f.flagIndexes[name]
	return has
}

func (f *FlagSet) isSubset(name string) bool {
	_, has := f.subsetIndexes[name]
	return has
}

func (f *FlagSet) isFlagOrSubset(name string) bool {
	return f.isFlag(name) || f.isSubset(name)
}

// UpdateMeta update flag metadata by the children identifier, only Desc, Arglist,
// Usage and Version will be updated.
// The children identifier will be split by ',', if children is empty, it update
// itself.
//
// E.g., "tool, cover, -html": Flag{Usage:"display coverage in html"}
func (f *FlagSet) UpdateMeta(children string, meta Flag) error {
	return defaultReguster.updateMeta(f, children, meta)
}

// ErrHandling change the way of error handling
func (f *FlagSet) ErrHandling(ehs ...ErrorHandling) *FlagSet {
	var e ErrorHandling
	for _, eh := range ehs {
		e |= eh
	}
	f.errorHandling = e
	for i := range f.subsets {
		f.subsets[i].ErrHandling(f.errorHandling)
	}
	return f
}

// NeedHelpFlag toggle help flags auto-defining. By default, if there is no help flag, it will
// be defined when Parse is called.
func (f *FlagSet) NeedHelpFlag(need bool) *FlagSet {
	f.noHelpFlag = !need
	for i := range f.subsets {
		f.subsets[i].NeedHelpFlag(need)
	}
	return f
}

// Flag add a flag to current flagset, it should not duplicate with parent/current/children levels' flag or flagset.
func (f *FlagSet) Flag(flag Flag) error {
	return f.errorHandling.handle(defaultReguster.registerFlag(nil, f, flag))
}

// Subset add a flagset to current flagset and return the subset
func (f *FlagSet) Subset(flag Flag) (*FlagSet, error) {
	child, err := defaultReguster.registerSet(nil, f, flag)
	return child, f.errorHandling.handle(err)
}

// FindSubset search flagset by the children identifier, children is subset names split by ','.
func (f *FlagSet) FindSubset(children string) (*FlagSet, error) {
	_, subset, err := defaultReguster.searchChildrenFlag(f, children)
	if subset == nil && err == nil {
		err = newErrorf(errFlagNotFound, "subset %s is not found", children)
	}
	return subset, err
}

// FindFlag search flag by the children identifier, children is set subset/flag names split by ','.
func (f *FlagSet) FindFlag(children string) (*Flag, error) {
	flag, _, err := defaultReguster.searchChildrenFlag(f, children)
	if flag == nil && err == nil {
		err = newErrorf(errFlagNotFound, "flag %s is not found", children)
	}
	return flag, err
}

// StructFlags parse the structure pointer and add exported fields to flagset.
// if parent is not nil, it will checking duplicate flags with parent.
func (f *FlagSet) StructFlags(val interface{}, parent ...*FlagSet) error {
	var p *FlagSet
	if len(parent) > 0 {
		p = parent[0]
	}
	return f.errorHandling.handle(defaultReguster.registerStructure(p, f, val))
}

// Parse parse arguments, if empty, os.Args will be used.
func (f *FlagSet) Parse(args ...string) error {
	if len(args) == 0 {
		args = os.Args
	}
	if !f.noHelpFlag && !f.helpFlagDefined {
		err := defaultReguster.registerHelpFlags(nil, f)
		if err != nil {
			return f.errorHandling.handle(err)
		}
	}
	var (
		s scanner
		r resolver
	)
	s.scan(f, args)
	err := r.resolve(f, &s.Result)
	if err != nil {
		return f.errorHandling.handle(err)
	}

	show, verbose := defaultReguster.helpFlagValues(f)
	if show {
		fmt.Print(r.LastSet.ToString(verbose))
		os.Exit(0)
	}
	return nil
}

// ParseStruct is the combination of StructFlags and Parse
func (f *FlagSet) ParseStruct(val interface{}, args ...string) error {
	err := f.StructFlags(val)
	if err != nil {
		return err
	}
	return f.Parse(args...)
}

// ToString return help message, if verbose, all subset will be expand.
func (f *FlagSet) ToString(verbose bool) string {
	var buf bytes.Buffer
	(&writer{
		buf:          &buf,
		isTop:        true,
		forceVerbose: verbose,
	}).writeSet(f)
	return buf.String()
}

// Help print help message to stdout
func (f *FlagSet) Help(verbose bool) {
	fmt.Print(f.ToString(verbose))
}

// Reset reset values of each registered flags.
func (f *FlagSet) Reset() {
	var r resolver
	r.reset(f)
}

var (
	// Commandline is the default FlagSet instance.
	Commandline = NewFlagSet(Flag{})
)

// ParseStruct is short way of Commandline.ParseStruct
func ParseStruct(val interface{}, args ...string) error {
	return Commandline.ParseStruct(val, args...)
}

// Help is the short way of Commandline.Help
func Help(verbose bool) {
	Commandline.Help(verbose)
}
