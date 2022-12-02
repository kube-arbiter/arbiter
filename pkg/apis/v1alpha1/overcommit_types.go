/*
Copyright 2022 The Arbiter Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RatioLimit struct {
	Ratio string `json:"ratio,omitempty"`
	Limit string `json:"limit,omitempty"`
}

// OverCommitSpec defines the desired state of OverCommit
type OverCommitSpec struct {
	NamespaceList []string   `json:"namespaceList,omitempty"`
	CPU           RatioLimit `json:"cpu,omitempty"`
	Memory        RatioLimit `json:"memory,omitempty"`
}

// OverCommitStatus defines the observed state of OverCommit
type OverCommitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// OverCommit is the Schema for the overcommits API
type OverCommit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OverCommitSpec   `json:"spec,omitempty"`
	Status OverCommitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OverCommitList contains a list of OverCommit
type OverCommitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OverCommit `json:"items"`
}
