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

package manager

import (
	v1 "k8s.io/api/core/v1"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
)

type OBI struct {
	Metric map[string]FullMetrics `json:"metric"` // Metric is a map, key is metric type
}

type PodWithOBI struct {
	Pod v1.Pod         `json:"raw"`
	OBI map[string]OBI `json:"obi"` // OBI is a map, key is obi name
}

type NodeWithOBI struct {
	Node   v1.Node        `json:"raw"`
	CPUReq int64          `json:"cpuReq"`
	MemReq int64          `json:"memReq"`
	OBI    map[string]OBI `json:"obi"` // OBI is a map, key is obi name
}

type FullMetrics struct {
	v1alpha1.ObservabilityIndicantStatusMetricInfo
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
	Min float64 `json:"min"`
}
