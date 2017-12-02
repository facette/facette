package flag

import (
	"fmt"
	"os"
	"reflect"
)

var envParser = os.Getenv

type resolver struct {
	LastSet *FlagSet
}

func (r *resolver) fromDefault(f *Flag) []string {
	if !isSlicePtr(f.Ptr) {
		return []string{fmt.Sprint(f.Default)}
	}

	refval := reflect.ValueOf(f.Default)
	vals := make([]string, 0, refval.Len())
	for i, l := 0, refval.Len(); i < l; i++ {
		val := fmt.Sprint(refval.Index(i).Interface())
		if val != "" {
			vals = append(vals, val)
		}
	}
	return vals
}

func (r *resolver) fromEnv(f *Flag) []string {
	val := envParser(f.Env)
	if val == "" {
		return nil
	}

	var vals []string
	if isSlicePtr(f.Ptr) {
		vals = splitAndTrimSpace(val, f.ValSep)
	} else {
		vals = []string{val}
	}
	return vals
}

func (r *resolver) applyVals(f *Flag, vals ...string) error {
	for _, val := range vals {
		err := applyValToPtr(f.Names, f.Ptr, val, f.Selects)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *resolver) applyEnvOrDefault(f *FlagSet, applied map[*Flag]bool) error {
	for i := range f.flags {
		flag := &f.flags[i]
		if applied[flag] {
			continue
		}
		applied[flag] = true

		var vals []string
		if flag.Env != "" {
			vals = r.fromEnv(flag)
		}
		if len(vals) == 0 && flag.Default != nil {
			vals = r.fromDefault(flag)
		}
		err := r.applyVals(flag, vals...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *resolver) resolveFlags(f *FlagSet, context []string, args []argument) error {
	var (
		applied  = make(map[*Flag]bool)
		flag     *Flag
		argCount int
		err      error

		applyLastFlag = func() error {
			if argCount != 1 { // applied or no arg
				return nil
			}

			if isBoolPtr(flag.Ptr) {
				applied[flag] = true
				return r.applyVals(flag, "true")
			}
			return newErrorf(errStandaloneFlag, "standalone flag without values: %v.%s", context, flag.Names)
		}
		hasFlag = func(args []argument) bool {
			for i := range args {
				if args[i].Type == argumentFlag {
					return true
				}
			}
			return false
		}
		appendRemainArgs = func(args []argument) error {
			if f.self.ArgsPtr == nil || hasFlag(args[1:]) {
				return newErrorf(errStandaloneValue, "standalone value without flag: %v %s", context, args[0].Value)
			}
			slice := *f.self.ArgsPtr
			for i := range args {
				slice = append(slice, args[i].Value)
			}
			*f.self.ArgsPtr = slice
			return nil
		}
	)

	for i, arg := range args {
		switch arg.Type {
		case argumentFlag:
			err = applyLastFlag()
			if err != nil {
				return err
			}

			flag = f.searchFlag(arg.Value)
			if flag == nil {
				return newErrorf(errFlagNotFound, "unsupported flag: %v.%s", context, arg.Value)
			}
			if applied[flag] && !isSlicePtr(flag.Ptr) {
				return newErrorf(errDuplicateFlagParsed, "duplicate flag: %v.%s", context, flag.Names)
			}
			argCount = 1
		case argumentStopConsumption:
			err = applyLastFlag()
			if err != nil {
				return err
			}
			flag = nil
		case argumentValue:
			if flag == nil {
				err = appendRemainArgs(args[i:])
				if err != nil {
					return err
				}
				goto END
			}
			if argCount == 1 {
				applied[flag] = true
			} else if !isSlicePtr(flag.Ptr) {
				err = appendRemainArgs(args[i:])
				if err != nil {
					return err
				}
				goto END
			}

			argCount++
			err = r.applyVals(flag, arg.Value)
			if err != nil {
				return err
			}
		default:
			panic("unreachable")
		}
	}
	err = applyLastFlag()
	if err != nil {
		return err
	}
END:
	return r.applyEnvOrDefault(f, applied)
}

func (r *resolver) resolveSet(f *FlagSet, context []string, args *scanArgs) (lastSubset *FlagSet, err error) {
	context = append(context, f.self.Names)
	err = r.resolveFlags(f, context, args.Flags[1:])
	if err != nil {
		return nil, err
	}
	for sub, subArgs := range args.Sets {
		set := &f.subsets[f.subsetIndexes[sub]]
		err = r.applyVals(&set.self, "true")
		if err != nil {
			return nil, err
		}

		last, err := r.resolveSet(set, context, subArgs)
		if err != nil {
			return nil, err
		}
		if sub == args.FirstSubset {
			lastSubset = last
			if lastSubset == nil {
				lastSubset = set
			}
		}
	}

	if lastSubset == nil {
		lastSubset = f
	}
	return lastSubset, nil
}

func (r *resolver) resolve(f *FlagSet, args *scanArgs) error {
	var err error
	r.LastSet, err = r.resolveSet(f, nil, args)
	return err
}

func (r *resolver) reset(f *FlagSet) {
	if f.self.ArgsPtr != nil {
		*f.self.ArgsPtr = nil
	}
	resetPtrVal(f.self.Ptr)
	for i := range f.flags {
		resetPtrVal(f.flags[i].Ptr)
	}
	for i := range f.subsets {
		r.reset(&f.subsets[i])
	}
}
