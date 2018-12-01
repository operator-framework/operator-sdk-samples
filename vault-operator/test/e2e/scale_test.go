package e2e

import (
	"fmt"
	"testing"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/operator-framework/operator-sdk-samples/vault-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestScaleUp(t *testing.T) {
	f := framework.Global
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(nil)
	if err != nil {
		t.Fatalf("could not initialize cluster resources: %v", err)
	}
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("failed to get namespace: %v", err)
	}

	vaultServiceList := &vaultv1alpha1.VaultServiceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VaultService",
			APIVersion: "vault.security.coreos.com/v1alpha1",
		},
	}
	err = framework.AddToFrameworkScheme(vaultv1alpha1.SchemeBuilder.AddToScheme, vaultServiceList)
	if err != nil {
		t.Fatalf("could not add scheme to framework scheme: %v", err)
	}

	vaultCR, err := e2eutil.CreateCluster(t, f, e2eutil.NewCluster("test-vault-", namespace, 1))
	if err != nil {
		t.Fatalf("failed to create vault cluster: %v", err)
	}
	ctx.AddCleanupFn(func() error {
		if err := e2eutil.DeleteCluster(t, f, vaultCR); err != nil {
			return fmt.Errorf("failed to delete vault cluster: %v", err)
		}
		return nil
	})

	vaultCR, tlsConfig := e2eutil.WaitForCluster(t, f, vaultCR)

	// Init vault via the first sealed node
	podName := vaultCR.Status.VaultStatus.Sealed[0]
	vClient := e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	vaultCR, initResp := e2eutil.InitializeVault(t, f, vaultCR, vClient)

	// Unseal the vault node and wait for it to become active
	podName = vaultCR.Status.VaultStatus.Sealed[0]
	vClient = e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	if err := e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	vaultCR, err = e2eutil.WaitActiveVaultsUp(t, f, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for any node to become active: %v", err)
	}

	// TODO: Write secret to active node, read secret from new node later

	// Resize cluster to 2 nodes
	vaultCR, err = e2eutil.ResizeCluster(t, f, vaultCR, 2)
	if err != nil {
		t.Fatalf("failed to resize vault cluster: %v", err)
	}

	// Wait for 1 unsealed node and create a vault client for it
	vaultCR, err = e2eutil.WaitSealedVaultsUp(t, f, 1, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for vault nodes to become sealed: %v", err)
	}

	podName = vaultCR.Status.VaultStatus.Sealed[0]
	// Unseal the new node and wait for it to become standby
	vClient = e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	if err := e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	vaultCR, err = e2eutil.WaitStandbyVaultsUp(t, f, 1, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for vault nodes to become standby: %v", err)
	}

}
