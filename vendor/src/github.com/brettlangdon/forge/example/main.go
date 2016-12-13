package main

import (
	"fmt"

	"github.com/brettlangdon/forge"
)

func main() {
	// Parse a `SectionValue` from `example.cfg`
	settings, err := forge.ParseFile("example.cfg")
	if err != nil {
		panic(err)
	}

	str_val, err := settings.GetString("global")
	if err != nil {
		panic(err)
	}
	fmt.Printf("global = \"%s\"\r\n", str_val)

	// Get a nested value
	// value, err := settings.Resolve("primary.included_setting")
	// fmt.Printf("primary.included_setting = \"%s\"\r\n", value.GetValue())

	// Convert settings to a map
	settingsMap := settings.ToMap()
	fmt.Printf("global = \"%s\"\r\n", settingsMap["global"])

	// Convert settings to JSON
	jsonBytes, err := settings.ToJSON()
	fmt.Printf("\r\n\r\n%s\r\n", string(jsonBytes))
}
