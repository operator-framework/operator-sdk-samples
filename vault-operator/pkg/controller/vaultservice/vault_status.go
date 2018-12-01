package vaultservice

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"
	"github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/vaultutil"

	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ReconcileVaultService) updateVaultStatus(vr *vaultv1alpha1.VaultService, status *vaultv1alpha1.VaultServiceStatus) error {
	// don't update the status if there aren't any changes.
	if reflect.DeepEqual(vr.Status, *status) {
		return nil
	}
	vr.Status = *status
	return r.client.Update(context.TODO(), vr)
}

// getVaultStatus retrieves the status of the vault cluster for the given Custom Resource "vr",
// and it only succeeds if all of the nodes from vault cluster are reachable.
func (r *ReconcileVaultService) getVaultStatus(vr *vaultv1alpha1.VaultService, nsName types.NamespacedName) (*vaultv1alpha1.VaultServiceStatus, error) {
	pods := &corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}
	sel := vaultutil.LabelsForVault(vr.Name)
	opt := &client.ListOptions{
		Namespace:     vr.GetNamespace(),
		LabelSelector: labels.SelectorFromSet(sel),
	}
	err := r.client.List(context.TODO(), opt, pods)
	if err != nil {
		return nil, fmt.Errorf("failed to get vault's pods: %v", err)
	}

	tc, err := r.vaultTLSFromSecret(vr, nsName)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS config for vault client: %v", err)
	}

	var (
		initialized bool
		active      string
		standby     []string
		sealed      []string
		updated     []string
	)
	for _, p := range pods.Items {
		// If a pod is terminating, then we can't access the corresponding vault node's status.
		// so we break from here and return an error.
		if p.Status.Phase != corev1.PodRunning || p.DeletionTimestamp != nil {
			return nil, errors.New("vault pod is terminating")
		}

		vapi, err := vaultutil.NewVaultClient(vaultutil.PodDNSName(p), "8200", tc)
		if err != nil {
			return nil, fmt.Errorf("failed creating client for the vault pod (%s/%s): %v", vr.GetNamespace(), p.GetName(), err)
		}

		hr, err := vapi.Sys().Health()
		if err != nil {
			return nil, fmt.Errorf("failed requesting health info for the vault pod (%s/%s): %v", vr.GetNamespace(), p.GetName(), err)
		}

		if isVaultVersionMatch(p.Spec, vr.Spec) {
			updated = append(updated, p.GetName())
		}

		if hr.Initialized && !hr.Sealed && !hr.Standby {
			active = p.GetName()
		}
		if hr.Initialized && !hr.Sealed && hr.Standby {
			standby = append(standby, p.GetName())
		}
		if hr.Sealed {
			sealed = append(sealed, p.GetName())
		}
		if hr.Initialized {
			initialized = true
		}
	}

	return &vaultv1alpha1.VaultServiceStatus{
		Phase:       vaultv1alpha1.ClusterPhaseRunning,
		Initialized: initialized,
		ServiceName: vr.GetName(),
		ClientPort:  vaultClientPort,
		VaultStatus: vaultv1alpha1.VaultStatus{
			Active:  active,
			Standby: standby,
			Sealed:  sealed,
		},
		UpdatedNodes: updated,
	}, nil
}

// VaultTLSFromSecret reads Vault CR's TLS secret and converts it into a vault client's TLS config struct.
func (r *ReconcileVaultService) vaultTLSFromSecret(vr *vaultv1alpha1.VaultService, nsName types.NamespacedName) (*vaultapi.TLSConfig, error) {
	cs := vr.Spec.TLS.Static.ClientSecret
	se := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cs,
			Namespace: vr.GetNamespace(),
		},
	}
	err := r.client.Get(context.TODO(), nsName, se)
	if err != nil {
		return nil, fmt.Errorf("read client tls failed: failed to get secret (%s): %v", cs, err)
	}

	// Read the secret and write ca.crt to a temporary file
	caCertData := se.Data[vaultv1alpha1.CATLSCertName]
	f, err := ioutil.TempFile("", vaultv1alpha1.CATLSCertName)
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
