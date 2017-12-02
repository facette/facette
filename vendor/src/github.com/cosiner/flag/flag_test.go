package flag

import (
	"os"
	"reflect"
	"testing"

	"fmt"

	"github.com/cosiner/argv"
)

var tarCase = TestCases{
	Category: "Tar",
	Dst:      new(Tar),
	Cases: []TypeCase{
		{
			Cmds: []string{
				"tar -zcf a.tgz a b c",
				"tar -zc -f=a.tgz a b c",
				"tar -z -c -f a.tgz a b c",
				"tar --gz -c -f a.tgz a b c",
			},
			Value: &Tar{
				SourceFiles: []string{"a", "b", "c"},
				GZ:          true,
				Create:      true,
				File:        "a.tgz",
			},
		},
		{
			Cmds: []string{
				"tar -- -file -",
			},
			Value: &Tar{
				SourceFiles: []string{"-file", "-"},
			},
		},
		{
			Cmds: []string{
				"tar -- -file --",
			},
			Value: &Tar{
				SourceFiles: []string{"-file", "--"},
			},
		},
		{
			Cmds: []string{
				"tar - -z",
			},
			Value: &Tar{
				GZ: true,
			},
		},
		{
			Cmds: []string{
				"tar -Jxf a.txz -C /",
				"tar -Jxf a.txz -C/",
				"tar -Jxf a.txz -C=/",
			},
			Value: &Tar{
				XZ:        true,
				Extract:   true,
				File:      "a.txz",
				Directory: "/",
			},
		},
		{
			Cmds: []string{
				"tar -A",
				"tar --A",
				"tar -f",
				"tar -z aaa",
				"tar -z true bbb -f a.tgz",
				"tar -z true -z true",
			},
			Errors: []errorType{
				errFlagNotFound,
				errFlagNotFound,
				errStandaloneFlag,
				errInvalidValue,
				errStandaloneValue,
				errDuplicateFlagParsed,
			},
		},
	},
}

type GoBuild struct {
	Enable bool

	Object string   `names:"-o"`
	Files  []string `args:"1"`
}

type GoTest struct {
	Enable  bool
	CPUNum  int    `names:"-cpu"`
	Object  string `names:"-o"`
	Package string `names:"-p"`
}

type GoToolCover struct {
	Enable bool
	HTML   string
}

type GoToolPprof struct {
	Enable bool
	Web    bool
}

type GoTool struct {
	Enable bool
	Cover  GoToolCover
}

type Go struct {
	Build GoBuild
	Test  GoTest
	Tool  GoTool
}

func (*Go) Metadata() map[string]Flag {
	return map[string]Flag{
		"build": {
			Usage:   "build binary or library",
			Arglist: "[-o BINARY] [PATH]...",
		},
	}
}

var goCases = TestCases{
	Category: "Go",
	Dst:      new(Go),
	Cases: []TypeCase{
		{
			Cmds: []string{
				"go build -o a.out a.go b.go",
			},
			Value: &Go{
				Build: GoBuild{
					Enable: true,
					Object: "a.out",
					Files:  []string{"a.go", "b.go"},
				},
			},
		},
		{
			Cmds: []string{
				"go test -o a.test -cpu 2 -p math/rand",
			},
			Value: &Go{
				Test: GoTest{
					Enable:  true,
					CPUNum:  2,
					Object:  "a.test",
					Package: "math/rand",
				},
			},
		},
		{
			Cmds: []string{
				"go tool cover -html=cover.out",
			},
			Value: &Go{
				Tool: GoTool{
					Enable: true,
					Cover: GoToolCover{
						Enable: true,
						HTML:   "cover.out",
					},
				},
			},
		},
		{
			Cmds: []string{
				"go test bld",
			},
			Errors: []errorType{
				errStandaloneValue,
			},
		},
	},
}

type DefaultSelectEnv struct {
	Def      string   `names:"-d" default:"def" selects:"def,eef,fef"`
	Path     []string `names:"-p" env:"PATH" valsep:":"`
	Nums     []uint64 `names:"-n" default:"1,2,3" selects:"1,2,3,4,5"`
	Bool     bool     `default:"false"`
	Bools    []bool   `default:"true,false"`
	String   string   `default:""`
	Strings  []string `default:""`
	Int8     int8     `default:"0"`
	Int8s    []int8   `default:""`
	Int16    int16    `default:"0"`
	Int16s   []int16  `default:""`
	Int32    int32    `default:"0"`
	Int32s   []int32  `default:""`
	Int64    int64    `default:"0"`
	Int64s   []int64  `default:""`
	Int      int      `default:"0"`
	Ints     []int    `default:""`
	Uint8    uint8    `default:"0"`
	Uint8s   []uint8  `default:""`
	Uint16   uint16   `default:"0"`
	Uint16s  []uint16 `default:""`
	Uint32   uint32   `default:"0"`
	Uint32s  []uint32 `default:""`
	Uint64   uint64   `default:"0"`
	Uint64s  []uint64 `default:""`
	Uint     uint     `default:"0"`
	Uints    []uint   `default:""`
	Float32  float32
	Float32s []float32
	Float64  float64
	Float64s []float64
}

var defSelCase = TestCases{
	Category: "DefaultSelectEnv",
	Dst:      new(DefaultSelectEnv),
	Cases: []TypeCase{
		{
			Env: map[string]string{
				"PATH": "/bin:/usr/bin",
			},
			Cmds: []string{
				"def",
			},
			Value: &DefaultSelectEnv{
				Def:   "def",
				Path:  []string{"/bin", "/usr/bin"},
				Nums:  []uint64{1, 2, 3},
				Bools: []bool{true, false},
			},
		},
	},
}

type errorFlags1 struct {
	Selects []int `selects:"a,b,c"`
}

type errorFlags2 struct {
	Input string
	Build struct {
		Enable bool
		Input  struct {
			Enable bool
		} `names:"-input"`
	}
}

type TypeCase struct {
	Env    map[string]string
	Cmds   []string
	Errors []errorType
	Value  interface{}
}

type TestCases struct {
	Category string
	Dst      interface{}
	Error    errorType
	Cases    []TypeCase
}

func TestFlags(t *testing.T) {
	var typeCases = []TestCases{
		tarCase,
		goCases,
		defSelCase,
		{
			Dst:   DefaultSelectEnv{},
			Error: errNonPointer,
		},
		{
			Dst:   new(errorFlags1),
			Error: errInvalidSelects,
		},
		{
			Dst:   new(errorFlags2),
			Error: errDuplicateFlagRegister,
		},
	}
	var (
		gotErrorType = func(err error) errorType {
			if err != nil {
				return err.(flagError).Type
			}
			return 0
		}
		expectErrorType = func(errors []errorType, index int) errorType {
			if index < len(errors) {
				return errors[index]
			}
			return 0
		}
	)

	for i, typ := range typeCases {
		flags := NewFlagSet(Flag{}).ErrHandling(0)
		if i == 0 {
			flags.NeedHelpFlag(false)
			flags.UpdateMeta("", Flag{})
		}
		err := flags.StructFlags(typ.Dst)
		gotErr, expectErr := gotErrorType(err), typ.Error
		if gotErr.String() != expectErr.String() {
			t.Errorf("%s: parse flags failed: %d, expect error: %s, got error %s", typ.Category, i, expectErr.String(), gotErr.String())
			continue
		}

		for i, c := range typ.Cases {
			if c.Env != nil {
				envParser = func(name string) string {
					if val, has := c.Env[name]; has {
						return val
					}
					return os.Getenv(name)
				}
			} else {
				envParser = os.Getenv
			}
			for j, cmd := range c.Cmds {
				argv, err := argv.Argv([]rune(cmd), nil, nil)
				if err != nil {
					t.Fatal(i, j, err)
				}

				err = flags.Parse(argv[0]...)
				gotErr, expectErr := gotErrorType(err), expectErrorType(c.Errors, j)
				if gotErr.String() != expectErr.String() {
					t.Errorf("%s: parse failed: Case:%d Cmd:%d, expect error: %s, got error %s",
						typ.Category,
						i+1,
						j+1,
						expectErr.String(),
						gotErr.String(),
					)
				} else if gotErr == 0 && !reflect.DeepEqual(typ.Dst, c.Value) {
					t.Errorf("%s: match failed: Case:%d Cmd:%d, expect: %+v, got %+v",
						typ.Category,
						i+1,
						j+1,
						c.Value,
						typ.Dst,
					)
				}

				flags.ToString(true)
				flags.Reset()
			}
		}
	}
}

func TestHelp(t *testing.T) {
	var tar Tar

	set := NewFlagSet(Flag{})
	set.StructFlags(&tar)
	set.Help(false)
}

func TestSubset(t *testing.T) {
	var g GoCmd

	set := NewFlagSet(Flag{})
	set.StructFlags(&g)
	set.Help(false)

	fmt.Println()
	fmt.Println()
	fmt.Println()

	build, _ := set.FindSubset("build")
	build.Help(false)
}

func TestStopConsumption(t *testing.T) {
	type Flags struct {
		Slice []string
		Rest  []string `args:"true"`
	}

	var flags Flags
	NewFlagSet(Flag{}).ParseStruct(&flags, "test", "-slice", "a", "b", "--", "-!", "-!", "c", "d")
	if !reflect.DeepEqual(flags.Slice, []string{"a", "b", "-!"}) {
		t.FailNow()
	}
	if !reflect.DeepEqual(flags.Rest, []string{"c", "d"}) {
		t.FailNow()
	}
}
