package main

import (
	"fmt"
	"os"
	"strings"

	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/urfave/cli/v2"
)

func newKopiaMaintenanceCommand() *cli.Command {

	return &cli.Command{
		Name:   "maintenance",
		Usage:  "Runs the maintenance on the given repository",
		Action: runMaintenance,
		Flags:  getRepositoryParams(),
	}
}

func runMaintenance(c *cli.Context) error {
	logger := logger.AppLogger(c.Context).WithName("maintenance")
	logger.Info("starting maintenance")

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	owner := strings.ToLower(fmt.Sprintf("%s@%s", "kopia-k8s", hostname))

	kopia := newKopiaInstance(c)
	return kopia.RunMaintenance(owner)
}
