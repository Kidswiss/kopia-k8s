package main

import (
	"git.earthnet.ch/simon.beck/kopia-k8s/k8s"
	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

func newOperatorBackupCommand() *cli.Command {

	return &cli.Command{
		Name:   "backup",
		Usage:  "Schedules backup jobs on the cluster",
		Action: runOperatorBackup,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "pre-backup-annotation",
				Value:   "kopia.earthnet.ch/prebackup",
				Usage:   "The annotation that contains the pre-backup command",
				EnvVars: envVars("PRE_BACKUP_ANNOTATION"),
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Value:   3,
				Usage:   "How many backup pods should run at the same time",
				EnvVars: envVars("CONCURRENCY"),
			},
			&cli.StringFlag{
				Name:    "uuid",
				Value:   uuid.New().String(),
				Usage:   "Random ID for jobs, so that each kopia-k8s instance can operate on their own jobs",
				EnvVars: envVars("UUID"),
			},
		}, getRepositoryParams()...),
	}
}

func runOperatorBackup(c *cli.Context) error {
	logger := logger.AppLogger(c.Context).WithName("operator")
	logger.V(1).Info("starting operator")

	operator := newOperator(c)
	mgr := operator.initManager()
	operator.registerController(mgr)
	operator.startManager(mgr)

	pvcList, err := k8s.ListEligiblePVCs(c, mgr.GetClient())
	if err != nil {
		return err
	}

	jobRunner := k8s.JobRunner{
		CliCtx:      c,
		K8sClient:   mgr.GetClient(),
		Concurrency: c.Int("concurrency"),
		PvcList:     pvcList,
	}

	err = k8s.ExecutePrebackupCommand(c, mgr.GetClient())
	if err != nil {
		return err
	}

	return jobRunner.RunAndWatchBackupJobs()
}
