package kopia

// setMaintenanceOwner sets the owner for the maintenance.
// In kopia the maintenance owner for each repository has to be set.
// As the amount of pods can change in k8s, we need to ensure that there's a maintenance owner for each run.
func (k *Kopia) setMaintenanceOwner(owner string) error {
	k.log.WithName("maintenance_set_owner").V(1).Info("setting owner", "ownerName", owner)

	return k.runKopiaCommand("maintenance_set_owner", []string{
		"maintenance",
		"set",
		"--owner",
		owner,
	})
}

func (k *Kopia) enableQuickMaintenance() error {
	k.log.V(1).WithName("maintenance_enable_quick").Info("enabling quick maintenance")

	return k.runKopiaCommand("maintenance_enable_quick", []string{
		"maintenance",
		"set",
		"--enable-quick",
		"true",
	})
}

// RunMaintenance will run the maintenance for this kopia instance.
// It will first set the maintenance owner, then enable quick maintenance and finally run it.
func (k *Kopia) RunMaintenance(owner string) error {
	err := k.setMaintenanceOwner(owner)
	if err != nil {
		return err
	}

	err = k.enableQuickMaintenance()
	if err != nil {
		return err
	}

	k.log.V(1).WithName("maintenance").Info("running maintenance")
	return k.runKopiaCommand("maintenance", []string{
		"maintenance",
		"run",
	})
}
