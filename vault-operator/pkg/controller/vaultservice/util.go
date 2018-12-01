package vaultservice

import (
	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

// asOwner returns an owner reference set as the vault cluster CR
func asOwner(v *vaultv1alpha1.VaultService) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: vaultv1alpha1.SchemeGroupVersion.String(),
		Kind:       "VaultService",
		Name:       v.Name,
		UID:        v.UID,
		Controller: &trueVar,
	}
}
