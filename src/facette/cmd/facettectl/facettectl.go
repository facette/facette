package main

import (
	"facette/utils"
	"flag"
	"fmt"
	"os"
)

var (
	flagConfig string
	flagDebug  int
	flagHelp   bool
)

func init() {
	flag.StringVar(&flagConfig, "c", "", "configuration file path")
	flag.IntVar(&flagDebug, "d", 0, "debugging level")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.Usage = func() { utils.PrintUsage(os.Stderr) }
	flag.Parse()

	if flagHelp {
		utils.PrintUsage(os.Stdout)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		utils.PrintUsage(os.Stderr)
	}
}

func main() {
}
