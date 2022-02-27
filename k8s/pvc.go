package k8s

import (
	"fmt"

	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// MountedPVC describes a PVC and the POD it belongs to.
type MountedPVC struct {
	Pod *v1.Pod
	PVC *v1.PersistentVolumeClaim
}

// BackupPVCList contains two lists of PVCs.
// One list contains the PVCs that are currently mounted.
// The other contains the PVCs that aren't mounted.
type BackupPVCList struct {
	// MountedPVCs consists of the key <pvcname:namespace>.
	// This helps with deduplicating the PVCs afterwards.
	MountedPVCs   map[string]MountedPVC
	UnmountedPVCs *v1.PersistentVolumeClaimList
}

// ListEligiblePVCs will list all PVCs that fullfill certain constraints.
func ListEligiblePVCs(cliCtx *cli.Context, k8sClient client.Client) (*BackupPVCList, error) {
	log := logger.AppLogger(cliCtx.Context).WithName("PVCLister")

	pods, err := listPodsWithPVCs(cliCtx, k8sClient)
	if err != nil {
		return nil, err
	}

	backupList := &BackupPVCList{MountedPVCs: map[string]MountedPVC{}}

	for _, tmp := range pods.Items {
		pod := tmp
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				pvc, err := getPVCFromClaimSource(cliCtx, k8sClient, volume.PersistentVolumeClaim, &pod)
				backupListKey := fmt.Sprintf("%s:%s", pvc.Name, pvc.Namespace)
				if err != nil {
					return nil, fmt.Errorf("could not get pvc for pod %s: %w", pod.Name, err)
				}
				backupList.MountedPVCs[backupListKey] = MountedPVC{Pod: &pod, PVC: pvc}
				log.V(1).Info("found pod and pvc", "podname", pod.Name, "pvcname", pvc.Name, "namespace", pod.Namespace)
			}
		}
	}

	allPVCs := v1.PersistentVolumeClaimList{}
	err = k8sClient.List(cliCtx.Context, &allPVCs)

	backupList.UnmountedPVCs = &v1.PersistentVolumeClaimList{}
	for _, pvc := range allPVCs.Items {
		pvcKey := fmt.Sprintf("%s:%s", pvc.Name, pvc.Namespace)
		if _, ok := backupList.MountedPVCs[pvcKey]; !ok {
			backupList.UnmountedPVCs.Items = append(backupList.UnmountedPVCs.Items, pvc)
			log.V(1).Info("found unmounted PVC", "pvcname", pvc.Name, "namespace", pvc.Namespace)
		}
	}

	return backupList, err
}

func getPVCFromClaimSource(cliCtx *cli.Context, k8sClient client.Client, PVCSource *v1.PersistentVolumeClaimVolumeSource, pod *v1.Pod) (*v1.PersistentVolumeClaim, error) {
	objectKey := client.ObjectKey{
		Namespace: pod.Namespace,
		Name:      PVCSource.ClaimName,
	}
	pvc := &v1.PersistentVolumeClaim{}
	err := k8sClient.Get(cliCtx.Context, objectKey, pvc)
	return pvc, err
}
