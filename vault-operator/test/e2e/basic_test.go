package e2e

import (
	"fmt"
	"testing"

	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/operator-framework/operator-sdk-samples/vault-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateHAVault(t *testing.T) {
	f := framework.Global
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	err := ctx.InitializeClusterResources()
	if err != nil {
		t.Fatalf("could not initialize cluster resources: %v", err)
	}

	vaultServiceList := &api.VaultServiceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VaultService",
			APIVersion: "vault.security.coreos.com/v1alpha1",
		},
	}
	err = framework.AddToFrameworkScheme(api.AddToScheme, vaultServiceList)
	if err != nil {
		t.Fatalf("could not add scheme to framework scheme: %v", err)
	}

	vaultCR, tlsConfig, rootToken := e2eutil.SetupUnsealedVaultCluster(t, f.KubeClient, f.DynamicClient, ctx.Namespace)
	ctx.AddFinalizerFn(func() error {
		if err := e2eutil.DeleteCluster(t, f.DynamicClient, vaultCR); err != nil {
			return fmt.Errorf("failed to delete vault cluster: %v", err)
		}
		return nil
	})
	vClient, keyPath, secretData, podName := e2eutil.WriteSecretData(t, vaultCR, f.KubeClient, tlsConfig, rootToken, ctx.Namespace)
	e2eutil.VerifySecretData(t, vClient, secretData, keyPath, podName)
}
