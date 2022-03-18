package kopia

import (
	"context"
	"os"

	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/go-logr/logr"
)

// Kopia holds all information necessary to run Kopia
type Kopia struct {
	ctx                context.Context
	log                logr.Logger
	configPath         string
	endpoint           string
	bucket             string
	accessKeyID        string
	secretAccessKey    string
	encryptionPassword string
	kopiaPath          string
	LastExitCode       error
	hostname           string
}

// New returns a new reference of kopia
func New(ctx context.Context, configPath, accessKeyID, secretAccessKey, encryptionPassword, endpoint, bucket, kopiaPath, hostname string) *Kopia {
	k := &Kopia{
		log:                logger.AppLogger(ctx),
		ctx:                ctx,
		configPath:         configPath,
		accessKeyID:        accessKeyID,
		secretAccessKey:    secretAccessKey,
		encryptionPassword: encryptionPassword,
		endpoint:           endpoint,
		bucket:             bucket,
		kopiaPath:          kopiaPath,
		hostname:           hostname,
	}
	os.Mkdir(configPath, os.FileMode(0755))
	k.initRepo()
	k.writeConfigFile()
	return k
}
