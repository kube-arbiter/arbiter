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

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Scheduler ...
type Scheduler struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// ElasticQuotaSpec defines the Min and Max for Quota.
	// +optional
	Spec SchedulerSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// ElasticQuotaStatus defines the observed use.
	// +optional
	Status SchedulerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// SchedulerSpec ...
type SchedulerSpec struct {
	// +optional
	Score string `json:"score,omitempty" protobuf:"bytes,1,opt,name=score"`
	// todo other type
}

// SchedulerStatus ...
type SchedulerStatus struct {
	// todo
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SchedulerList ...
type SchedulerList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is a list of ElasticQuota objects.
	Items []Scheduler `json:"items" protobuf:"bytes,2,rep,name=items"`
}

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
	// todo
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
