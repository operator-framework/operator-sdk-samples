package vault

import (
	"fmt"

	api "github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	"k8s.io/api/core/v1"
)

func isVaultVersionMatch(ps v1.PodSpec, vs api.VaultServiceSpec) bool {
	return ps.Containers[0].Image == vaultImage(vs)
}

func vaultImage(vs api.VaultServiceSpec) string {
	return fmt.Sprintf("%s:%s", vs.BaseImage, vs.Version)
}
