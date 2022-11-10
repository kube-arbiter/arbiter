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

package provider

import (
	"context"
	"strconv"
	"sync"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/external_metrics"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"

	"github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
)

// arbiterMetricsProvider: external metrics provider for arbiter
type arbiterMetricsProvider struct {
	kubeClient *versioned.Clientset
	mapper     apimeta.RESTMapper

	valuesLock sync.RWMutex

	// ExternalMetricsProvider is a source of external metrics.
	// Metric is normally identified by a name and a set of labels/tags. It is up to a specific
	// implementation how to translate metricSelector to a filter for metric values.
	// Namespace can be used by the implemetation for metric identification, access control or ignored.
	provider.ExternalMetricsProvider
}

func NewArbiterProvider(mapper apimeta.RESTMapper, kubeClient *versioned.Clientset) provider.ExternalMetricsProvider {
	return &arbiterMetricsProvider{
		mapper:     mapper,
		kubeClient: kubeClient,
	}
}

// List metrics for a specified OBI within a namespace
func (p *arbiterMetricsProvider) GetExternalMetric(ctx context.Context, namespace string, metricSelector labels.Selector, info provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	matchingMetrics := []external_metrics.ExternalMetricValue{}
	// We'll use the name of OBI as the name of metric to match
	obiObject, err := p.kubeClient.ArbiterV1alpha1().ObservabilityIndicants(namespace).Get(ctx, info.Metric, metav1.GetOptions{})
	if err != nil {
		klog.Errorln("failed to get obi data", err)
		return &external_metrics.ExternalMetricValueList{
			Items: matchingMetrics,
		}, err
	}
	if obiObject.Spec.TargetRef.Kind != "Node" && obiObject.Spec.TargetRef.Kind != "Pod" {
		// For now, we only support to define external metrics for node and pod
		// TODO: make it more extendable
		return &external_metrics.ExternalMetricValueList{
			Items: matchingMetrics,
		}, nil
	}
	// If it's node or pod, get the gvr object
	// TODO: label selector will be matched later
	// WindowSeconds: indicates the window ([Timestamp-Window, Timestamp]) from which these metrics were calculated, when returning rate
	// metrics calculated from cumulative metrics (or zero for non-calculated instantaneous metrics).
	// So use 0 for instantaneous metric
	var metricIntervalWindow int64
	for metricKey := range obiObject.Spec.Metric.Metrics {
		if len(obiObject.Status.Metrics[metricKey]) > 0 {
			// For now, we only use 1st item that has the records, and the latest value of records
			recordLength := len(obiObject.Status.Metrics[metricKey][0].Records)
			if recordLength > 0 {
				latestRecord := obiObject.Status.Metrics[metricKey][0].Records[recordLength-1]
				emv := external_metrics.ExternalMetricValue{}
				emv.MetricName = info.Metric
				// Skip the empty value
				if latestRecord.Value != "" {
					// Get the integer value
					fValue, err := strconv.ParseFloat(latestRecord.Value, 64)
					if err != nil {
						klog.Errorf("error while parsing float value %s", latestRecord.Value)
						continue
					}
					stringValue := strconv.Itoa(int(fValue))
					emv.Value = resource.MustParse(stringValue + obiObject.Status.Metrics[metricKey][0].Unit)
				} else {
					klog.Warningf("found empty metrics value")
				}
				emv.Timestamp = metav1.Time{Time: time.UnixMilli(latestRecord.Timestamp)}
				emv.WindowSeconds = &metricIntervalWindow
				matchingMetrics = append(matchingMetrics, emv)
			}
		}
	}

	return &external_metrics.ExternalMetricValueList{
		Items: matchingMetrics,
	}, nil
}

// List metrics for all namespaces or specified namespace
func (p *arbiterMetricsProvider) ListAllExternalMetrics() []provider.ExternalMetricInfo {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	externalMetricsInfo := []provider.ExternalMetricInfo{}
	// We'll use the name of OBI as the name of metric to match
	obiList, err := p.kubeClient.ArbiterV1alpha1().ObservabilityIndicants("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorln("failed to list obi data", err)
		return externalMetricsInfo
	}
	for _, obiObject := range obiList.Items {
		if obiObject.Spec.TargetRef.Kind != "Node" && obiObject.Spec.TargetRef.Kind != "Pod" {
			// For now, we only support to define external metrics for node and pod
			// TODO: make it more extendable
			continue
		}
		// Just return the namespace and obi name for metrics reference
		externalMetricsInfo = append(externalMetricsInfo,
			provider.ExternalMetricInfo{
				Metric: obiObject.GetObjectMeta().GetName(),
			})
	}

	return externalMetricsInfo
}
