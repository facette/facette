// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

// Package store provides a store sub-command.
package store

import "github.com/urfave/cli/v2"

// Command is a store command.
var Command = &cli.Command{
	Name:   "store",
	Usage:  "Manage back-end storage",
	Action: cli.ShowSubcommandHelp,
	Subcommands: []*cli.Command{
		{
			Name:   "dump",
			Usage:  "Dump back-end storage data",
			Action: dump,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "output file path",
				},
			},
		},
		{
			Name:   "restore",
			Usage:  "Restore back-end storage data",
			Action: restore,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "input",
					Aliases: []string{"i"},
					Usage:   "input file path",
				},
			},
		},
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "url",
			Usage:   "service URL",
			EnvVars: []string{"FACETTE_URL"},
			Value:   "http://localhost:12003",
		},
	},
	HideHelpCommand: true,
}

func dump(ctx *cli.Context) error {
	// TODO: implement
	return nil
}

func restore(ctx *cli.Context) error {
	// TODO: implement
	return nil
}
