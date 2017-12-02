package flag

import (
	"go/ast"
	"reflect"
	"strings"
	"unicode"
)

type register struct {
}

var defaultReguster register

func (r register) addIndexes(indexes map[string]int, keys []string, index int) {
	for _, key := range keys {
		indexes[key] = index
	}
}

func (r register) findDuplicates(parent, set *FlagSet, names []string) []string {
	var duplicates []string
	for _, name := range names {
		if set.isFlagOrSubset(name) || (parent != nil && parent.isFlagOrSubset(name)) {
			duplicates = append(duplicates, name)
			continue
		}
		for i := range set.subsets {
			if set.subsets[i].isFlagOrSubset(name) {
				duplicates = append(duplicates, name)
				break
			}
		}
	}
	return duplicates
}

const (
	flagNameSeparatorForSplit = ","
	flagNameSeparatorForJoin  = ", "
)

func (r register) joinFlagNames(names []string) string {
	return strings.Join(names, flagNameSeparatorForJoin)
}

func (r register) cleanFlagNames(names string) ([]string, string) {
	ns := splitAndTrimSpace(names, flagNameSeparatorForSplit)
	return ns, r.joinFlagNames(ns)
}

func (r register) cleanFlag(flag *Flag) {
	if flag.ValSep == "" {
		flag.ValSep = ","
	}
	r.updateFlagDesc(flag, flag.Desc)
	r.updateFlagVersion(flag, flag.Version)
}

func (r register) registerFlag(parent, set *FlagSet, flag Flag) error {
	refval := reflect.ValueOf(flag.Ptr)
	if refval.Kind() != reflect.Ptr {
		return newErrorf(errNonPointer, "illegal flag pointer: %s", flag.Names)
	}
	if typeName(flag.Ptr) == "" {
		return newErrorf(errInvalidType, "unsupported flag type: %s", flag.Names)
	}
	if flag.Default != nil {
		var compatible bool

		refdef := reflect.ValueOf(flag.Default)
		if isSlicePtr(flag.Ptr) {
			compatible = refdef.Kind() == reflect.Slice
			compatible = compatible && isKindCompatible(sliceElemKind(refval.Elem()), sliceElemKind(refdef))
		} else {
			compatible = isKindCompatible(refval.Elem().Kind(), refdef.Kind())
		}
		if !compatible {
			return newErrorf(errInvalidType, "incompatible default value type: %s", flag.Names)
		}
	}
	if flag.Selects != nil {
		var err error
		flag.Selects, err = parseSelectsValue(flag.Ptr, flag.Selects)
		if err != nil {
			return newErrorf(errInvalidSelects, "%s: %s", flag.Names, err.Error())
		}
	}

	ns, names := r.cleanFlagNames(flag.Names)
	if duplicates := r.findDuplicates(parent, set, ns); len(duplicates) > 0 {
		return newErrorf(errDuplicateFlagRegister, "duplicate flags with parent/self/childs: %s->%s, %v", parent.self.Names, set.self.Names, duplicates)
	}

	flag.Names = names
	r.cleanFlag(&flag)

	set.flags = append(set.flags, flag)
	r.addIndexes(set.flagIndexes, ns, len(set.flags)-1)
	return nil
}

func (r register) checkSubsetValid(flag *Flag) error {
	if flag.Names == "" {
		return newErrorf(errInvalidNames, "subset names should not be empty")
	}
	return nil
}

func (r register) registerSet(parent, set *FlagSet, flag Flag) (*FlagSet, error) {
	var ns []string

	ns, flag.Names = r.cleanFlagNames(flag.Names)
	err := r.checkSubsetValid(&flag)
	if err != nil {
		return nil, err
	}

	if duplicates := r.findDuplicates(parent, set, ns); len(duplicates) > 0 {
		return nil, newErrorf(errDuplicateFlagRegister, "duplicate subset name: %v", duplicates)
	}

	child := newFlagSet(flag)
	child.self.Default = false
	child.errorHandling = set.errorHandling

	set.subsets = append(set.subsets, *child)
	r.addIndexes(set.subsetIndexes, ns, len(set.subsets)-1)
	return &set.subsets[len(set.subsets)-1], nil
}

func (r register) registerStructure(parent, set *FlagSet, st interface{}) error {
	// parent is used to checking duplicate flags and indicate that subset must has a 'Enable' field
	const (
		tagNames     = "names"
		tagArglist   = "arglist"
		tagUsage     = "usage"
		tagDesc      = "desc"
		tagVersion   = "version"
		tagImportant = "important"

		tagEnv      = "env"
		tagValsep   = "valsep"
		tagDefault  = "default"
		tagSelects  = "selects"
		tagExpand   = "expand"
		tagArgs     = "args"
		tagShowType = "showType"

		fieldSubsetEnable = "Enable"
		fieldArgs         = "Args"
	)

	refval := reflect.ValueOf(st)
	if refval.Kind() != reflect.Ptr || refval.Elem().Kind() != reflect.Struct {
		return newErrorf(errNonPointer, "not pointer of structure")
	}
	refval = refval.Elem()
	reftyp := refval.Type()
	numfield := refval.NumField()
	for i := 0; i < numfield; i++ {
		fieldType := reftyp.Field(i)
		if !ast.IsExported(fieldType.Name) {
			continue
		}

		fieldVal := refval.Field(i)

		args := fieldType.Tag.Get(tagArgs)
		isArgs, err := parseBool(args, "false")
		if err != nil {
			return newErrorf(errInvalidValue, "non-bool tag args value: %s.%s %s", set.self.Names, fieldType.Name, args)
		}
		if fieldType.Name == fieldArgs || isArgs {
			if set.self.ArgsPtr != nil {
				return newErrorf(errDuplicateFlagRegister, "duplicate args field: %s", set.self.Names)
			}
			if _, ok := fieldVal.Interface().([]string); !ok {
				return newErrorf(errInvalidType, "invalid %s:Args field type, expect []string", set.self.Names)
			}
			set.self.ArgsPtr = fieldVal.Addr().Interface().(*[]string)
			continue
		}

		ptr := fieldVal.Addr().Interface()
		if fieldType.Name == fieldSubsetEnable {
			if fieldType.Type.Kind() != reflect.Bool {
				return newErrorf(errInvalidType, "illegal type of field '%s', expect bool", fieldSubsetEnable)
			}
			if set.self.Ptr == nil {
				set.self.Ptr = ptr
			}
			continue
		}

		var (
			names     = fieldType.Tag.Get(tagNames)
			usage     = fieldType.Tag.Get(tagUsage)
			desc      = fieldType.Tag.Get(tagDesc)
			version   = fieldType.Tag.Get(tagVersion)
			arglist   = fieldType.Tag.Get(tagArglist)
			important = fieldType.Tag.Get(tagImportant)
		)
		if names == "-" {
			continue
		}
		_, ok := ptr.(NoFlag)
		if ok {
			continue
		}

		importantVal, err := parseBool(important, "false")
		if err != nil {
			return newErrorf(errInvalidValue, "invalid tag import value: %s.%s %s", set.self.Names, fieldType.Name, important)
		}
		if fieldVal.Kind() != reflect.Struct {
			var (
				env      = fieldType.Tag.Get(tagEnv)
				def      = fieldType.Tag.Get(tagDefault)
				valsep   = fieldType.Tag.Get(tagValsep)
				selects  = fieldType.Tag.Get(tagSelects)
				showType = fieldType.Tag.Get(tagShowType)
			)
			if names == "" {
				names = "-" + unexportedName(fieldType.Name)
			}
			if valsep == "" {
				valsep = ","
			}
			if typeName(ptr) == "" {
				continue
			}
			defVal, err := parseDefault(def, valsep, ptr)
			if err != nil {
				return err
			}
			selectsVal, err := parseSelectsString(selects, valsep, ptr)
			if err != nil {
				return err
			}
			showTypeVal, err := parseBool(showType, "false")
			if err != nil {
				return newErrorf(errInvalidValue, "invalid tag showType value: %s.%s %s", set.self.Names, fieldType.Name, showType)
			}
			err = r.registerFlag(parent, set, Flag{
				Names:     names,
				Arglist:   arglist,
				Usage:     usage,
				Desc:      desc,
				Version:   version,
				Important: importantVal,
				ShowType:  showTypeVal,

				Ptr:     ptr,
				Env:     env,
				ValSep:  valsep,
				Default: defVal,
				Selects: selectsVal,
			})
			if err != nil {
				return err
			}
		} else {
			expand := fieldType.Tag.Get(tagExpand)
			if names == "" {
				names = unexportedName(fieldType.Name)
			}
			expandVal, err := parseBool(expand, "false")
			if err != nil {
				return newErrorf(errInvalidValue, "parse expand value %s as bool failed", expand)
			}
			child, err := r.registerSet(parent, set, Flag{
				Names:     names,
				Arglist:   arglist,
				Usage:     usage,
				Desc:      desc,
				Version:   version,
				Important: importantVal,
				Expand:    expandVal,
			})
			if err != nil {
				return err
			}
			err = r.registerStructure(set, child, fieldVal.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}
	if md, ok := st.(Metadata); ok {
		for children, meta := range md.Metadata() {
			err := r.updateMeta(set, children, meta)
			if err != nil {
				return err
			}
		}
	}
	if parent != nil && set.self.Ptr == nil {
		return newErrorf(errInvalidStructure, "child structure must has a 'Enable' field")
	}
	return nil
}

func (r register) registerBoolFlags(parent, set *FlagSet, names []string, usage string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	if duplicates := r.findDuplicates(parent, set, names); len(duplicates) > 0 {
		return false, nil
	}
	var value bool
	err := r.registerFlag(parent, set, Flag{
		Ptr:   &value,
		Names: strings.Join(names, ","),
		Usage: usage,
	})
	return err == nil, err
}

func (r register) registerHelpFlags(parent, set *FlagSet) error {
	registered, err := r.registerBoolFlags(parent, set, []string{"-h", "--help"}, "show help")
	if err == nil && registered && len(set.subsets) > 0 {
		_, err = r.registerBoolFlags(parent, set, []string{"-v", "--verbose"}, "show verbose help")
	}
	return err
}

func (r register) boolFlagVal(set *FlagSet, flag string) (val, has bool) {
	index, has := set.flagIndexes[flag]
	if !has {
		return false, false
	}
	return *set.flags[index].Ptr.(*bool), true
}

func (r register) helpFlagValues(set *FlagSet) (show, verbose bool) {
	var has bool
	show, has = r.boolFlagVal(set, "-h")
	if show {
		if has {
			verbose, _ = r.boolFlagVal(set, "-v")
		}
	}
	return
}

func (r register) prefixSpaceCount(s string) int {
	var c int
	for _, r := range s {
		if unicode.IsSpace(r) {
			c++
		} else {
			break
		}
	}
	return c
}

func (r register) splitLines(line string) []string {
	var (
		lines    = strings.Split(line, "\n")
		begin    = -1
		end      = -1
		minSpace = -1
	)
	for i := range lines {
		if strings.TrimSpace(lines[i]) != "" {
			if begin < 0 {
				begin = i
			}
			end = i + 1
			c := r.prefixSpaceCount(lines[i])
			if minSpace > c || minSpace < 0 {
				minSpace = c
			}
		}
	}
	if begin < 0 {
		return nil
	}
	lines = lines[begin:end]
	for i := range lines {
		line := lines[i]
		if len(line) >= minSpace {
			line = line[minSpace:]
		}
		lines[i] = line
	}
	return lines
}

func (r register) updateFlagDesc(flag *Flag, desc string) {
	flag.Desc = desc
	flag.descLines = r.splitLines(flag.Desc)
}

func (r register) updateFlagVersion(flag *Flag, version string) {
	flag.Version = version
	flag.versionLines = r.splitLines(flag.Version)
}

func (r register) searchChildrenFlag(set *FlagSet, children string) (*Flag, *FlagSet, error) {
	var (
		currSet  = set
		currFlag *Flag

		sections = splitAndTrimSpace(children, flagNameSeparatorForSplit)
		last     = len(sections) - 1
	)
	for i, sec := range sections {
		index, has := currSet.subsetIndexes[sec]
		if has {
			currSet = &currSet.subsets[index]
			continue
		}
		if i != last {
			return nil, nil, newErrorf(errFlagNotFound, "subset/flag %s is not found", sec)
		}
		index, has = currSet.flagIndexes[sec]
		if !has {
			return nil, nil, newErrorf(errFlagNotFound, "subset/flag %s is not found", sec)
		}
		currFlag = &currSet.flags[index]
	}
	if currFlag == nil {
		currFlag = &currSet.self
	}
	return currFlag, currSet, nil
}

func (r register) updateMeta(set *FlagSet, children string, meta Flag) error {
	flag, subset, err := r.searchChildrenFlag(set, children)
	if err != nil {
		return err
	}
	if meta.Desc != "" {
		flag.Desc = meta.Desc
	}
	if subset != nil && meta.Version != "" {
		flag.Version = meta.Version
	}
	if meta.Arglist != "" {
		flag.Arglist = meta.Arglist
	}
	if meta.Usage != "" {
		flag.Usage = meta.Usage
	}
	r.cleanFlag(flag)
	return nil
}
