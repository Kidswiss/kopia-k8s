package k8s

import (
	"fmt"

	"git.earthnet.ch/simon.beck/kopia-k8s/logger"
	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;list;watch;create;update;patch;delete

// listPodsWithPVCs returns a list with pods that have a PVC mounted.
func listPodsWithPVCs(cliCtx *cli.Context, k8sClient client.Client) (*v1.PodList, error) {
	tmp := &v1.PodList{}

	selector, err := createLabelSelector()
	if err != nil {
		return nil, err
	}

	err = k8sClient.List(cliCtx.Context, tmp, &client.ListOptions{LabelSelector: selector})

	pods := &v1.PodList{}
	for _, pod := range tmp.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && pod.Status.Phase == v1.PodRunning {
				pods.Items = append(pods.Items, pod)
				break
			}
		}
	}

	return pods, err
}

func createLabelSelector() (labels.Selector, error) {
	podReq, err := labels.NewRequirement(JobLabel, selection.DoesNotExist, []string{})

	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector().Add(*podReq)
	return selector, err
}

func listPodsWithPrebackupAnnotation(cliCtx *cli.Context, k8sClient client.Client) (*v1.PodList, error) {
	tmp := &v1.PodList{}
	err := k8sClient.List(cliCtx.Context, tmp)
	if err != nil {
		return nil, err
	}

	annotation := cliCtx.String("pre-backup-annotation")

	pods := &v1.PodList{}
	for _, pod := range tmp.Items {
		if _, ok := pod.Annotations[annotation]; ok {
			pods.Items = append(pods.Items, pod)
		}
	}
	return pods, nil
}

// ExecutePrebackupCommand rund prebackup commands on the pods before actually starting the backup
func ExecutePrebackupCommand(cliCtx *cli.Context, k8sClient client.Client) error {
	log := logger.AppLogger(cliCtx.Context).WithName("prebackupExec")

	pods, err := listPodsWithPrebackupAnnotation(cliCtx, k8sClient)
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		log.Info("running prebackup command", "podname", pod.Name, "command", pod.Annotations[cliCtx.String("pre-backup-annotation")])
		err := execPod(cliCtx, &pod)
		if err != nil {
			return err
		}
	}

	return nil
}

func getClientConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		err1 := err
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		config, err = kubeconfig.ClientConfig()
		if err != nil {
			err = fmt.Errorf("InClusterConfig as well as BuildConfigFromFlags Failed. Error in InClusterConfig: %+v\nError in BuildConfigFromFlags: %+v", err1, err)
			return nil, err
		}
	}

	return config, nil
}

// func getKubeClient() error {

// 	client := kube.NewDefault()
// 	_, err := client.Dial()
// 	if err != nil {
// 		return err
// 	}
// 	client.Exec(pod string, container string, namespace string, command string)
// 	return nil

// }

func execPod(cliCtx *cli.Context, pod *v1.Pod) error {
	cmd := []string{
		"sh",
		"-c",
		pod.Annotations[cliCtx.String("pre-backup-annotation")],
	}

	config, err := getClientConfig()
	if err != nil {
		return fmt.Errorf("cannot get rest client config: %w", err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("cannot get rest client: %w", err)
	}

	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).SubResource("exec")
	option := &v1.PodExecOptions{
		Command:   cmd,
		Container: pod.Spec.Containers[0].Name,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	execLog := execLogger{
		log:       logger.AppLogger(cliCtx.Context).WithName("k8sexec"),
		podname:   pod.Name,
		namespace: pod.Namespace,
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: logger.New(execLog.execStdout),
		Stderr: logger.New(execLog.execStderr),
	})
	if err != nil {
		return fmt.Errorf("can't exec %s: %w", pod.Name, err)
	}

	return nil
}
