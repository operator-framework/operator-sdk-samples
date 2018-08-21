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
	goctx "context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	oputil "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/vault"
	"github.com/operator-framework/operator-sdk-samples/vault-operator/test/e2e/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"

	eopapi "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	"github.com/coreos/etcd-operator/pkg/util/etcdutil"
	eopk8sutil "github.com/coreos/etcd-operator/pkg/util/k8sutil"
	eope2eutil "github.com/coreos/etcd-operator/test/e2e/e2eutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// getEndpoints returns endpoints of an etcd cluster given the cluster name and the residing namespace.
func getEndpoints(kubeClient kubernetes.Interface, secureClient bool, namespace, clusterName string) ([]string, error) {
	podList, err := kubeClient.Core().Pods(namespace).List(eopk8sutil.ClusterListOpt(clusterName))
	if err != nil {
		return nil, err
	}

	var pods []*v1.Pod
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.Phase == v1.PodRunning {
			pods = append(pods, pod)
		}
	}

	if len(pods) == 0 {
		return nil, errors.New("no running etcd pods found")
	}

	endpoints := make([]string, len(pods))
	for i, pod := range pods {
		m := &etcdutil.Member{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			SecureClient: secureClient,
		}
		endpoints[i] = m.ClientURL()
	}
	return endpoints, nil
}

func createBackup(t *testing.T, vaultCR *api.VaultService, ctx framework.TestCtx, etcdClusterName, s3Path string) {
	f := framework.Global
	endpoints, err := getEndpoints(f.KubeClient, true, ctx.Namespace, etcdClusterName)
	if err != nil {
		t.Fatalf("failed to get endpoints: %v", err)
	}
	backupCR := eope2eutil.NewS3Backup(endpoints, etcdClusterName, s3Path, os.Getenv("TEST_AWS_SECRET"), oputil.EtcdClientTLSSecretName(vaultCR.Name))
	backupCR.SetNamespace(ctx.Namespace)
	if err != nil {
		t.Fatalf("failed to set namespace to etcd backup cd: %v", err)
	}
	err = f.DynamicClient.Create(goctx.TODO(), backupCR)
	if err != nil {
		t.Fatalf("failed to create etcd backup cr: %v", err)
	}
	eb := &eopapi.EtcdBackup{}
	err = f.DynamicClient.Get(goctx.TODO(), types.NamespacedName{Namespace: backupCR.Namespace, Name: backupCR.Name}, eb)
	if err != nil {
		t.Fatalf("failed to get etcd backup cr: %v", err)
	}
	defer func() {
		if err := f.DynamicClient.Delete(goctx.TODO(), eb); err != nil {
			t.Fatalf("failed to delete etcd backup cr: %v", err)
		}
	}()

	// local testing shows that it takes around 1 - 2 seconds from creating backup cr to verifying the backup from s3.
	// 4 seconds timeout via retry is enough; duration longer than that may indicate internal issues and
	// is worthy of investigation.
	err = wait.PollImmediate(time.Second, 4*time.Second, func() (bool, error) {
		reb := &eopapi.EtcdBackup{}
		err := f.DynamicClient.Get(goctx.TODO(), types.NamespacedName{Namespace: eb.Namespace, Name: eb.Name}, reb)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve backup CR: %v", err)
		}
		if reb.Status.Succeeded {
			if reb.Status.EtcdVersion == eopapi.DefaultEtcdVersion && reb.Status.EtcdRevision > 1 {
				return true, nil
			}
			return false, fmt.Errorf("expect EtcdVersion==%v and EtcdRevision > 1, but got EtcdVersion==%v and EtcdRevision==%v", eopapi.DefaultEtcdVersion, reb.Status.EtcdVersion, reb.Status.EtcdRevision)
		}
		if len(reb.Status.Reason) != 0 {
			return false, fmt.Errorf("backup failed with reason: %v ", reb.Status.Reason)
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("failed to verify backup: %v", err)
	}
	t.Logf("backup for cluster (%s) has been saved", etcdClusterName)
}

func killEtcdCluster(t *testing.T, ctx framework.TestCtx, etcdClusterName string) {
	f := framework.Global
	lops := eopk8sutil.ClusterListOpt(etcdClusterName)
	err := f.KubeClient.CoreV1().Pods(ctx.Namespace).DeleteCollection(metav1.NewDeleteOptions(0), lops)
	if err != nil {
		t.Fatalf("failed to delete etcd cluster pods: %v", err)
	}
	if _, err := e2eutil.WaitPodsDeletedCompletely(f.KubeClient, ctx.Namespace, 6, lops); err != nil {
		t.Fatalf("failed to see the etcd cluster pods to be completely removed: %v", err)
	}
}

// restoreEtcdCluster restores an etcd cluster with name "etcdClusterName" from a backup saved on "s3Path".
func restoreEtcdCluster(t *testing.T, ctx framework.TestCtx, s3Path, etcdClusterName string) {
	f := framework.Global
	restoreSource := eopapi.RestoreSource{S3: eope2eutil.NewS3RestoreSource(s3Path, os.Getenv("TEST_AWS_SECRET"))}
	er := eope2eutil.NewEtcdRestore(etcdClusterName, 3, restoreSource, eopapi.BackupStorageTypeS3)
	err := f.DynamicClient.Create(goctx.TODO(), er)
	if err != nil {
		t.Fatalf("failed to create etcd restore cr: %v", err)
	}
	defer func() {
		if err := f.DynamicClient.Delete(goctx.TODO(), er); err != nil {
			t.Fatalf("failed to delete etcd restore cr: %v", err)
		}
	}()

	err = wait.Poll(10*time.Second, 10*time.Second, func() (bool, error) {
		err = f.DynamicClient.Get(goctx.TODO(), types.NamespacedName{Namespace: er.Namespace, Name: er.Name}, er)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve restore CR: %v", err)
		}
		if er.Status.Succeeded {
			return true, nil
		} else if len(er.Status.Reason) != 0 {
			return false, fmt.Errorf("restore failed with reason: %v ", er.Status.Reason)
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("failed to verify restore succeeded: %v", err)
	}

	// Verify that the restored etcd cluster scales to 3 ready members
	restoredCluster := &eopapi.EtcdCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      etcdClusterName,
			Namespace: ctx.Namespace,
		},
		Spec: eopapi.ClusterSpec{
			Size: 3,
		},
	}
	if _, err := e2eutil.EtcdWaitUntilSizeReached(t, f.DynamicClient, 3, 6, restoredCluster); err != nil {
		t.Fatalf("failed to see restored etcd cluster(%v) reach 3 members: %v", restoredCluster.Name, err)
	}
}

// verifyRestoredVault ensures that the vault cluster that's restored from an earlier backup contains the correct data:
func verifyRestoredVault(t *testing.T, vaultCR *api.VaultService, ctx framework.TestCtx, secretData map[string]interface{}, keyPath, rootToken string) {
	f := framework.Global

	vaultCR, tlsConfig := e2eutil.WaitForCluster(t, f.KubeClient, f.DynamicClient, vaultCR)
	vaultCR, err := e2eutil.WaitActiveVaultsUp(t, f.DynamicClient, 6, vaultCR)
	if err != nil {
		t.Fatalf("failed to wait for any node to become active: %v", err)
	}

	podName := vaultCR.Status.VaultStatus.Active
	vClient := e2eutil.SetupVaultClient(t, f.KubeClient, ctx.Namespace, tlsConfig, podName)
	vClient.SetToken(rootToken)
	e2eutil.VerifySecretData(t, vClient, secretData, keyPath, podName)
}

func TestBackupRestoreOnVault(t *testing.T) {
	return
	f := framework.Global
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup(t)
	err := ctx.InitializeClusterResources()
	if err != nil {
		t.Fatalf("could not initialize cluster resources: %v", err)
	}
	s3Path := path.Join(os.Getenv("TEST_S3_BUCKET"), "jenkins", strconv.Itoa(int(rand.Uint64())), time.Now().Format(time.RFC3339), "etcd.backup")

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

	etcdClusterName := oputil.EtcdNameForVault(vaultCR.Name)
	createBackup(t, vaultCR, ctx, etcdClusterName, s3Path)

	killEtcdCluster(t, ctx, etcdClusterName)
	restoreEtcdCluster(t, ctx, s3Path, etcdClusterName)
	verifyRestoredVault(t, vaultCR, ctx, secretData, keyPath, rootToken)
}
