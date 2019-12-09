package apis

import (
	v1alpha12 "github.com/operator-framework/operator-sdk-samples/go/memcached-operator/pkg/apis/cache/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha12.SchemeBuilder.AddToScheme)
}
