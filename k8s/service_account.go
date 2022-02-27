package k8s

import (
	"fmt"

	"github.com/urfave/cli/v2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete

func createServiceAccount(cliCtx cli.Context, k8sClient client.Client, namespace string) error {
	serviceAccount := getServiceAccount(namespace)

	clusterRoleBinding := getClusterRoleBinding(serviceAccount)

	err := k8sClient.Create(cliCtx.Context, serviceAccount)
	if !errors.IsAlreadyExists(err) {
		return fmt.Errorf("cannot create service account: %w", err)
	}

	err = k8sClient.Create(cliCtx.Context, clusterRoleBinding)
	if !errors.IsAlreadyExists(err) {
		return fmt.Errorf("cannot create cluster role binding: %w", err)
	}

	return nil
}

func getServiceAccount(namespace string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kopia-k8s",
			Namespace: namespace,
		},
	}
}

func getClusterRoleBinding(sa *v1.ServiceAccount) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kopia-k8s",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "kopia-k8s",
		},
	}
}
