package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/facette/httputil"
)

type httpInfo struct {
	Version    string   `json:"version"`
	BuildDate  string   `json:"build_date"`
	BuildHash  string   `json:"build_hash"`
	Compiler   string   `json:"compiler"`
	Drivers    []string `json:"drivers"`
	Connectors []string `json:"connectors"`
}

func execVersion() {
	fmt.Println("Client:")
	fmt.Printf("   Version:     %s\n", version)
	fmt.Printf("   Build date:  %s\n", buildDate)
	fmt.Printf("   Build hash:  %s\n", buildHash)
	fmt.Printf("   Compiler:    %s (%s)\n", runtime.Version(), runtime.Compiler)
	fmt.Println("")

	// Create new HTTP request
	hc := httputil.NewClient(time.Duration(upstreamTimeout)*time.Second, true, false)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/", upstreamAddress), nil)
	if err != nil {
		die("%s", err)
	}

	req.Header.Add("User-Agent", "facettectl/"+version)

	// Retrieve items list from backend
	resp, err := hc.Do(req)
	if err != nil {
		die("%s", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		die("%s", err)
	}

	info := httpInfo{}
	if err := json.Unmarshal(data, &info); err != nil {
		die("%s", err)
	}

	fmt.Println("Server:")
	fmt.Printf("   Version:     %s\n", info.Version)
	fmt.Printf("   Build date:  %s\n", info.BuildDate)
	fmt.Printf("   Build hash:  %s\n", info.BuildHash)
	fmt.Printf("   Compiler:    %s\n", info.Compiler)
	fmt.Printf("   Drivers:     %s\n", strings.Join(info.Drivers, ", "))
	fmt.Printf("   Connectors:  %s\n", strings.Join(info.Connectors, ", "))
}
