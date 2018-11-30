package vaultservice

import (
	"context"
	"fmt"
	"reflect"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// syncVaultClusterSize ensures that the vault cluster is at the desired size.
func (r *ReconcileVaultService) syncVaultClusterSize(vr *vaultv1alpha1.VaultService, nsName types.NamespacedName) error {
	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      vr.GetName(),
			Namespace: vr.GetNamespace(),
		},
	}
	err := r.client.Get(context.TODO(), nsName, d)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", d.Name, err)
	}

	if *d.Spec.Replicas != vr.Spec.Nodes {
		d.Spec.Replicas = &(vr.Spec.Nodes)
		err = r.client.Update(context.TODO(), d)
		if err != nil {
			return fmt.Errorf("failed to update size of deployment (%s): %v", d.Name, err)
		}
	}
	return nil
}

func (r *ReconcileVaultService) syncUpgrade(vr *vaultv1alpha1.VaultService, status *vaultv1alpha1.VaultServiceStatus, nsName types.NamespacedName) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("syncUpgrade failed: %v", err)
		}
	}()

	d := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      vr.GetName(),
			Namespace: vr.GetNamespace(),
		},
	}
	err = r.client.Get(context.TODO(), nsName, d)
	if err != nil {
		return fmt.Errorf("failed to get deployment (%s): %v", d.Name, err)
	}

	// If the deployment version hasn't been updated, roll forward the deployment version
	// but keep the existing active Vault node alive though.
	if !isVaultVersionMatch(d.Spec.Template.Spec, vr.Spec) {
		err = r.upgradeDeployment(vr, d)
		if err != nil {
			return err
		}
	}

	var (
		active  = status.VaultStatus.Active
		standby = status.VaultStatus.Standby
		updated = status.UpdatedNodes
		sealed  = status.VaultStatus.Sealed
	)
	// If there is one active node belonging to the old version, and all other nodes are
	// standby and uptodate, then trigger step-down on active node.
	// It maps to the following conditions on Status:
	// 1. check standby == updated
	// 2. check Available - Updated == Active
	readyToTriggerStepdown := func() bool {
		if len(active) == 0 {
			return false
		}

		if !reflect.DeepEqual(standby, updated) {
			return false
		}

		ava := append(standby, sealed...)
		if !reflect.DeepEqual(ava, updated) {
			return false
		}
		return true
	}()

	if readyToTriggerStepdown {
		// This will send SIGTERM to the active Vault pod. It should release HA lock and exit properly.
		// If it failed for some reason, kubelet will send SIGKILL after default grace period (30s) eventually.
		// It take longer but the the lock will get released eventually on failure case.
		p := &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      active,
				Namespace: vr.GetNamespace(),
			},
		}
		err = r.client.Delete(context.TODO(), p)
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("step down: failed to delete active Vault pod (%s): %v", active, err)
		}
	}
	return nil
}

// upgradeDeployment sets deployment spec to:
// - roll forward version
// - keep active Vault node available by setting maxUnavailable=N-1
func (r *ReconcileVaultService) upgradeDeployment(vr *vaultv1alpha1.VaultService, d *appsv1.Deployment) error {
	mu := intstr.FromInt(int(vr.Spec.Nodes - 1))
	d.Spec.Strategy.RollingUpdate.MaxUnavailable = &mu
	d.Spec.Template.Spec.Containers[0].Image = vaultImage(vr.Spec)
	err := r.client.Update(context.TODO(), d)
	if err != nil {
		return fmt.Errorf("failed to upgrade deployment to (%s): %v", vaultImage(vr.Spec), err)
	}
	return nil
}
