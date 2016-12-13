package forge_test

import (
	"bytes"
	"testing"

	"github.com/brettlangdon/forge"
)

var exampleConfigBytes = []byte(`
# Global stuff
global = "global value";
# Primary stuff
primary {
  string = "primary string value";
  integer = 500;
  float = 80.80;
  negative = -50;
  boolean = true;
  not_true = FALSE;
  nothing = NULL;
   # Primary-sub stuff
  sub {
      key = "primary sub key value";
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

var exampleConfigString = string(exampleConfigBytes)
var exampleConfigReader = bytes.NewReader(exampleConfigBytes)

func BenchmarkParseBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := forge.ParseBytes(exampleConfigBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := forge.ParseString(exampleConfigString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		exampleConfigReader.Seek(0, 0)
		_, err := forge.ParseReader(exampleConfigReader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := forge.ParseFile("./test.cfg")
		if err != nil {
			b.Fatal(err)
		}
	}
}
