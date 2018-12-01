package e2eutil

import (
	goctx "context"
	"fmt"
	"testing"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"k8s.io/apimachinery/pkg/types"
)

// CreateCluster creates a vault CR with the desired spec
func CreateCluster(t *testing.T, f *framework.Framework, vs *vaultv1alpha1.VaultService) (*vaultv1alpha1.VaultService, error) {
	err := f.Client.Create(goctx.TODO(), vs, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create CR: %v", err)
	}
	vault := &vaultv1alpha1.VaultService{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Namespace: vs.Namespace, Name: vs.Name}, vault)
	if err != nil {
		return nil, fmt.Errorf("failed to get CR: %v", err)
	}
	LogfWithTimestamp(t, "created vault cluster: %s", vault.Name)
	return vault, nil
}

// ResizeCluster updates the Nodes field of the vault CR
func ResizeCluster(t *testing.T, f *framework.Framework, vs *vaultv1alpha1.VaultService, size int) (*vaultv1alpha1.VaultService, error) {
	vault := &vaultv1alpha1.VaultService{}
	namespacedName := types.NamespacedName{Namespace: vs.GetNamespace(), Name: vs.GetName()}
	err := f.Client.Get(goctx.TODO(), namespacedName, vault)
	if err != nil {
		return nil, fmt.Errorf("failed to get CR: %v", err)
	}
	vault.Spec.Nodes = int32(size)
	err = f.Client.Update(goctx.TODO(), vault)
	if err != nil {
		return nil, fmt.Errorf("failed to update CR: %v", err)
	}
	LogfWithTimestamp(t, "updated vault cluster(%v) to size(%v)", vault.Name, size)
	return vault, nil
}

// UpdateVersion updates the Version field of the vault CR
func UpdateVersion(t *testing.T, f *framework.Framework, vs *vaultv1alpha1.VaultService, version string) (*vaultv1alpha1.VaultService, error) {
	vault := &vaultv1alpha1.VaultService{}
	namespacedName := types.NamespacedName{Namespace: vs.GetNamespace(), Name: vs.GetName()}
	err := f.Client.Get(goctx.TODO(), namespacedName, vault)
	if err != nil {
		return nil, fmt.Errorf("failed to get CR: %v", err)
	}
	vault.Spec.Version = version
	err = f.Client.Update(goctx.TODO(), vault)
	if err != nil {
		return nil, fmt.Errorf("failed to update CR: %v", err)
	}
	LogfWithTimestamp(t, "updated vault cluster(%v) to version(%v)", vault.Name, version)
	return vault, nil
}

// DeleteCluster deletes the vault CR specified by cluster spec
func DeleteCluster(t *testing.T, f *framework.Framework, vs *vaultv1alpha1.VaultService) error {
	t.Logf("deleting vault cluster: %v", vs.Name)
	err := f.Client.Delete(goctx.TODO(), vs)
	if err != nil {
		return fmt.Errorf("failed to delete CR: %v", err)
	}
	// TODO: Wait for cluster resources to be deleted
	return nil
}
