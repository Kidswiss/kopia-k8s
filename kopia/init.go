package kopia

import "path"

func (k *Kopia) initRepo() {
	log := k.log.WithName("initRepo")

	backupCommand := newCommand(k.ctx, log.WithName("kopia"), k.kopiaPath)
	backupCommand.args = []string{
		"repository",
		"create",
		"s3",
		"--bucket",
		k.bucket,
		"--access-key",
		k.accessKeyID,
		"--secret-access-key",
		k.secretAccessKey,
		"--endpoint",
		k.endpoint,
		"--config-file",
		path.Join(k.configPath, "kopia.json"),
		"--password",
		k.encryptionPassword,
	}
	err := backupCommand.run()
	if err != nil {
		log.Error(err, "error during repository creation")
	}
}
