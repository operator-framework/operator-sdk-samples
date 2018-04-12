package stub

import (
	"fmt"

	api "github.com/coreos-inc/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/sirupsen/logrus"

	eopapi "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/query"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	size = 3
)

// reconcileVault reconciles the vault cluster's state to the spec specified by vr
// by preparing the TLS secrets, deploying the etcd and vault cluster,
// and finally updating the vault deployment if needed.
func reconcileVault(vr *api.VaultService) (err error) {
	vr = vr.DeepCopy()
	// Simulate initializer.
	changed := vr.SetDefaults()
	if changed {
		return action.Update(vr)
	}
	// After first time reconcile, phase will switch to "Running".
	if vr.Status.Phase == api.ClusterPhaseInitial {
		err = prepareEtcdTLSSecrets(vr)
		if err != nil {
			return err
		}
		// etcd cluster should only be created in first time reconcile.
		ec, err := deployEtcdCluster(vr)
		if err != nil {
			return err
		}
		// Check if etcd cluster is up and running.
		// If not, we need to wait until etcd cluster is up before proceeding to the next state;
		// Hence, we return from here and let the Watch triggers the handler again.
		ready, err := isEtcdClusterReady(ec)
		if err != nil {
			return fmt.Errorf("failed to check if etcd cluster is ready: %v", err)
		}
		if !ready {
			logrus.Infof("Waiting for EtcdCluster (%v) to become ready: %v", ec.Name)
			return nil
		}
	}

	err = prepareDefaultVaultTLSSecrets(vr)
	if err != nil {
		return err
	}

	// TODO: Deploy vault
	return nil
}

func isEtcdClusterReady(ec *eopapi.EtcdCluster) (bool, error) {
	err := query.Get(ec)
	if err != nil {
		return false, err
	}
	return (len(ec.Status.Members.Ready) < size), nil
}

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

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

// labelsForVault returns the labels for selecting the resources
// belonging to the given vault name.
func labelsForVault(name string) map[string]string {
	return map[string]string{"app": "vault", "vault_cluster": name}
}

// asOwner returns an owner reference set as the vault cluster CR
func asOwner(v *api.VaultService) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.VaultServiceKind,
		Name:       v.Name,
		UID:        v.UID,
		Controller: &trueVar,
	}
}
