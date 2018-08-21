// Copyright 2018 The vault-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
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
	defer func(vaultCR *api.VaultService) {
		if err := e2eutil.DeleteCluster(t, f.DynamicClient, vaultCR); err != nil {
			t.Fatalf("failed to delete vault cluster: %v", err)
		}
	}(vaultCR)
	vClient, keyPath, secretData, podName := e2eutil.WriteSecretData(t, vaultCR, f.KubeClient, tlsConfig, rootToken, ctx.Namespace)
	e2eutil.VerifySecretData(t, vClient, secretData, keyPath, podName)
}
