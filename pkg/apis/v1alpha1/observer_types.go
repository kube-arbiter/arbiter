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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConditionReason string

const (
	NotFoundResource ConditionReason = "NotFoundResource"

	FetchDataError   ConditionReason = "FetchDataError"
	FetchDataSuccess ConditionReason = "FetchDataDone"
)

// ConditionType represent a resource's status
type ConditionType string

const (
	ConditionFetching ConditionType = "Fetching"

	ConditionFetchDataDone ConditionType = "FetchDataDone"

	// ConditionFailure represents Failure state of an object
	ConditionFailure ConditionType = "Failure"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:shortName={obi}
type ObservabilityIndicant struct {
	// +optional
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ObservabilityIndicantSpec `json:"spec,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	// +nullable
	Status ObservabilityIndicantStatus `json:"status,omitempty"`
}

type ObservabilityIndicantSpecTargetRef struct {
	Group   string `json:"group"`   // apps
	Version string `json:"version"` // v1
	Kind    string `json:"kind"`    // deployments

	// Under which namespace is the resource
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// When using label, there will be a set of resources,
	// these resources are sorted by name, and the target resource is selected through the index field.
	// +optional
	// +kubebuilder:default=0
	Index uint64 `json:"index,omitempty"`
}

type ObservabilityIndicantSpecMetricAbility struct {
	// allow user to define query sql
	// +optional
	Query string `json:"query,omitempty"`
	// +optional
	Description string `json:"description,omitempty"`

	Unit string `json:"unit"`
	// +optional
	Aggregations []string `json:"aggregations"`
}

type ObservabilityIndicantSpecMetric struct {
	Metrics          map[string]ObservabilityIndicantSpecMetricAbility `json:"metrics"`
	TimeRangeSeconds int64                                             `json:"timeRangeSeconds"`

	// +kubebuilder:validation:Minimum=5
	MetricIntervalSeconds int64 `json:"metricIntervalSeconds"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1
	// Currently this field can only be set to 1
	HistoryLimit int64 `json:"historyLimit"`
}

type ObservabilityIndicantSpecTrace struct {
}

type ObservabilityIndicantSpecLog struct {
}

type ObservabilityIndicantSpec struct {
	TargetRef ObservabilityIndicantSpecTargetRef `json:"targetRef"`
	Source    string                             `json:"source"`

	// +optional
	Metric *ObservabilityIndicantSpecMetric `json:"metric"`
	// +optional
	Trace *ObservabilityIndicantSpecTrace `json:"trace,omitempty"`
	// +optional
	Log *ObservabilityIndicantSpecLog `json:"log,omitempty"`
}

type ObservabilityIndicantStatusTimeInfo struct {
	// +optional
	StartTime metav1.Time `json:"startTime"`
	// +optional
	EndTime metav1.Time `json:"endTime"`
}

type ObservabilityIndicantStatusAggregationItem struct {
	// +optional
	Value string `json:"value"`
	// +optional
	TargetItem string `json:"targetItem"`
}

type Record struct {
	// +optional
	Timestamp int64 `json:"timestamp"`
	// +optional
	Value string `json:"value"`
}
type ObservabilityIndicantStatusMetricInfo struct {
	// +optional
	Unit string `json:"unit,omitempty"`
	// +optional
	TargetItem string `json:"targetItem,omitempty"`

	// +optional
	Records []Record `json:"records"`

	// +optional
	StartTime metav1.Time `json:"startTime"`
	// +optional
	EndTime metav1.Time `json:"endTime"`
}

type Condition struct {
	// +optional
	Type ConditionType `json:"type,omitempty"`
	// +optional
	Status v1.ConditionStatus `json:"status,omitempty"`
	// +optional
	Reason ConditionReason `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

type ObservabilityIndicantStatus struct {
	// +optional
	Phase ConditionType `json:"phase"`

	// log and trace retain
	// +optional
	Metrics map[string][]ObservabilityIndicantStatusMetricInfo `json:"metrics"`

	// +optional
	Conditions []Condition `json:"conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ObservabilityIndicantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ObservabilityIndicant `json:"items"`
}
