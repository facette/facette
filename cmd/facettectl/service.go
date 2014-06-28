package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/connector"
)

func handleService(config *config.Config, args []string) error {
	cmd := &cmdServer{config: config}

	switch args[0] {
	case "reload":
		return cmd.reload(args[1:])
	case "support":
		return cmd.support(args[1:])
	}

	return nil
}

type cmdServer struct {
	config *config.Config
}

func (cmd *cmdServer) reload(args []string) error {
	if len(args) > 0 {
		return os.ErrInvalid
	}

	if cmd.config.PidFile == "" {
		return fmt.Errorf("missing pid configuration")
	} else if _, err := os.Stat(cmd.config.PidFile); os.IsNotExist(err) {
		return fmt.Errorf("missing pid file")
	}

	data, err := ioutil.ReadFile(cmd.config.PidFile)
	if err != nil {
		return err
	}

	pid, err := strconv.Atoi(strings.Trim(string(data), "\n"))
	if err != nil {
		return err
	}

	return syscall.Kill(pid, syscall.SIGHUP)
}

func (cmd *cmdServer) support(args []string) error {
	fmt.Println("Provider connectors:")

	keys := []string{}

	for key := range connector.Connectors {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for key := range connector.Connectors {
		fmt.Printf("   %s\n", key)
	}

	return nil
}
