package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Label structure for instancepool node labels
type Label struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
