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

package controller

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
)

type KeyFunc func(string, *v1alpha1.ObservabilityIndicant) error

func TestControllerKeyFunc(t *testing.T) {
	baseCr := &v1alpha1.ObservabilityIndicant{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-new-cr",
			Namespace: "arbiter",
		},
		Spec: v1alpha1.ObservabilityIndicantSpec{
			TargetRef: v1alpha1.ObservabilityIndicantSpecTargetRef{
				Group:     "",
				Version:   "v1",
				Kind:      "Pod",
				Name:      "pod1",
				Namespace: "default",
			},
			Source: "metric-server",
			Metric: &v1alpha1.ObservabilityIndicantSpecMetric{
				Metrics: map[string]v1alpha1.ObservabilityIndicantSpecMetricAbility{
					"cpu": {
						Query:       "",
						Description: "test cpu query",
						Unit:        "m",
						Aggregations: []string{
							"time",
						},
					},
				},
				TimeRangeSeconds:      12,
				HistoryLimit:          1, // max and min both are one.
				MetricIntervalSeconds: 3,
			},
		},
	}

	testKeyFuncs := map[string]KeyFunc{
		"testAddMetric":  testAddMetric,
		"testLoopUpdate": testLoopUpdate,
		"testOneRecord":  testOneRecord,
	}
	for funcName, fn := range testKeyFuncs {
		if err := fn(funcName, baseCr.DeepCopy()); err != nil {
			t.Fatalf("%s %s", funcName, err)
		}
	}
}

func testAddMetric(name string, instance *v1alpha1.ObservabilityIndicant) error {
	log.Printf("%s start to test...\n", name)
	now := time.Now()
	loopMetrics := []*v1alpha1.ObservabilityIndicantStatusMetricInfo{
		{
			Unit:       "m",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.UnixMilli(),
					Value:     "0.1",
				},
			},
		},
		{
			Unit:       "m",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.Add(6 * time.Second).UnixMilli(),
					Value:     "0.4",
				},
			},
		},
		{
			Unit:       "m",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.Add(12 * time.Second).UnixMilli(),
					Value:     "0.7",
				},
			},
		},
	}
	expectStatus := map[string][]v1alpha1.ObservabilityIndicantStatusMetricInfo{
		"cpu": {
			{
				Unit:       "m",
				TargetItem: "pod1",
				Records: []v1alpha1.Record{
					{
						Timestamp: now.UnixMilli(),
						Value:     "0.1",
					},
					{
						Timestamp: now.Add(6 * time.Second).UnixMilli(),
						Value:     "0.4",
					},
					{
						Timestamp: now.Add(12 * time.Second).UnixMilli(),
						Value:     "0.7",
					},
				},
			},
		},
	}

	for _, metricResult := range loopMetrics {
		new(controller).addMetric("cpu", metricResult)(instance)
	}

	if !reflect.DeepEqual(expectStatus, instance.Status.Metrics) {
		return fmt.Errorf("expect %v get %v", expectStatus, instance.Status.Metrics)
	}
	return nil
}

func testLoopUpdate(name string, instance *v1alpha1.ObservabilityIndicant) error {
	log.Printf("%s start to test...\n", name)
	now := time.Now()
	loopMetrics := []*v1alpha1.ObservabilityIndicantStatusMetricInfo{
		{
			Unit:       "C",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.UnixMilli(),
					Value:     "0.1",
				},
				{
					Timestamp: now.Add(2 * time.Second).UnixMilli(),
					Value:     "0.2",
				},
				{
					Timestamp: now.Add(4 * time.Second).UnixMilli(),
					Value:     "0.3",
				},
			},
		},
		{
			Unit:       "C",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.Add(6 * time.Second).UnixMilli(),
					Value:     "0.4",
				},
				{
					Timestamp: now.Add(8 * time.Second).UnixMilli(),
					Value:     "0.5",
				},
				{
					Timestamp: now.Add(10 * time.Second).UnixMilli(),
					Value:     "0.6",
				},
			},
		},
		{
			Unit:       "C",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.Add(12 * time.Second).UnixMilli(),
					Value:     "0.7",
				},
				{
					Timestamp: now.Add(14 * time.Second).UnixMilli(),
					Value:     "0.8",
				},
				{
					Timestamp: now.Add(16 * time.Second).UnixMilli(),
					Value:     "0.9",
				},
			},
		},
		{
			Unit:       "C",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.Add(18 * time.Second).UnixMilli(),
					Value:     "0.11",
				},
				{
					Timestamp: now.Add(20 * time.Second).UnixMilli(),
					Value:     "0.12",
				},
				{
					Timestamp: now.Add(22 * time.Second).UnixMilli(),
					Value:     "0.13",
				},
			},
		},
	}
	expectStatus := map[string][]v1alpha1.ObservabilityIndicantStatusMetricInfo{
		"cpu": {
			{
				Unit:       "C",
				TargetItem: "pod1",
				Records: []v1alpha1.Record{
					{
						Timestamp: now.Add(4 * time.Second).UnixMilli(),
						Value:     "0.3",
					},
					{
						Timestamp: now.Add(10 * time.Second).UnixMilli(),
						Value:     "0.6",
					},
					{
						Timestamp: now.Add(16 * time.Second).UnixMilli(),
						Value:     "0.9",
					},
					{
						Timestamp: now.Add(22 * time.Second).UnixMilli(),
						Value:     "0.13",
					},
				},
			},
		},
	}

	for _, metricResult := range loopMetrics {
		new(controller).addMetric("cpu", metricResult)(instance)
	}

	if !reflect.DeepEqual(expectStatus, instance.Status.Metrics) {
		return fmt.Errorf("expect %v get %v", expectStatus, instance.Status.Metrics)
	}

	return nil
}
func testOneRecord(name string, instance *v1alpha1.ObservabilityIndicant) error {
	log.Printf("%s start to test...\n", name)
	// NOTE: please deepcopy
	instance.Spec.Metric.TimeRangeSeconds = 1
	instance.Spec.Metric.MetricIntervalSeconds = 1

	now := time.Now()
	loopMetrics := []*v1alpha1.ObservabilityIndicantStatusMetricInfo{
		{
			Unit:       "C",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.UnixMilli(),
					Value:     "0.1",
				},
				{
					Timestamp: now.Add(2 * time.Second).UnixMilli(),
					Value:     "0.2",
				},
				{
					Timestamp: now.Add(4 * time.Second).UnixMilli(),
					Value:     "0.3",
				},
			},
		},
		{
			Unit:       "C",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.Add(6 * time.Second).UnixMilli(),
					Value:     "0.4",
				},
				{
					Timestamp: now.Add(8 * time.Second).UnixMilli(),
					Value:     "0.5",
				},
				{
					Timestamp: now.Add(10 * time.Second).UnixMilli(),
					Value:     "0.6",
				},
			},
		},
		{
			Unit:       "C",
			TargetItem: "pod1",
			Records: []v1alpha1.Record{
				{
					Timestamp: now.Add(12 * time.Second).UnixMilli(),
					Value:     "0.7",
				},
				{
					Timestamp: now.Add(14 * time.Second).UnixMilli(),
					Value:     "0.8",
				},
				{
					Timestamp: now.Add(16 * time.Second).UnixMilli(),
					Value:     "0.9",
				},
			},
		},
	}
	expectStatus := map[string][]v1alpha1.ObservabilityIndicantStatusMetricInfo{
		"cpu": {
			{
				Unit:       "C",
				TargetItem: "pod1",
				Records: []v1alpha1.Record{
					{
						Timestamp: now.Add(4 * time.Second).UnixMilli(),
						Value:     "0.3",
					},
				},
			},
			{
				Unit:       "C",
				TargetItem: "pod1",
				Records: []v1alpha1.Record{
					{
						Timestamp: now.Add(10 * time.Second).UnixMilli(),
						Value:     "0.6",
					},
				},
			},
			{
				Unit:       "C",
				TargetItem: "pod1",
				Records: []v1alpha1.Record{
					{
						Timestamp: now.Add(16 * time.Second).UnixMilli(),
						Value:     "0.9",
					},
				},
			},
		},
	}

	for idx, metricResult := range loopMetrics {
		new(controller).addMetric("cpu", metricResult)(instance)

		if r := len(instance.Status.Metrics["cpu"]); r != 1 {
			return fmt.Errorf("expect len 1 get %d", r)
		}
		if !reflect.DeepEqual(expectStatus["cpu"][idx], instance.Status.Metrics["cpu"][0]) {
			return fmt.Errorf("expect %v get %v", expectStatus["cpu"][idx], instance.Status.Metrics["cpu"][0])
		}
	}

	return nil
}
