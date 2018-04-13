package stub

import (
	"fmt"

	api "github.com/coreos-inc/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/query"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// syncVaultClusterSize ensures that the vault cluster is at the desired size.
func syncVaultClusterSize(vr *api.VaultService) error {
	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      vr.GetName(),
			Namespace: vr.GetNamespace(),
		},
	}
	err := query.Get(d)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", d.Name, err)
	}

	if *d.Spec.Replicas != vr.Spec.Nodes {
		d.Spec.Replicas = &(vr.Spec.Nodes)
		err = action.Update(d)
		if err != nil {
			return fmt.Errorf("failed to update size of deployment (%s): %v", d.Name, err)
		}
	}
	return nil
}
