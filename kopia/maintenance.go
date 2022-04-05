package kopia

import "path"

// SetMaintenanceOwner sets the owner for the maintenance.
// In kopia the maintenance owner for each repository has to be set.
// As the amount of pods can change in k8s, we need to ensure that there's a maintenance owner for each run.
func (k *Kopia) SetMaintenanceOwner(owner string) error {
	log := k.log.WithName("maintenance")
	log.V(1).Info("setting owner", "ownerName", owner)

	maintenanceOwnerCommand := newCommand(k.ctx, log.WithName("kopia"), k.kopiaPath)
	maintenanceOwnerCommand.args = []string{
		"--config-file",
		path.Join(k.configPath, "kopia.json"),
		"maintenance",
		"set",
		"--owner",
		owner,
	}
	err := maintenanceOwnerCommand.run()
	if err != nil {
		log.Error(err, "error during kopia execution")
		k.LastExitCode = err
	}
	return err
}
