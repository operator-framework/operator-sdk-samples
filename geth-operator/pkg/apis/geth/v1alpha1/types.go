package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GethNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []GethNode `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GethNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              GethNodeSpec   `json:"spec"`
	Status            GethNodeStatus `json:"status,omitempty"`
}

type GethNodeSpec struct {
	Size int32 `json:"size"`
}
type GethNodeStatus struct {
	Nodes []string `json:"nodes"`
}
