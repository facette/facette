package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/facette/facette/pkg/auth"
	"github.com/facette/facette/pkg/config"
	"github.com/facette/facette/pkg/utils"
	"github.com/facette/facette/thirdparty/github.com/howeyc/gopass"
)

func handleUser(config *config.Config, args []string) error {
	cmd := &cmdAuth{handler: auth.AuthSimpleHandler{Config: config.Auth}}

	if err := cmd.handler.Update(); err != nil {
		return err
	}

	switch args[0] {
	case "useradd", "usermod":
		return cmd.set(args[1:], args[0] == "useradd")
	case "userdel":
		return cmd.unset(args[1:])
	case "userlist":
		return cmd.list(args[1:])
	}

	return nil
}

type cmdAuth struct {
	handler auth.AuthSimpleHandler
}

func (cmd *cmdAuth) list(args []string) error {
	if len(args) > 0 {
		return os.ErrInvalid
	}

	for name := range cmd.handler.Users {
		fmt.Println(name)
	}

	return nil
}

func (cmd *cmdAuth) save() error {
	return utils.JSONDump(cmd.handler.Config["path"], &cmd.handler.Users, time.Now())
}

func (cmd *cmdAuth) set(args []string, create bool) error {
	if len(args) != 1 {
		return os.ErrInvalid
	}

	exists := false

	// Check for possible conflicts
	_, exists = cmd.handler.Users[args[0]]

	if create && exists {
		return fmt.Errorf("user `%s' already exists", args[0])
	} else if !create && !exists {
		return fmt.Errorf("user `%s' not found", args[0])
	}

	// Set user password
	fmt.Print("Password: ")
	password := gopass.GetPasswd()

	fmt.Print("Repeat Password: ")
	if !bytes.Equal(password, gopass.GetPasswd()) {
		return fmt.Errorf("passwords don't match")
	}

	cmd.handler.Users[args[0]] = cmd.handler.Hash(string(password))

	return cmd.save()
}

func (cmd *cmdAuth) unset(args []string) error {
	if len(args) != 1 {
		return os.ErrInvalid
	}

	delete(cmd.handler.Users, args[0])

	return cmd.save()
}
