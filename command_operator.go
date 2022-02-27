package main

import (
	"github.com/urfave/cli/v2"
)

var operatorCommandName = "operator"

func newOperatorCommand() *cli.Command {
	return &cli.Command{
		Name:  operatorCommandName,
		Usage: "Runs operator commands",
		Subcommands: []*cli.Command{
			newOperatorBackupCommand(),
		},
	}
}
