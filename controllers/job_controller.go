package controllers

import (
	"context"

	"git.earthnet.ch/simon.beck/kopia-k8s/k8s"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	r.Log.V(1).Info("watched job", "name", req.Name)

	myJob := &batchv1.Job{}

	r.Client.Get(ctx, req.NamespacedName, myJob)

	r.Log.V(1).Info("job status", "status", myJob.Status)

	if myJob.ObjectMeta.Labels[k8s.JobLabel] == r.uuid {
		if myJob.Status.Succeeded > 0 && myJob.DeletionTimestamp == nil {
			backgroundDelete := v1.DeletePropagationBackground
			err := r.Client.Delete(ctx, myJob, &client.DeleteOptions{PropagationPolicy: &backgroundDelete})
			if err != nil {
				r.Log.Error(err, "job finished successfull, but cannot be cleaned up")
			} else {
				r.Log.Info("job finished successfully, cleaning up", "name", myJob.Name)
			}
			k8s.FinishedJobChannel <- true
		}
		if myJob.Status.Failed > 0 && myJob.DeletionTimestamp == nil {
			r.Log.Error(nil, "job failed, not cleaning up", "name", myJob.Name)
			k8s.FinishedJobChannel <- true
		}
	}

	return ctrl.Result{}, nil
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
