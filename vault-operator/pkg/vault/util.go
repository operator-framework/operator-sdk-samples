package vault

import (
	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
