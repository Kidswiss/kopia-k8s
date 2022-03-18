package main

import (
	"git.earthnet.ch/simon.beck/kopia-k8s/kopia"
	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/urfave/cli/v2"
)

var kopiaCommandName = "kopia"

func newKopiaCommand() *cli.Command {
	return &cli.Command{
		Name:  kopiaCommandName,
		Usage: "Runs kopia commands",
		Subcommands: []*cli.Command{
			newKopiaBackupCommand(),
		},
		Flags: append([]cli.Flag{
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to the folder where the config should be written",
				EnvVars: envVars("CONFIG_PATH"),
				Value:   "/tmp",
			},
			&cli.StringFlag{
				Name:    "kopia-bin-path",
				Usage:   "Kopia binary path",
				EnvVars: envVars("KOPIA_PATH"),
				Value:   "/usr/local/bin/kopia",
			},
		}, getRepositoryParams()...),
	}
}

func getRepositoryParams() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "access-key-id",
			Usage:   "AWS access key ID",
			EnvVars: []string{"AWS_ACCESS_KEY_ID"},
		},
		&cli.StringFlag{
			Name:    "secret-access-key",
			Usage:   "AWS secret access key",
			EnvVars: []string{"AWS_SECRET_ACCESS_KEY"},
		},
		&cli.StringFlag{
			Name:    "encryption-password",
			Usage:   "Kopia encryption password",
			EnvVars: envVars("ENCRYPTION_PASSWORD"),
		},
		&cli.StringFlag{
			Name:    "bucket",
			Usage:   "Kopia S3 repository",
			EnvVars: envVars("BUCKET"),
		},
		&cli.StringFlag{
			Name:    "s3-endpoint",
			Usage:   "Kopia S3 endpoint",
			EnvVars: envVars("ENDPOINT"),
		},
	}
}

func newKopiaInstance(c *cli.Context) *kopia.Kopia {
	logger.AppLogger(c.Context).V(1).Info("flag values",
		"access-key-id", c.String("access-key-id"),
		"secret-access-key", c.String("secret-access-key"),
		"encryption-password", c.String("encryption-password"),
		"endpoint", c.String("s3-endpoint"),
		"bucket", c.String("bucket"))
	return kopia.New(c.Context, c.Path("config"),
		c.String("access-key-id"),
		c.String("secret-access-key"),
		c.String("encryption-password"),
		c.String("s3-endpoint"),
		c.String("bucket"),
		c.String("kopia-bin-path"),
		c.String("hostname"))
}
