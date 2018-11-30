package apis

import (
	"github.com/operator-framework/operator-sdk-samples/vault-operator/pkg/apis/vault/v1alpha1"

	eopapi "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	// register etcd operator's scheme
	AddToSchemes = append(AddToSchemes, eopapi.AddToScheme)
}
