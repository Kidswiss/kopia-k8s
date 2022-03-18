package kopia

import "path"

// Backup does a backup of the given Path
func (k *Kopia) Backup(backupPath string) {
	log := k.log.WithName("backup")
	log.V(1).Info("starting backup", "c", k.ctx)
	log.V(1).Info("repository config", "path", k.configPath)

	backupCommand := newCommand(k.ctx, log.WithName("kopia"), k.kopiaPath)
	backupCommand.args = []string{
		"--config-file",
		path.Join(k.configPath, "kopia.json"),
		"snapshot",
		"--json",
		backupPath,
		"--password",
		k.encryptionPassword,
	}
	err := backupCommand.run()
	if err != nil {
		log.Error(err, "error during kopia execution")
		k.LastExitCode = err
	}
}
