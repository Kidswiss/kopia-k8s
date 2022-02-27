package main

import (
	"log"
	"strings"
	"sync/atomic"

	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setupLogging(c *cli.Context) error {
	log := newZapLogger(appName, c.Bool("debug"), usesProductionLoggingConfig(c))
	c.Context.Value(logger.ContextKey{}).(*atomic.Value).Store(log)
	return nil
}

func usesProductionLoggingConfig(c *cli.Context) bool {
	return strings.EqualFold("JSON", c.String("log-format"))
}

func newZapLogger(name string, debug bool, useProductionConfig bool) logr.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.ConsoleSeparator = " | "
	cfg.DisableStacktrace = true
	if useProductionConfig {
		cfg = zap.NewProductionConfig()
	}
	if debug {
		// Zap's levels get more verbose as the number gets smaller,
		// bug logr's level increases with greater numbers.
		cfg.Level = zap.NewAtomicLevelAt(zapcore.Level(-2)) // max logger.V(2)
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
	z, err := cfg.Build()
	zap.ReplaceGlobals(z)
	if err != nil {
		log.Fatalf("error configuring the logging stack")
	}
	log := zapr.NewLogger(z).WithName(name)
	if useProductionConfig {
		// Append the version to each log so that logging stacks like EFK/Loki can correlate errors with specific versions.
		return log.WithValues("version", version)
	}
	return log
}
