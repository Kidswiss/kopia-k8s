package kopia

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type repository struct {
	Storage                 storage `json:"storage"`
	Caching                 caching `json:"caching"`
	Hostname                string  `json:"hostname"`
	Username                string  `json:"username"`
	Description             string  `json:"description"`
	EnableActions           bool    `json:"enableActions"`
	FormatBlobCacheDuration int64   `json:"formatBlobCacheDuration"`
}
type config struct {
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
}
type storage struct {
	Type   string `json:"type"`
	Config config `json:"config"`
}
type caching struct {
	CacheDirectory       string `json:"cacheDirectory"`
	MaxCacheSize         int64  `json:"maxCacheSize"`
	MaxMetadataCacheSize int64  `json:"maxMetadataCacheSize"`
	MaxListCacheDuration int    `json:"maxListCacheDuration"`
}

func (k *Kopia) newRepositoryConfigFile() repository {
	return repository{
		Storage: storage{
			Type: "s3",
			Config: config{
				Bucket:          k.bucket,
				Endpoint:        k.endpoint,
				AccessKeyID:     k.accessKeyID,
				SecretAccessKey: k.secretAccessKey,
			},
		},
		Caching: caching{
			CacheDirectory:       k.cachePath,
			MaxCacheSize:         5242880000,
			MaxMetadataCacheSize: 5242880000,
			MaxListCacheDuration: 30,
		},
		Hostname:                k.hostname,
		Username:                "kopia-k8s",
		Description:             "Repository in S3: restic.earthnet.ch kopia",
		EnableActions:           false,
		FormatBlobCacheDuration: 900000000000,
	}
}

func (k *Kopia) writeConfigFile() {
	repo := k.newRepositoryConfigFile()

	repoBytes, err := json.Marshal(repo)
	if err != nil {
		k.log.Error(err, "could not write kopia config")
	}

	err = os.WriteFile(filepath.Join(k.configPath, "kopia.json"), repoBytes, os.FileMode(0700))
	if err != nil {
		k.log.Error(err, "could not write kopia config")
	}
}
