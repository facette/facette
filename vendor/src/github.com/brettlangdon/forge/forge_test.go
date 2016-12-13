package forge_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/brettlangdon/forge"
)

var testConfigBytes = []byte(`
# Global stuff
global = "global value";
# Primary stuff
primary {
  string = "primary string value";
  string_with_quote = "some \"quoted\" str\\ing";
  single = 'hello world';
  empty = '';
  single_with_quote = '\'hello\' "world"';

  # Semicolons are optional
  integer500 = 500
  float = 80.80
  negative = -50
  boolean = true
  not_true = FALSE
  nothing = NULL

  list = [
       TRUE,
       FALSE,
       50.5,
       "hello",
       'list',
  ]

  # Reference secondary._under (which hasn't been defined yet)
  sec_ref = secondary._under;
   # Primary-sub stuff
  sub {
      key = "primary sub key value";
      include "./test_include.cfg";
  }

  sub_section {
      # Testing of a special case that had previous caused failures
      # Was caused by an array with no ending semicolon, followed directly by another setting
      nested_array_no_semi_colon = ["a", "b"]
      another = true
  }
}

secondary {
  another = "secondary another value";
  global_reference = global;
  primary_sub_key = primary.sub.key;
  another_again = .another;  # References secondary.another
  _under = 50;
}
`)

var testConfigString = string(testConfigBytes)
var testConfigReader = bytes.NewReader(testConfigBytes)

func assertEqual(a interface{}, b interface{}, t *testing.T) {
	if a != b {
		t.Fatal(fmt.Sprintf("'%v' != '%v'", a, b))
	}
}

func assertDirectives(values map[string]interface{}, t *testing.T) {
	// Global
	assertEqual(values["global"], "global value", t)

	// Primary
	primary := values["primary"].(map[string]interface{})
	assertEqual(primary["string"], "primary string value", t)
	assertEqual(primary["string_with_quote"], "some \"quoted\" str\\ing", t)
	assertEqual(primary["single"], "hello world", t)
	assertEqual(primary["empty"], "", t)
	assertEqual(primary["single_with_quote"], "'hello' \"world\"", t)
	assertEqual(primary["integer500"], int64(500), t)
	assertEqual(primary["float"], float64(80.80), t)
	assertEqual(primary["negative"], int64(-50), t)
	assertEqual(primary["boolean"], true, t)
	assertEqual(primary["not_true"], false, t)
	assertEqual(primary["nothing"], nil, t)
	assertEqual(primary["sec_ref"], int64(50), t)

	// Primary list
	list := primary["list"].([]interface{})
	assertEqual(len(list), 5, t)
	assertEqual(list[0], true, t)
	assertEqual(list[1], false, t)
	assertEqual(list[2], float64(50.5), t)
	assertEqual(list[3], "hello", t)
	assertEqual(list[4], "list", t)

	// Primary Sub
	sub := primary["sub"].(map[string]interface{})
	assertEqual(sub["key"], "primary sub key value", t)
	assertEqual(sub["included_setting"], "primary sub included_setting value", t)

	// Secondary
	secondary := values["secondary"].(map[string]interface{})
	assertEqual(secondary["another"], "secondary another value", t)
	assertEqual(secondary["global_reference"], "global value", t)
	assertEqual(secondary["primary_sub_key"], "primary sub key value", t)
	assertEqual(secondary["another_again"], "secondary another value", t)
	assertEqual(secondary["_under"], int64(50), t)
}

func TestParseBytes(t *testing.T) {
	settings, err := forge.ParseBytes(testConfigBytes)
	if err != nil {
		t.Fatal(err)
	}

	values := settings.ToMap()
	assertDirectives(values, t)
}

func TestParseString(t *testing.T) {
	settings, err := forge.ParseString(testConfigString)
	if err != nil {
		t.Fatal(err)
	}
	values := settings.ToMap()
	assertDirectives(values, t)
}

func TestParseReader(t *testing.T) {
	settings, err := forge.ParseReader(testConfigReader)
	if err != nil {
		t.Fatal(err)
	}
	values := settings.ToMap()
	assertDirectives(values, t)
}

func TestParseFile(t *testing.T) {
	settings, err := forge.ParseFile("./test.cfg")
	if err != nil {
		t.Fatal(err)
	}
	values := settings.ToMap()
	assertDirectives(values, t)
}
