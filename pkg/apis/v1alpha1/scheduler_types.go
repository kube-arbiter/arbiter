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

// LogicPhase ...
type LogicPhase string

// These are the valid phase of LogicPhase.
const (
	// LogicPhasePending ...
	LogicPhasePending LogicPhase = "Pending"

	// LogicPhaseRunning means logic has been in running phase.
	LogicPhaseRunning LogicPhase = "Running"

	// LogicPhaseFailed means logic is failed.
	LogicPhaseFailed LogicPhase = "Failed"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Score ...
type Score struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the pod group.
	// +optional
	Spec ScoreSpec `json:"spec,omitempty"`

	// Status represents the current information about a pod group.
	// This data may not be up to date.
	// +optional
	Status ScoreStatus `json:"status,omitempty"`
}

// ScoreSpec ...
type ScoreSpec struct {
	// Weight of the Score
	Weight int64 `json:"weight,omitempty"`
	// Logic is the Javascript code
	Logic string `json:"logic"`
}

// ScoreStatus ...
type ScoreStatus struct {
	// Current phase of Logic
	LogicPhase LogicPhase `json:"phase,omitempty"`

	// ScheduleStartTime of the group
	ScheduleStartTime metav1.Time `json:"scheduleStartTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScoreList is a list of ElasticQuota items.
type ScoreList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of ElasticQuota objects.
	Items []Score `json:"items" protobuf:"bytes,2,rep,name=items"`
}
