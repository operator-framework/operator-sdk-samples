package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DexList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Dex `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Dex struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DexSpec   `json:"spec"`
	Status            DexStatus `json:"status,omitempty"`
}

type DexSpec struct {
	// Size is the size of the memcached deployment
	Size int32 `json:"size"`
}
type DexStatus struct {
	// Nodes are the names of the memcached pods
	Nodes []string `json:"nodes"`
}
