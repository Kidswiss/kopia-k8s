# Kopia-k8s

This is a POC how a new, very simple backup operator for K8s could look like. Its main purposes is to make backups for my homelab and to serve as a POC for a K8up v3.

## How does it work?
Kopia-k8s has two operation modes:
* Operator: This mode is intended to run within or against a k8s cluster. It doesn't trigger Kopia directly, but rather spawn jobs on K8s that run Kopia. It also runs the defined pre-backup scripts beforehand.
* Kopia: This mode runs actual Kopia commands. They are intended to run within a container.

When running `kopia-k8s operator backup` on either a workstation with a kubeconfig, or within a cluster as a job, it will first list all the running pods on the cluster. Then it filters out those that have a pre-backup annotation.

The command found in the pre-backup command annotation will be run sequentially and block until finished. They are executed by using exec on the pods. If one of the pre-backup commands fail with an exit code != 0 then the whole backup will be aborted and also exit with a non-zero code.

Once the pre-backup commands have finished, it will start to spawn the actual backup jobs within k8s. By default, it spawns 3 parallel jobs. Each of those jobs has pod-affinity rules, so that it's scheduled on the same host as the running pod. This ensures that the backup jobs can read the data from the same `RWO` PVCs. If Kopia encounters a critical error, it will exit with a non-zero exit code, thus failing the entire job. So Kopia-k8s jobs can be monitored by simply monitoring for failed jobs on the cluster.

## Todos
Some todos:
* Currently two parallel `kopia-k8s operator backup` instances could clash, because the spawned jobs don't have randomized names
* Maybe add some metrics, for example if there were some non-critical errors (e.g. a file could not be read)
