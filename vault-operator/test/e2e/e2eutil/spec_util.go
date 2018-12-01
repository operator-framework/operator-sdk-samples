package e2eutil

import (
	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCluster returns a minimal vault cluster CR
func NewCluster(genName, namespace string, size int) *vaultv1alpha1.VaultService {
	return &vaultv1alpha1.VaultService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VaultService",
			APIVersion: vaultv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: genName,
			Namespace:    namespace,
		},
		Spec: vaultv1alpha1.VaultServiceSpec{
			Nodes: int32(size),
		},
	}
}
