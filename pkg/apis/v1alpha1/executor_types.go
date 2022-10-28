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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// ActionCondition is the condition for the action
type ActionCondition struct {
	Expression  string `json:"expression,omitempty"`
	Operator    string `json:"operator,omitempty"`
	TargetValue string `json:"targetValue,omitempty"`
}

// ObservabilityActionPolicySpec defines the desired state of ObservabilityActionPolicy
type ObservabilityActionPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObIndicantName is the name of a ObservabilityIndicant.
	ObIndicantName string `json:"obIndicantname,omitempty"`

	Condition ActionCondition `json:"condition,omitempty"`

	// +optional
	Action string `json:"action,omitempty"`

	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	ActionData *runtime.RawExtension `json:"actionData,omitempty"`
}

// ActionInfo is the information to identify whether a resource is updated
type ActionInfo struct {
	// ResourceName is the name of the pod
	ResourceName string `json:"resourceName,omitempty"`

	ExpressionValue string `json:"expressionValue,omitempty"`

	ConditionValue string `json:"conditionValue,omitempty"`
}

// ObservabilityTimeInfo defines the observed time range of ObservabilityIndicant
type ObservabilityTimeInfo struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	StartTime       metav1.Time `json:"startTime,omitempty"`
	EndTime         metav1.Time `json:"endTime,omitempty"`
	IntervalSeconds int64       `json:"metricIntervalSeconds,omitempty"`
}

// ObservabilityActionPolicyStatus defines the observed state of ObservabilityActionPolicy
type ObservabilityActionPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Names of the updated resources
	ActionInfo []ActionInfo `json:"actionInfo,omitempty"`

	TimeInfo ObservabilityTimeInfo `json:"timeInfo,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:shortName={oap}

// ObservabilityActionPolicy is the Schema for the ObservabilityActionpolicies API
type ObservabilityActionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObservabilityActionPolicySpec   `json:"spec,omitempty"`
	Status ObservabilityActionPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ObservabilityActionPolicyList contains a list of ObservabilityActionPolicy
type ObservabilityActionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObservabilityActionPolicy `json:"items"`
}
