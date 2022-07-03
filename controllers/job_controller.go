package controllers

import (
	"context"
	"time"

	"git.earthnet.ch/simon.beck/kopia-k8s/k8s"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	uuid   string
}

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status;jobs/finalizers,verbs=get;update;patch

// Reconcile is the entrypoint to manage the given resource.
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	myJob := &batchv1.Job{}

	err := r.Client.Get(ctx, req.NamespacedName, myJob)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	r.Log.V(1).Info("watched job", "name", req.Name, "status", myJob.Status)

	if myJob.ObjectMeta.Labels[k8s.JobLabel] == r.uuid && myJob.DeletionTimestamp == nil {
		if myJob.Status.Succeeded > 0 {
			backgroundDelete := v1.DeletePropagationBackground
			err := r.Client.Delete(ctx, myJob, &client.DeleteOptions{PropagationPolicy: &backgroundDelete})
			if err != nil {
				r.Log.Error(err, "job finished successfull, but cannot be cleaned up")
			} else {
				r.Log.Info("job finished successfully, cleaning up", "name", myJob.Name)
			}
			k8s.FinishedJobChannel <- true
		}
		if myJob.Status.Failed > 0 {
			r.Log.Error(nil, "job failed, not cleaning up", "name", myJob.Name)
			k8s.FinishedJobChannel <- true
		}
		if myJob.Status.Active > 0 {
			if time.Now().Sub(myJob.CreationTimestamp.Time).Minutes() > 15 {
				if r.isJobPodPending(ctx, myJob) {
					r.Log.Info("pod has been pending for over 5 minutes, skipping and starting next pod", "name", myJob.Name, "namespace", myJob.Namespace)
					k8s.FinishedJobChannel <- true
				}
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
}

func (r *JobReconciler) isJobPodPending(ctx context.Context, myJob *batchv1.Job) bool {
	podList := &corev1.PodList{}

	labelSelector, _ := createLabelSelector(myJob.Name)

	err := r.Client.List(ctx, podList, &client.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		r.Log.Error(err, "could not list pod to determine pending state", "name", myJob.Name, "namespace", myJob.Namespace)
		return false
	}

	if len(podList.Items) == 1 {
		return podList.Items[0].Status.Phase == corev1.PodPending
	}

	return false
}

func createLabelSelector(jobName string) (labels.Selector, error) {
	podReq, err := labels.NewRequirement("job-name", selection.In, []string{jobName})

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*podReq)
	return selector, err
}

// SetupWithManager configures the reconciler.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager, l logr.Logger, uuid string) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Log = l
	r.uuid = uuid
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{}).
		Complete(r)
}
