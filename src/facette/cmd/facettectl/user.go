package main

import (
	"bytes"
	"facette/auth"
	"facette/common"
	"facette/utils"
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"time"
)

func handleUser(config *common.Config, args []string) error {
	var (
		cmd *cmdAuth
		err error
	)

	cmd = &cmdAuth{auth: auth.NewAuth(config, flagDebug)}

	if err = cmd.auth.Update(); err != nil {
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
	auth *auth.Auth
}

func (cmd *cmdAuth) list(args []string) error {
	if len(args) > 0 {
		return os.ErrInvalid
	}

	for name := range cmd.auth.Users {
		fmt.Println(name)
	}

	return nil
}

func (cmd *cmdAuth) save() error {
	return utils.JSONDump(cmd.auth.Config.AuthFile, &cmd.auth.Users, time.Now())
}

func (cmd *cmdAuth) set(args []string, create bool) error {
	var (
		exists   bool
		password []byte
	)

	if len(args) != 1 {
		return os.ErrInvalid
	}

	// Check for possible conflicts
	_, exists = cmd.auth.Users[args[0]]

	if create && exists {
		return fmt.Errorf("user `%s' already exists", args[0])
	} else if !create && !exists {
		return fmt.Errorf("user `%s' not found", args[0])
	}

	// Set user password
	fmt.Print("Password: ")
	password = gopass.GetPasswd()

	fmt.Print("Repeat Password: ")
	if !bytes.Equal(password, gopass.GetPasswd()) {
		return fmt.Errorf("passwords don't match")
	}

	cmd.auth.Users[args[0]] = cmd.auth.Hash(string(password))

	return cmd.save()
}

func (cmd *cmdAuth) unset(args []string) error {
	if len(args) != 1 {
		return os.ErrInvalid
	}

	delete(cmd.auth.Users, args[0])

	return cmd.save()
}
