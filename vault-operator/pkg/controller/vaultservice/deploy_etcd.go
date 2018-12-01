package vaultservice

import (
	"context"
	"fmt"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/vaultutil"

	eopapi "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	size = 3
)

// deployEtcdCluster creates an etcd cluster for the given vault's name via etcd operator.
func (r *ReconcileVaultService) deployEtcdCluster(v *vaultv1alpha1.VaultService) (*eopapi.EtcdCluster, error) {
	ec := &eopapi.EtcdCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       eopapi.EtcdClusterResourceKind,
			APIVersion: eopapi.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      EtcdNameForVault(v.Name),
			Namespace: v.Namespace,
			Labels:    vaultutil.LabelsForVault(v.Name),
		},
		Spec: eopapi.ClusterSpec{
			Size: size,
			TLS: &eopapi.TLSPolicy{
				Static: &eopapi.StaticTLS{
					Member: &eopapi.MemberSecret{
						PeerSecret:   etcdPeerTLSSecretName(v.Name),
						ServerSecret: etcdServerTLSSecretName(v.Name),
					},
					OperatorSecret: EtcdClientTLSSecretName(v.Name),
				},
			},
			Pod: &eopapi.PodPolicy{
				EtcdEnv: []corev1.EnvVar{{
					Name:  "ETCD_AUTO_COMPACTION_RETENTION",
					Value: "1",
				}},
			},
		},
	}
	if v.Spec.Pod != nil {
		ec.Spec.Pod.Resources = v.Spec.Pod.Resources
	}
	addOwnerRefToObject(ec, asOwner(v))
	err := r.client.Create(context.TODO(), ec)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return ec, nil
		}
		return nil, fmt.Errorf("deploy etcd cluster failed: %v", err)
	}
	return ec, nil
}

// EtcdNameForVault returns the etcd cluster's name for the given vault's name
func EtcdNameForVault(name string) string {
	return name + "-etcd"
}

// etcdURLForVault returns the URL to talk to etcd cluster for the given vault's name
func etcdURLForVault(name string) string {
	return fmt.Sprintf("https://%s-client:2379", EtcdNameForVault(name))
}

func (r *ReconcileVaultService) isEtcdClusterReady(ec *eopapi.EtcdCluster, nsName types.NamespacedName) (bool, error) {
	err := r.client.Get(context.TODO(), nsName, ec)
	if err != nil {
		return false, err
	}
	return (len(ec.Status.Members.Ready) == size), nil
}
