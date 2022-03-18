package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func newKopiaBackupCommand() *cli.Command {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "kopia-k8s"
	}

	return &cli.Command{
		Name:   "backup",
		Usage:  "Does a backup",
		Action: runBackup,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "path",
				Aliases:  []string{"p"},
				Usage:    "Path which should get backed up, required",
				EnvVars:  envVars("BACKUP_PATH"),
				Required: true,
			},
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path, where the repository json is stored",
				EnvVars: envVars("CONFIG_PATH"),
				Value:   "/config",
			},
			&cli.StringFlag{
				Name:    "hostname",
				Usage:   "Set the hostname for the backup",
				EnvVars: []string{"HOSTNAME"},
				Value:   hostname,
			},
		},
	}
}

func runBackup(c *cli.Context) error {
	kopia := newKopiaInstance(c)
	kopia.Backup(c.Path("path"))
	return kopia.LastExitCode
}
