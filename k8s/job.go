package k8s

import (
	"fmt"
	"path"
	"strings"

	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/urfave/cli/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FinishedJobChannel informs RunAndWatchBackupJobs about what jobs
// have already finished.
var FinishedJobChannel chan bool = make(chan bool)

// JobRunner contains all necessary information to run the backup jobs.
type JobRunner struct {
	CliCtx      *cli.Context
	K8sClient   client.Client
	Concurrency int
	PvcList     *BackupPVCList
}

// JobLabel defines the label that should be used for the jobs
const JobLabel = "kopia.earthnet.ch"

// RunAndWatchBackupJobs will start all the jobs for the given PVC list.
// It will block until all the jobs have either finished or failed.
func (j *JobRunner) RunAndWatchBackupJobs() error {

	log := logger.AppLogger(j.CliCtx.Context).WithName("BackupAndWatch")

	mountedJobCount := 0

	for _, pvc := range j.PvcList.MountedPVCs {
		// TODO: this has some slight race condition.
		// It's possible that it spawns one job too many if two jobs finish
		// at exactly the same time.
		for mountedJobCount >= j.Concurrency {
			<-FinishedJobChannel
			mountedJobCount--
		}
		log.Info("starting backup", "pvcname", pvc.PVC.Name, "podname", pvc.Pod.Name)
		createServiceAccount(*j.CliCtx, j.K8sClient, pvc.Pod.Namespace)
		job := j.newBackupJob(pvc.PVC, pvc.Pod)
		err := j.K8sClient.Create(j.CliCtx.Context, job)
		if err != nil {
			return err
		}
		mountedJobCount++
	}

	for mountedJobCount > 0 {
		<-FinishedJobChannel
		mountedJobCount--
	}

	return nil
}

func (j *JobRunner) generateJobName(podname, pvcname string) string {
	seed := strings.Split(j.CliCtx.String("uuid"), "-")[0]
	name := fmt.Sprintf("kopia-%s-%s-%s", seed, podname, pvcname)
	if len(name) > 63 {
		name = name[:63]
	}
	// Names that end with "-" are invalid for k8s.
	// If that's the case we shorten it by one until that's not the case anymore.
	for strings.HasSuffix(name, "-") {
		name = name[:len(name)-1]
	}
	return name
}

func (j *JobRunner) getJobEnv() []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  "AWS_ACCESS_KEY_ID",
			Value: j.CliCtx.String("access-key-id"),
		},
		{
			Name:  "AWS_SECRET_ACCESS_KEY",
			Value: j.CliCtx.String("secret-access-key"),
		},
		{
			Name:  "KK_ENCRYPTION_PASSWORD",
			Value: j.CliCtx.String("encryption-password"),
		},
		{
			Name:  "KK_BUCKET",
			Value: j.CliCtx.String("bucket"),
		},
		{
			Name:  "KK_ENDPOINT",
			Value: j.CliCtx.String("s3-endpoint"),
		},
	}
}

func (j JobRunner) newBackupJob(pvc *v1.PersistentVolumeClaim, pod *v1.Pod) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.generateJobName(pod.Name, pvc.Name),
			Namespace: pvc.Namespace,
			Labels: map[string]string{
				JobLabel: j.CliCtx.String("uuid"),
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ServiceAccountName: "kopia-k8s",
					Affinity: &v1.Affinity{
						PodAffinity: &v1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: pod.Labels,
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "kopia-backup",
							Image: "192.168.6.10:5000/kopia-k8s:latest",
							Args: []string{
								"kopia",
								"backup",
								"--path",
								path.Join("/data", pvc.Name),
								"--hostname",
								pod.Namespace,
							},
							Env: j.getJobEnv(),
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "data",
									MountPath: path.Join("/data", pvc.Name),
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "data",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc.Name,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyOnFailure,
				},
			},
		},
	}
}
