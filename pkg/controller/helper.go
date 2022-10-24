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
	"bytes"
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	"github.com/kube-arbiter/arbiter/pkg/utils"
)

func (c *controller) renderQuery(
	instance *v1alpha1.ObservabilityIndicant,
	unstructuredResource *unstructured.Unstructured) (map[string]string, error) {
	receiver := bytes.NewBuffer([]byte{})
	metricName2Query := make(map[string]string)
	for metricName, params := range instance.Spec.Metric.Metrics {
		if params.Query == "" {
			continue
		}
		klog.V(10).Infof("ObservabilityIndicant %s base query is: %s\n", instance.GetName(), params.Query)
		if err := utils.RenderQuery(params.Query, unstructuredResource.Object, receiver); err != nil {
			klog.Errorf("%s error rendering template: %s, won't start goroutine\n", utils.ErrorLogPrefix, err)
			c.recorder.Eventf(instance,
				corev1.EventTypeWarning,
				RenderTemplateEventReason,
				"error rendering template: %s, won't start goroutine", err)
			return nil, err
		}
		metricName2Query[metricName] = receiver.String()
		klog.V(10).Infof("ObservabilityIndicant %s redner query is: %s\n", instance.GetName(), metricName2Query[metricName])
		receiver.Reset()
	}
	return metricName2Query, nil
}

func (c *controller) getRuntimeObject(
	ctx context.Context,
	gvr schema.GroupVersionResource,
	instance *v1alpha1.ObservabilityIndicant) (*unstructured.Unstructured, error) {
	targetRef := instance.Spec.TargetRef

	klog.V(10).Infof("getRuntimeObject by %s\n", gvr.String())
	unstructuredResource := &unstructured.Unstructured{}
	var err error
	resourceInterface := c.dynamicClient.Resource(gvr).Namespace(targetRef.Namespace)
	if targetRef.Name != "" {
		unstructuredResource, err = resourceInterface.Get(ctx, targetRef.Name, meta.GetOptions{})
		if err != nil {
			klog.Errorf("%s get resource %s by name %s error: %s\n", utils.ErrorLogPrefix, gvr, targetRef.Name, err)
			return unstructuredResource, err
		}
	} else if len(targetRef.Labels) > 0 {
		labelSelector := labels.FormatLabels(targetRef.Labels)
		rs, err := resourceInterface.List(ctx, meta.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			klog.Errorf("%s list resources %s by label %s error: %s\n", utils.ErrorLogPrefix, gvr, labelSelector, err)
			return unstructuredResource, err
		}
		sort.Slice(rs.Items, func(i, j int) bool {
			return rs.Items[i].GetName() < rs.Items[j].GetName()
		})
		if targetRef.Index >= uint64(len(rs.Items)) {
			klog.Errorf("%s list resources %s by label %s IndexOutOfRange. index out of range [%d] with length %d\n",
				utils.ErrorLogPrefix, gvr, labelSelector, targetRef.Index, len(rs.Items))
			c.recorder.Event(instance, corev1.EventTypeWarning, "IndexOutOfRange",
				fmt.Sprintf("index out of range [%d] with length %d", targetRef.Index, len(rs.Items)))
			return nil, fmt.Errorf("index out of range [%d] with length %d", targetRef.Index, len(rs.Items))
		}
		unstructuredResource = &rs.Items[targetRef.Index]
	}
	return unstructuredResource, nil
}

func (c *controller) updateStatus(
	ctx context.Context,
	namespace, name string, extraOpts ...func(*v1alpha1.ObservabilityIndicant) error,
) error {
	instance, err := c.observabilityIndicantLister.ObservabilityIndicants(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("%s get instance by name %s in namespace: %serror: %s", utils.ErrorLogPrefix, name, namespace, err))
		return err
	}

	for _, opt := range extraOpts {
		if e := opt(instance); e != nil {
			utilruntime.HandleError(fmt.Errorf("%s update instance: %s error: %s", utils.ErrorLogPrefix, instance.GetName(), e))
			return e
		}
	}

	_, err = c.observabilityClientSet.ArbiterV1alpha1().ObservabilityIndicants(namespace).Update(ctx, instance, meta.UpdateOptions{})
	return err
}

// addMetric For updating instance status
func (c *controller) addMetric(metricName string, data *v1alpha1.ObservabilityIndicantStatusMetricInfo) func(indicant *v1alpha1.ObservabilityIndicant) error {
	return func(instance *v1alpha1.ObservabilityIndicant) error {
		if instance == nil {
			utilruntime.HandleError(fmt.Errorf("%s instance is nil", utils.ErrorLogPrefix))
			return fmt.Errorf("instance is nil")
		}

		metrics := instance.Status.Metrics
		if metrics == nil {
			metrics = make(map[string][]v1alpha1.ObservabilityIndicantStatusMetricInfo)
		}

		historyRecord := instance.Spec.Metric.HistoryLimit
		if _, ok := metrics[metricName]; !ok {
			metrics[metricName] = make([]v1alpha1.ObservabilityIndicantStatusMetricInfo, 0, historyRecord)
		}

		recordLen := len(data.Records)
		if recordLen == 0 {
			klog.Warningf("ObservabilityIndicant %s's metric data is empty, skip", instance.GetName())
			return nil
		}
		// only keep the last item
		data.Records = data.Records[recordLen-1:]

		targetLen := instance.Spec.Metric.TimeRangeSeconds / instance.Spec.Metric.MetricIntervalSeconds
		if len(metrics[metricName]) == 0 {
			metrics[metricName] = append(metrics[metricName], *data)
		} else if len(data.Records) > 0 {
			currentLen := int64(len(metrics[metricName][0].Records))
			if currentLen >= targetLen {
				diff := currentLen - targetLen + 1
				metrics[metricName][0].Records = metrics[metricName][0].Records[diff:]
			}
			metrics[metricName][0].Records = append(metrics[metricName][0].Records, data.Records...)
			metrics[metricName][0].StartTime = data.StartTime
			metrics[metricName][0].EndTime = data.EndTime
		}

		instance.Status.Metrics = metrics
		klog.V(10).Infof("update ObservabilityIndicant %s.%s metrics, the new metrics is: %#v\n", instance.GetName(), metricName, instance.Status.Metrics)
		return nil
	}
}

// setCondition For updating instance status
func (c *controller) setCondition(
	errMsg string,
	phase v1alpha1.ConditionType,
	reason v1alpha1.ConditionReason,
) func(indicant *v1alpha1.ObservabilityIndicant) error {
	return func(instance *v1alpha1.ObservabilityIndicant) error {
		if instance == nil {
			utilruntime.HandleError(fmt.Errorf("%s instance is nil", utils.ErrorLogPrefix))
			return fmt.Errorf("instance is nil")
		}

		cond := v1alpha1.Condition{
			Type:               phase,
			Reason:             reason,
			Message:            errMsg,
			LastHeartbeatTime:  meta.Now(),
			LastTransitionTime: meta.Now(),
		}

		if instance.Status.Conditions == nil {
			instance.Status.Conditions = make([]v1alpha1.Condition, 0)
		}
		if len(instance.Status.Conditions) == ConditionRecords {
			instance.Status.Conditions = instance.Status.Conditions[1:]
		}
		instance.Status.Conditions = append(instance.Status.Conditions, cond)
		return nil
	}
}
