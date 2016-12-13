package forge_test

import (
	"bytes"
	"fmt"

	"github.com/brettlangdon/forge"
)

func Example() {
	// Parse a `SectionValue` from `example.cfg`
	settings, err := forge.ParseFile("example.cfg")
	if err != nil {
		panic(err)
	}

	// Get a single value
	if settings.Exists("global") {
		// Get `global` casted as a string
		value, _ := settings.GetString("global")
		fmt.Printf("global = \"%s\"\r\n", value)
	}

	// Get a nested value
	value, err := settings.Resolve("primary.included_setting")
	fmt.Printf("primary.included_setting = \"%s\"\r\n", value.GetValue())

	// You can also traverse down the sections manually
	primary, err := settings.GetSection("primary")
	strVal, err := primary.GetString("included_setting")
	fmt.Printf("primary.included_setting = \"%s\"\r\n", strVal)

	// Convert settings to a map
	settingsMap := settings.ToMap()
	fmt.Printf("global = \"%s\"\r\n", settingsMap["global"])

	// Convert settings to JSON
	jsonBytes, err := settings.ToJSON()
	fmt.Printf("\r\n\r\n%s\r\n", string(jsonBytes))
}

func ExampleParseFile() {
	// Parse a `SectionValue` from `example.cfg`
	settings, err := forge.ParseFile("example.cfg")
	if err != nil {
		panic(err)
	}
	fmt.Println(settings)
}

func ExampleParseString() {
	// Parse a `SectionValue` from string containing the config
	data := "amount = 500;"
	settings, err := forge.ParseString(data)
	if err != nil {
		panic(err)
	}

	fmt.Println(settings.GetInteger("amount"))
}

func ExampleParseBytes() {
	// Parse a `SectionValue` from []byte containing the config
	data := []byte("amount = 500;")
	settings, err := forge.ParseBytes(data)
	if err != nil {
		panic(err)
	}

	fmt.Println(settings.GetInteger("amount"))
}

func ExampleParseReader() {
	// Parse a `SectionValue` from []byte containing the config
	data := []byte("amount = 500;")
	reader := bytes.NewBuffer(data)
	settings, err := forge.ParseReader(reader)
	if err != nil {
		panic(err)
	}

	fmt.Println(settings.GetInteger("amount"))
}
