package main

import (
	"facette/common"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func handleServer(config *common.Config, args []string) error {
	var (
		cmd *cmdServer
	)

	cmd = &cmdServer{config: config}

	switch args[0] {
	case "reload":
		return cmd.reload(args[1:])
	}

	return nil
}

type cmdServer struct {
	config *common.Config
}

func (cmd *cmdServer) reload(args []string) error {
	var (
		data []byte
		err  error
		pid  int
	)

	if len(args) > 0 {
		return os.ErrInvalid
	}

	if cmd.config.PidFile == "" {
		return fmt.Errorf("missing pid configuration")
	} else if _, err = os.Stat(cmd.config.PidFile); os.IsNotExist(err) {
		return fmt.Errorf("missing pid file")
	}

	if data, err = ioutil.ReadFile(cmd.config.PidFile); err != nil {
		return err
	} else if pid, err = strconv.Atoi(strings.Trim(string(data), "\n")); err != nil {
		return err
	}

	return syscall.Kill(pid, syscall.SIGHUP)
}
