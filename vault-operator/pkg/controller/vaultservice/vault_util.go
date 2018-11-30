package vaultservice

import (
	"fmt"

	vaultv1alpha1 "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func isVaultVersionMatch(ps corev1.PodSpec, vs vaultv1alpha1.VaultServiceSpec) bool {
	return ps.Containers[0].Image == vaultImage(vs)
}

func vaultImage(vs vaultv1alpha1.VaultServiceSpec) string {
	return fmt.Sprintf("%s:%s", vs.BaseImage, vs.Version)
}
