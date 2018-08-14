package flag

import "strings"

const (
	argumentFlagSplittable = iota + 1
	argumentFlag
	argumentValue
	argumentStopConsumption
	argumentReserve

	dash            = "-"
	doubleDash      = "--"
	stopConsumption = "-!"
	equal           = "="
)

type argument struct {
	Type  int
	Value string
}

type scanArgs struct {
	Flags       []argument
	FirstSubset string
	Sets        map[string]*scanArgs
}

type scanner struct {
	SubsetStack []string
	Result      scanArgs
}

func (s *scanner) appendArg(arg argument, isSubset bool) {
	curr := &s.Result
	for _, subset := range s.SubsetStack {
		set := curr.Sets[subset]
		if set == nil {
			set = &scanArgs{}
			if curr.Sets == nil {
				curr.Sets = make(map[string]*scanArgs)
			}
			curr.Sets[subset] = set
		}
		if len(curr.Sets) == 1 {
			curr.FirstSubset = subset
		}
		curr = set
	}
	if !isSubset || len(curr.Flags) == 0 {
		curr.Flags = append(curr.Flags, arg)
	}
}

func (s *scanner) isAlphabet(r rune) bool {
	return ('a' <= r && r <= 'z') || 'A' <= r && r <= 'Z'
}

func (s *scanner) checkSplits(f *FlagSet, rs []rune) (allFlag, firstFlag bool) {
	allFlag = true
	for i, r := range rs {
		isFlag := f.isFlagOrSubset(dash + string(r))
		if isFlag {
			if i == 0 && s.isAlphabet(r) && !s.isAlphabet(rs[i+1]) {
				firstFlag = true
			}
		} else {
			allFlag = false
			break
		}
	}
	return
}

func (s *scanner) stackTopFlagSet(f *FlagSet, stack []string) *FlagSet {
	curr := f
	for _, subset := range stack {
		curr = &curr.subsets[curr.subsetIndexes[subset]]
	}
	return curr
}

func (s *scanner) reverseIterStack(f *FlagSet, fn func(*FlagSet, int) (result, continu bool)) (result bool) {
	for i := len(s.SubsetStack); ; {
		currSet := s.stackTopFlagSet(f, s.SubsetStack[:i])
		result, continu := fn(currSet, i)
		if !continu {
			return result
		}

		i--
		if i < 0 {
			result, _ := fn(nil, i)
			return result
		}
	}
}

func (s *scanner) tryAppendFlagOrSubset(f *FlagSet, arg argument, mustAppend bool) bool {
	return s.reverseIterStack(f, func(currSet *FlagSet, i int) (result, continu bool) {
		if currSet == nil {
			if mustAppend {
				s.appendArg(arg, false)
			}
			return mustAppend, false
		}

		isFlag, isSubset := currSet.isFlag(arg.Value), currSet.isSubset(arg.Value)
		if !isFlag && !isSubset {
			return false, true
		}

		s.SubsetStack = s.SubsetStack[:i]
		if isSubset {
			s.SubsetStack = append(s.SubsetStack, arg.Value)
		}
		arg.Type = argumentFlag
		s.appendArg(arg, isSubset)
		return true, false
	})
}

func (s *scanner) appendSplittable(f *FlagSet, arg argument) {
	flagRunes := []rune(arg.Value[1:])
	s.reverseIterStack(f, func(currSet *FlagSet, i int) (result, continu bool) {
		if currSet == nil {
			arg.Type = argumentFlag
			s.appendArg(arg, false)
			return false, false
		}
		allFlag, firstFlag := s.checkSplits(currSet, flagRunes)
		if allFlag || firstFlag {
			s.SubsetStack = s.SubsetStack[:i]
		}
		if allFlag {
			for _, r := range flagRunes {
				s.appendArg(argument{Type: argumentFlag, Value: dash + string(r)}, false)
			}
			return false, false
		}
		if firstFlag {
			s.appendArg(argument{Type: argumentFlag, Value: dash + string(flagRunes[0])}, false)
			s.appendArg(argument{Type: argumentValue, Value: string(flagRunes[1:])}, false)
			return false, false
		}
		return false, true
	})
}

func (s *scanner) append(f *FlagSet, arg argument) {
	switch arg.Type {
	case argumentValue, argumentStopConsumption:
		s.appendArg(arg, false)
	case argumentFlag, argumentReserve:
		s.tryAppendFlagOrSubset(f, arg, true)
	case argumentFlagSplittable:
		s.appendSplittable(f, arg)
	}
}

func (s *scanner) canBeSplitBy(arg, sep string) bool {
	index := strings.Index(arg, sep)
	return index > 0 && index <= len(arg)-1
}

func (s *scanner) scanArg(f *FlagSet, isFirst bool, curr, next string) (consumed int) {
	const (
		mustFlag    = dash
		disableFlag = doubleDash
	)

	consumed = 1
	switch {
	case isFirst:
		s.append(f, argument{Type: argumentFlag, Value: curr})
	case curr == disableFlag:
		if next != "" {
			curr = next
			consumed = 2
		}
		s.append(f, argument{Type: argumentValue, Value: curr})
	case curr == mustFlag:
		var typ int
		if next != "" {
			curr = next
			consumed = 2
			typ = argumentFlag
		} else {
			typ = argumentValue
		}
		s.append(f, argument{Type: typ, Value: curr})
	case curr == stopConsumption:
		s.append(f, argument{Type: argumentStopConsumption, Value: curr})
	case s.canBeSplitBy(curr, equal):
		secs := strings.SplitN(curr, equal, 2)
		for i, sec := range secs {
			typ := argumentFlag
			if i != 0 {
				typ = argumentValue
			}
			s.append(f, argument{Type: typ, Value: sec})
		}
	case strings.HasPrefix(curr, doubleDash):
		s.append(f, argument{Type: argumentFlag, Value: curr})
	case s.tryAppendFlagOrSubset(f, argument{Type: argumentReserve, Value: curr}, false):
	case strings.HasPrefix(curr, dash):
		s.append(f, argument{Type: argumentFlagSplittable, Value: curr})
	default:
		s.append(f, argument{Type: argumentValue, Value: curr})
	}
	return
}

func (s *scanner) scan(f *FlagSet, args []string) {
	var (
		consumed int
		next     string
	)
	for i, l := 0, len(args); i < l; {
		arg := args[i]
		if i < l-1 {
			next = args[i+1]
		} else {
			next = ""
		}

		consumed = s.scanArg(f, i == 0, arg, next)
		i += consumed
	}
}
