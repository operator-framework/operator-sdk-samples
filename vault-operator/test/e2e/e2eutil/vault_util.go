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

package e2eutil

import (
	goctx "context"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	oputil "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/vault"
	runtime "sigs.k8s.io/controller-runtime/pkg/client"

	vaultapi "github.com/hashicorp/vault/api"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const targetVaultPort = "8200"

// WaitForCluster waits for all available nodes of a cluster to appear in the vault CR status
// Returns the updated vault cluster and the TLS configuration to use for vault clients interacting with the cluster
func WaitForCluster(t *testing.T, kubeClient kubernetes.Interface, dynClient runtime.Client, vaultCR *api.VaultService) (*api.VaultService, *vaultapi.TLSConfig) {
	// Based on local testing, it took about ~50s for a normal deployment to finish.
	vaultCR, err := WaitAvailableVaultsUp(t, dynClient, int(vaultCR.Spec.Nodes), 10, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for cluster nodes to become available: %v", err)
	}

	tlsConfig, err := VaultTLSFromSecret(vaultCR, dynClient)
	if err != nil {
		t.Fatalf("failed to read TLS config for vault client: %v", err)
	}
	return vaultCR, tlsConfig
}

// InitializeVault initializes the specified vault cluster and waits for all available nodes to appear as sealed.
// Requires established portforwarded connections to the vault pods
// Returns the updated vault cluster and the initialization response which includes the unseal key
func InitializeVault(t *testing.T, dynClient runtime.Client, vault *api.VaultService, vClient *vaultapi.Client) (*api.VaultService, *vaultapi.InitResponse) {
	initOpts := &vaultapi.InitRequest{SecretShares: 1, SecretThreshold: 1}
	initResp, err := vClient.Sys().Init(initOpts)
	if err != nil {
		t.Fatalf("failed to initialize vault: %v", err)
	}
	// Wait until initialized nodes to be reflected on status.vaultStatus.Sealed
	vault, err = WaitSealedVaultsUp(t, dynClient, int(vault.Spec.Nodes), 6, vault)
	if err != nil {
		t.Fatalf("failed to wait for vault nodes to become sealed: %v", err)
	}
	return vault, initResp
}

// UnsealVaultNode unseals the specified vault pod by portforwarding to it via its vault client
func UnsealVaultNode(unsealKey string, vClient *vaultapi.Client) error {
	unsealResp, err := vClient.Sys().Unseal(unsealKey)
	if err != nil {
		return fmt.Errorf("failed to unseal vault: %v", err)
	}
	if unsealResp.Sealed {
		return fmt.Errorf("failed to unseal vault: unseal response still shows vault as sealed")
	}
	return nil
}

// Portforwarding is necessary if outside the cluster. This version of e2eutil in the vault-operator repo
// contained a port-forwarding mechanism: https://github.com/coreos/vault-operator/tree/e5d03827065b1429c163e8a5ed69c32c8d9a3046/test/e2e/e2eutil
// SetupVaultClient creates a vault client for the specified pod
func SetupVaultClient(t *testing.T, kubeClient kubernetes.Interface, namespace string, tlsConfig *vaultapi.TLSConfig, podName string) *vaultapi.Client {
	pod, err := kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("fail to get vault pod (%s): %v", podName, err)
	}
	vClient, err := oputil.NewVaultClient(oputil.PodDNSName(*pod), targetVaultPort, tlsConfig)
	if err != nil {
		t.Fatalf("failed creating vault client for (localhost:%v): %v", targetVaultPort, err)
	}
	return vClient
}

// SetupUnsealedVaultCluster initializes a vault cluster and unseals the 1st vault node.
func SetupUnsealedVaultCluster(t *testing.T, kubeClient kubernetes.Interface, dynClient runtime.Client, namespace string) (*api.VaultService, *vaultapi.TLSConfig, string) {
	vaultCR, err := CreateCluster(t, dynClient, NewCluster("test-vault-", namespace, 2))
	if err != nil {
		t.Fatalf("failed to create vault cluster: %v", err)
	}

	vaultCR, tlsConfig := WaitForCluster(t, kubeClient, dynClient, vaultCR)

	// Init vault via the first sealed node
	podName := vaultCR.Status.VaultStatus.Sealed[0]
	vClient := SetupVaultClient(t, kubeClient, namespace, tlsConfig, podName)
	vaultCR, initResp := InitializeVault(t, dynClient, vaultCR, vClient)

	// Unseal the 1st vault node and wait for it to become active
	podName = vaultCR.Status.VaultStatus.Sealed[0]
	vClient = SetupVaultClient(t, kubeClient, namespace, tlsConfig, podName)
	if err := UnsealVaultNode(initResp.Keys[0], vClient); err != nil {
		t.Fatalf("failed to unseal vault node(%v): %v", podName, err)
	}
	vaultCR, err = WaitActiveVaultsUp(t, dynClient, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for any node to become active: %v", err)
	}

	return vaultCR, tlsConfig, initResp.RootToken
}

// WriteSecretData writes secret data into vault.
func WriteSecretData(t *testing.T, vaultCR *api.VaultService, kubeClient kubernetes.Interface, tlsConfig *vaultapi.TLSConfig, rootToken, namespace string) (*vaultapi.Client, string, map[string]interface{}, string) {
	// Write secret to active node
	podName := vaultCR.Status.VaultStatus.Active
	vClient := SetupVaultClient(t, kubeClient, namespace, tlsConfig, podName)
	vClient.SetToken(rootToken)

	keyPath := "secret/login"
	data := &SampleSecret{Username: "user", Password: "pass"}
	secretData, err := MapObjectToArbitraryData(data)
	if err != nil {
		t.Fatalf("failed to create secret data (%+v): %v", data, err)
	}

	_, err = vClient.Logical().Write(keyPath, secretData)
	if err != nil {
		t.Fatalf("failed to write secret (%v) to vault node (%v): %v", keyPath, podName, err)
	}
	return vClient, keyPath, secretData, podName
}

// VerifySecretData gets secret of the "keyPath" and compares it against the given secretData.
func VerifySecretData(t *testing.T, vClient *vaultapi.Client, secretData map[string]interface{}, keyPath, podName string) {
	// Read secret back from active node
	secret, err := vClient.Logical().Read(keyPath)
	if err != nil {
		t.Fatalf("failed to read secret(%v) from vault node (%v): %v", keyPath, podName, err)
	}

	if !reflect.DeepEqual(secret.Data, secretData) {
		t.Fatalf("Read secret data (%+v) is not the same as written secret (%+v)", secret.Data, secretData)
	}
}

// VaultTLSFromSecret reads Vault CR's TLS secret and converts it into a vault client's TLS config struct.
func VaultTLSFromSecret(vr *api.VaultService, dynClient runtime.Client) (*vaultapi.TLSConfig, error) {
	cs := vr.Spec.TLS.Static.ClientSecret
	se := &v1.Secret{}
	err := dynClient.Get(goctx.TODO(), types.NamespacedName{Namespace: vr.Namespace, Name: cs}, se)
	if err != nil {
		return nil, fmt.Errorf("read client tls failed: failed to get secret (%s): %v", cs, err)
	}

	// Read the secret and write ca.crt to a temporary file
	caCertData := se.Data[api.CATLSCertName]
	f, err := ioutil.TempFile("", api.CATLSCertName)
	if err != nil {
		return nil, fmt.Errorf("read client tls failed: create temp file failed: %v", err)
	}
	defer f.Close()

	_, err = f.Write(caCertData)
	if err != nil {
		return nil, fmt.Errorf("read client tls failed: write ca cert file failed: %v", err)
	}
	if err = f.Sync(); err != nil {
		return nil, fmt.Errorf("read client tls failed: sync ca cert file failed: %v", err)
	}
	return &vaultapi.TLSConfig{CACert: f.Name()}, nil
}
