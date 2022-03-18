package kopia

import (
	"context"
	"os"
	"os/exec"

	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/go-logr/logr"
)

type command struct {
	args      []string
	kopiaPath string
	ctx       context.Context
	log       logr.Logger
}

func newCommand(ctx context.Context, log logr.Logger, kopiaPath string) command {
	return command{
		kopiaPath: kopiaPath,
		ctx:       ctx,
		log:       log,
	}
}

func (k *command) run() error {
	cmd := exec.CommandContext(k.ctx, k.kopiaPath, k.args...)
	cmd.Env = os.Environ()

	stdoutHandler := kopiaStdoutParser{log: k.log.WithName("stdout")}

	cmd.Stdout = logger.New(stdoutHandler.parseKopiaStdout)

	cmd.Stderr = logger.New(stdoutHandler.parseKopiaStdout)

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
