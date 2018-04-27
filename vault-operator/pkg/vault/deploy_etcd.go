package vault

import (
	"fmt"

	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	eopapi "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	"github.com/operator-framework/operator-sdk/pkg/sdk/action"
	"github.com/operator-framework/operator-sdk/pkg/sdk/query"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	size = 3
)

// deployEtcdCluster creates an etcd cluster for the given vault's name via etcd operator.
func deployEtcdCluster(v *api.VaultService) (*eopapi.EtcdCluster, error) {
	ec := &eopapi.EtcdCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       eopapi.EtcdClusterResourceKind,
			APIVersion: eopapi.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      etcdNameForVault(v.Name),
			Namespace: v.Namespace,
			Labels:    labelsForVault(v.Name),
		},
		Spec: eopapi.ClusterSpec{
			Size: size,
			TLS: &eopapi.TLSPolicy{
				Static: &eopapi.StaticTLS{
					Member: &eopapi.MemberSecret{
						PeerSecret:   etcdPeerTLSSecretName(v.Name),
						ServerSecret: etcdServerTLSSecretName(v.Name),
					},
					OperatorSecret: etcdClientTLSSecretName(v.Name),
				},
			},
			Pod: &eopapi.PodPolicy{
				EtcdEnv: []v1.EnvVar{{
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
	err := action.Create(ec)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return ec, nil
		}
		return nil, fmt.Errorf("deploy etcd cluster failed: %v", err)
	}
	return ec, nil
}

// etcdNameForVault returns the etcd cluster's name for the given vault's name
func etcdNameForVault(name string) string {
	return name + "-etcd"
}

// etcdURLForVault returns the URL to talk to etcd cluster for the given vault's name
func etcdURLForVault(name string) string {
	return fmt.Sprintf("https://%s-client:2379", etcdNameForVault(name))
}

func isEtcdClusterReady(ec *eopapi.EtcdCluster) (bool, error) {
	err := query.Get(ec)
	if err != nil {
		return false, err
	}
	return (len(ec.Status.Members.Ready) == size), nil
}
