package kopia

// Backup does a backup of the given Path
func (k *Kopia) Backup(backupPath string) error {
	k.log.WithName("backup").V(1).Info("starting backup", "c", k.ctx)
	k.log.WithName("backup").V(1).Info("repository config", "path", k.configPath)

	return k.runKopiaCommand("backup", []string{
		"snapshot",
		"create",
		"--json",
		backupPath,
	})
}
