package e2e

import (
	"fmt"
	"testing"

	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/operator-framework/operator-sdk-samples/vault-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestScaleUp(t *testing.T) {
	f := framework.Global
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	err := ctx.InitializeClusterResources()
	if err != nil {
		t.Fatalf("could not initialize cluster resources: %v", err)
	}

	vaultServiceList := &api.VaultServiceList{}
	err = framework.AddToFrameworkScheme(api.AddToScheme, vaultServiceList)
	if err != nil {
		t.Fatalf("could not add scheme to framework scheme: %v", err)
	}

	vaultCR, err := e2eutil.CreateCluster(t, f.DynamicClient, e2eutil.NewCluster("test-vault-", ctx.Namespace, 1))
	if err != nil {
		t.Fatalf("failed to create vault cluster: %v", err)
	}
	ctx.AddFinalizerFn(func() error {
		if err := e2eutil.DeleteCluster(t, f.DynamicClient, vaultCR); err != nil {
			return fmt.Errorf("failed to delete vault cluster: %v", err)
		}
		return nil
	})

	vaultCR, tlsConfig := e2eutil.WaitForCluster(t, f.KubeClient, f.DynamicClient, vaultCR)

	// Init vault via the first sealed node
	podName := vaultCR.Status.VaultStatus.Sealed[0]
	vClient := e2eutil.SetupVaultClient(t, f.KubeClient, ctx.Namespace, tlsConfig, podName)
	vaultCR, initResp := e2eutil.InitializeVault(t, f.DynamicClient, vaultCR, vClient)

	// Unseal the vault node and wait for it to become active
	podName = vaultCR.Status.VaultStatus.Sealed[0]
	vClient = e2eutil.SetupVaultClient(t, f.KubeClient, ctx.Namespace, tlsConfig, podName)
	if err := e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	vaultCR, err = e2eutil.WaitActiveVaultsUp(t, f.DynamicClient, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for any node to become active: %v", err)
	}

	// TODO: Write secret to active node, read secret from new node later

	// Resize cluster to 2 nodes
	vaultCR, err = e2eutil.ResizeCluster(t, f.DynamicClient, vaultCR, 2)
	if err != nil {
		t.Fatalf("failed to resize vault cluster: %v", err)
	}

	// Wait for 1 unsealed node and create a vault client for it
	vaultCR, err = e2eutil.WaitSealedVaultsUp(t, f.DynamicClient, 1, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for vault nodes to become sealed: %v", err)
	}

	podName = vaultCR.Status.VaultStatus.Sealed[0]
	// Unseal the new node and wait for it to become standby
	vClient = e2eutil.SetupVaultClient(t, f.KubeClient, ctx.Namespace, tlsConfig, podName)
	if err := e2eutil.UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	vaultCR, err = e2eutil.WaitStandbyVaultsUp(t, f.DynamicClient, 1, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for vault nodes to become standby: %v", err)
	}

}
