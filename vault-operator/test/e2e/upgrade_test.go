package e2e

import (
	"fmt"
	"testing"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	"github.com/operator-framework/operator-sdk-samples/vault-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpgradeVault(t *testing.T) {
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

	vaultCR := e2eutil.NewCluster("test-vault-", namespace, 2)
	vaultCR.Spec.Version = "0.9.1-0"
	vaultCR, err = e2eutil.CreateCluster(t, f, vaultCR)
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

	// Initialize vault via the 1st sealed node
	podName := vaultCR.Status.VaultStatus.Sealed[0]
	vClient := e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	vaultCR, initResp := e2eutil.InitializeVault(t, f, vaultCR, vClient)

	// Unseal the 1st vault node and wait for it to become active
	podName = vaultCR.Status.VaultStatus.Sealed[0]
	vClient = e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	if err := e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	vaultCR, err = e2eutil.WaitActiveVaultsUp(t, f, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for any node to become active: %v", err)
	}

	// Unseal the 2nd vault node and wait for it to become standby
	podName = vaultCR.Status.VaultStatus.Sealed[0]
	vClient = e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	if err = e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	vaultCR, err = e2eutil.WaitStandbyVaultsUp(t, f, 1, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for vault nodes to become standby: %v", err)
	}

	// Upgrade vault version
	newVersion := "0.9.1-1"
	vaultCR, err = e2eutil.UpdateVersion(t, f, vaultCR, newVersion)
	if err != nil {
		t.Fatalf("failed to update vault version: %v", err)
	}

	// Check for 2 sealed nodes
	vaultCR, err = e2eutil.WaitSealedVaultsUp(t, f, 2, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for updated sealed vault nodes: %v", err)
	}

	podName = vaultCR.Status.VaultStatus.Sealed[0]
	vClient = e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	if err = e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	podName = vaultCR.Status.VaultStatus.Sealed[1]
	vClient = e2eutil.SetupVaultClient(t, f, namespace, tlsConfig, podName)
	if err = e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}

	upgradedNodes := vaultCR.Status.VaultStatus.Sealed

	// Check that the active node is one of the newly unsealed nodes
	vaultCR, err = e2eutil.WaitUntilActiveIsFrom(t, f, 6, vaultCR, upgradedNodes...)
	if err != nil {
		t.Fatalf("failed to see the active node to be from the newly unsealed pods (%v): %v", upgradedNodes, err)
	}

	// Check that the standby node(s) are all from the newly unsealed nodes
	vaultCR, err = e2eutil.WaitUntilStandbyAreFrom(t, f, 6, vaultCR, upgradedNodes...)
	if err != nil {
		t.Fatalf("failed to see all the standby nodes to be from the newly unsealed pods (%v): %v", upgradedNodes, err)
	}

	// Check that the available nodes are all from the newly unsealed nodes, i.e the old nodes are deleted
	vaultCR, err = e2eutil.WaitUntilAvailableAreFrom(t, f, 6, vaultCR, upgradedNodes...)
	if err != nil {
		t.Fatalf("failed to see all available nodes to be from the newly unsealed pods (%v): %v", upgradedNodes, err)
	}

	// Check that 1 active and 1 standby are of the updated version
	err = e2eutil.CheckVersionReached(t, f, newVersion, 6, vaultCR, vaultCR.Status.VaultStatus.Active)
	if err != nil {
		t.Fatalf("failed to wait for active node to become updated: %v", err)
	}
	err = e2eutil.CheckVersionReached(t, f, newVersion, 6, vaultCR, vaultCR.Status.VaultStatus.Standby...)
	if err != nil {
		t.Fatalf("failed to wait for standby nodes to become updated: %v", err)
	}
}
