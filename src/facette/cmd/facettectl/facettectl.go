package main

import (
	"facette/common"
	"facette/utils"
	"flag"
	"fmt"
	"os"
)

const (
	cmdUsage = `Usage: %s [OPTIONS] useradd NAME
       %[1]s [OPTIONS] userdel NAME
       %[1]s [OPTIONS] userlist
       %[1]s [OPTIONS] usermod NAME`
)

var (
	flagConfig string
	flagDebug  int
	flagHelp   bool
)

func init() {
	flag.StringVar(&flagConfig, "c", common.DefaultConfigFile, "configuration file path")
	flag.IntVar(&flagDebug, "d", 0, "debugging level")
	flag.BoolVar(&flagHelp, "h", false, "display this help and exit")
	flag.Usage = func() { utils.PrintUsage(os.Stderr, cmdUsage) }
	flag.Parse()

	if flagHelp {
		utils.PrintUsage(os.Stdout, cmdUsage)
	} else if flagConfig == "" {
		fmt.Fprintf(os.Stderr, "Error: configuration file path is mandatory\n")
		utils.PrintUsage(os.Stderr, cmdUsage)
	}
}

func main() {
	var (
		config  *common.Config
		err     error
		handler func(*common.Config, []string) error
	)

	config = &common.Config{}
	if err = config.Load(flagConfig); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	if len(flag.Args()) == 0 {
		utils.PrintUsage(os.Stderr, cmdUsage)
	}

	switch flag.Args()[0] {
	case "userlist", "useradd", "userdel", "usermod":
		handler = handleUser
	}

	err = handler(config, flag.Args())

	switch err {
	case nil:
		break

	case os.ErrInvalid:
		utils.PrintUsage(os.Stderr, cmdUsage)
		break

	default:
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		break
	}
}
