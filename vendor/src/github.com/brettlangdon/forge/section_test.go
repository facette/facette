package forge_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/brettlangdon/forge"
)

func TestSectionKeys(t *testing.T) {
	t.Parallel()

	section := forge.NewSection()
	section.SetString("key1", "value1")
	section.SetString("key2", "value2")
	section.SetString("key3", "value3")

	keys := section.Keys()

	if len(keys) != 3 {
		t.Error("expected Section to have 3 keys")
	}

	if keys[0] != "key1" {
		t.Error("expected 'key1' to be in the list of keys")
	}
	if keys[1] != "key2" {
		t.Error("expected 'key2' to be in the list of keys")
	}
	if keys[2] != "key3" {
		t.Error("expected 'key3' to be in the list of keys")
	}
}

func TestMergeSection(t *testing.T) {
	config1Str := `
global = "global value";

prod {
	value = "string value";
	integer = 500
	float = 80.80
	boolean = true
	negative = FALSE
	nothing = NULL
}
	`

	config2Str := `
integer = 500
float = 80.80
boolean = true
negative = FALSE
nothing = NULL

new_section {
	integer = 500
	float = 80.80
	boolean = true
	negative = FALSE
	nothing = NULL
}

prod {
	value = "new value";
	secret = "shhh";
	nothing = "value"
	negative = false
	boolean = false
}
	`

	config1, err := forge.ParseString(config1Str)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	config2, err := forge.ParseString(config2Str)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = config1.Merge(config2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Validation
	valueFloat, _ := config1.GetFloat("float")
	if valueFloat != float64(80.80) {
		t.Errorf("Excepted '80.80' got %v", valueFloat)
	}

	valueNegative, _ := config1.Resolve("new_section.negative")
	if valueNegative.GetValue().(bool) {
		t.Errorf("Excepted 'false' got %v", valueNegative)
	}

	valueString, _ := config1.Resolve("prod.value")
	if valueString.GetValue().(string) == "string value" {
		t.Errorf("Excepted 'new value' got %v", valueString)
	}

	valueBoolean, _ := config1.Resolve("prod.boolean")
	if valueBoolean.GetValue().(bool) {
		t.Errorf("Excepted 'false' got %v", valueBoolean)
	}

	bytes, _ := json.MarshalIndent(config1.ToMap(), "", "   ")
	fmt.Println(string(bytes))
}

func TestMergeSectionFailSectionToField(t *testing.T) {
	config1Str := `
global = "global value";

prod {
	value = "string value";
	integer = 500
	float = 80.80
	boolean = true
	negative = FALSE
	nothing = NULL
}
	`

	config2Str := `
global = "global value";

prod = "I'm prod value"
	`

	config1, err := forge.ParseString(config1Str)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	config2, err := forge.ParseString(config2Str)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = config1.Merge(config2)
	if err.Error() != "source (STRING) and target (SECTION) type doesn't match: prod" {
		t.Error(err)
	}
}
