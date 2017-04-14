package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

func execVersion() error {
	fmt.Println("Client:")
	fmt.Printf("   Version:     %s\n", version)
	fmt.Printf("   Build date:  %s\n", buildDate)
	fmt.Printf("   Build hash:  %s\n", buildHash)
	fmt.Printf("   Compiler:    %s (%s)\n", runtime.Version(), runtime.Compiler)
	fmt.Println("")
	fmt.Println("Server:")

	// Retrieve version information from back-end
	info := apiInfo{}
	if err := apiRequest("GET", "/", nil, nil, &info); err != nil {
		return errors.Wrap(err, "failed to retrieve version information")
	}

	fmt.Printf("   Version:     %s\n", info.Version)
	fmt.Printf("   Build date:  %s\n", info.BuildDate)
	fmt.Printf("   Build hash:  %s\n", info.BuildHash)
	fmt.Printf("   Compiler:    %s\n", info.Compiler)
	fmt.Printf("   Drivers:     %s\n", strings.Join(info.Drivers, ", "))
	fmt.Printf("   Connectors:  %s\n", strings.Join(info.Connectors, ", "))

	return nil
}
