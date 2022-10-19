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
	"reflect"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	"github.com/kube-arbiter/arbiter/pkg/fetch"
	clientset "github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
	obiScheme "github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned/scheme"
	obiInformer "github.com/kube-arbiter/arbiter/pkg/generated/informers/externalversions/apis/v1alpha1"
	obiLister "github.com/kube-arbiter/arbiter/pkg/generated/listers/apis/v1alpha1"
	obi "github.com/kube-arbiter/arbiter/pkg/proto/lib/observer"
	"github.com/kube-arbiter/arbiter/pkg/utils"
)

const (
	controllerName = "observabilityindicant-controller"
	// finalizer cancel goroutine
	finalizer = "obi.finalizers.k8s.com.cn"

	ConditionRecords = 3

	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "ObservabilityIndicant synced successfully"

	// RenderTemplateEventReason means there is an error in rendering the template via runtimeObject
	RenderTemplateEventReason = "RenderTemplateError"

	// FetchDataErrorEventReason  means that some data for the resource is fetched incorrectly
	FetchDataErrorEventReason = "FetchDataError"

	// FindResourceErrorEventReason  means that can not find resources by kind.
	FindResourceErrorEventReason = "FailedFindingResource"
)

type cancelGoroutine struct {
	ctx        context.Context //nolint
	cancelFunc context.CancelFunc
}

type controller struct {
	watchNamespaces map[string]struct{}

	kubeClientSet          kubernetes.Interface
	dynamicClient          dynamic.Interface
	observabilityClientSet clientset.Interface

	podLister corelister.PodLister
	podSynced cache.InformerSynced

	observabilityIndicantLister obiLister.ObservabilityIndicantLister
	observabilityIndicantSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder

	// Every observability CR should have a context.
	// When the CR is deleted, the relevant processes should be stopped in time.
	ctxs map[string]cancelGoroutine

	baseCtx context.Context //nolint
	fetcher fetch.Fetcher
}

func NewController(
	namespace string,
	kubeClientSet kubernetes.Interface,
	dynamicClient dynamic.Interface,
	observabilityClientSet clientset.Interface,
	podInformer coreinformers.PodInformer,
	observabilityInformer obiInformer.ObservabilityIndicantInformer,
	fetcher fetch.Fetcher,
) *controller {
	utilruntime.Must(obiScheme.AddToScheme(scheme.Scheme))
	klog.Info("Creating event broadcaster...")

	c := &controller{
		watchNamespaces: make(map[string]struct{}),
		kubeClientSet:   kubeClientSet,

		dynamicClient:               dynamicClient,
		observabilityClientSet:      observabilityClientSet,
		podLister:                   podInformer.Lister(),
		podSynced:                   podInformer.Informer().HasSynced,
		observabilityIndicantLister: observabilityInformer.Lister(),
		observabilityIndicantSynced: observabilityInformer.Informer().HasSynced,
		workqueue:                   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "obi"),
		ctxs:                        make(map[string]cancelGoroutine),
		baseCtx:                     context.Background(),
		fetcher:                     fetcher,
	}
	if len(namespace) != 0 {
		for _, n := range strings.Split(namespace, ",") {
			c.watchNamespaces[n] = struct{}{}
		}
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	c.recorder = eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})

	observabilityInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDel,
	})

	return c
}

func (c *controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Infoln("Starting observability indicant controller")
	klog.Infoln("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(stopCh, c.podSynced, c.observabilityIndicantSynced); !ok {
		return fmt.Errorf("%s failed to wait for caches to sync", utils.ErrorLogPrefix)
	}

	klog.Infoln("Starting worker")
	go wait.Until(c.runWorker, time.Second, stopCh)

	klog.Infoln("Started worker")
	<-stopCh
	klog.Infoln("Shutting down worker")

	return nil
}

func (c *controller) onAdd(obj interface{}) {
	var (
		key string
		err error
	)
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.workqueue.Add(key)
}

func (c *controller) onUpdate(old, new interface{}) {
	oldObi, ok := old.(*v1alpha1.ObservabilityIndicant)
	if !ok {
		klog.Errorf("Update %s/%s error: old is not ObservabilityIndicant\n",
			oldObi.GetNamespace(), oldObi.GetName())
		return
	}
	newObi, ok := new.(*v1alpha1.ObservabilityIndicant)
	if !ok {
		klog.Errorf("Update %s/%s error: new is not ObservabilityIndicant\n",
			oldObi.GetNamespace(), oldObi.GetName())
		return
	}
	if !reflect.DeepEqual(oldObi.Spec.Metric, newObi.Spec.Metric) ||
		len(oldObi.Finalizers) != len(newObi.Finalizers) ||
		newObi.DeletionTimestamp != nil ||
		!reflect.DeepEqual(oldObi.Spec.TargetRef, newObi.Spec.TargetRef) {
		klog.Infof("Update %s/%s\n", oldObi.GetNamespace(), oldObi.GetName())
		c.onAdd(new)
		return
	}
}

func (c *controller) onDel(obj interface{}) {
	o, ok := obj.(*v1alpha1.ObservabilityIndicant)
	if !ok {
		klog.Errorf("Delete %s/%s error: obj is not ObservabilityIndicant\n", o.GetNamespace(), o.GetName())
		return
	}
	klog.Infof("Delete %s/%s \n", o.GetNamespace(), o.GetName())
	c.onAdd(obj)
}

func (c *controller) runWorker() {
	for c.processNextWorkItem() {

	}
}

func (c *controller) processNextWorkItem() bool {
	obi, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	if err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var (
			key string
			ok  bool
		)

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("%s expected string in workqueue but got %#v", utils.ErrorLogPrefix, obi))
			return nil
		}
		if err := c.syncHandler(key, true); err != nil {
			if utils.IsAddFinalizerError(err) {
				c.workqueue.Add(key)
				return nil
			}

			utilruntime.HandleError(fmt.Errorf("%s syncHandler error: %s", utils.ErrorLogPrefix, err))
			c.workqueue.AddRateLimited(key)
			return err
		}

		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'\n", key)
		return nil
	}(obi); err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (c *controller) syncHandler(key string, startFetchRoutine bool) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(err)
		return err
	}
	obi, err := c.observabilityIndicantLister.ObservabilityIndicants(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("%s observability indicant '%s' in workqueue no longer exists",
				utils.ErrorLogPrefix, key))
			return nil
		}
		utilruntime.HandleError(fmt.Errorf("%s get observability indicant '%s' error: %s", utils.AddFinalizerErrorMsg, key, err))
		return err
	}

	_, ok := c.watchNamespaces[obi.GetNamespace()]
	ok = ok || len(c.watchNamespaces) == 0 // watchNamespace=0 means all namespace.

	klog.V(10).Infof("startFetchRoutine: %v, namespace in watch scope: %v\n", startFetchRoutine, ok)
	if startFetchRoutine && !(ok && obi.Spec.Source == c.fetcher.GetPluginName()) {
		klog.Infof("SyncHandler: expect sourceName: %s get %s or namespace %s is not in watch scope",
			c.fetcher.GetPluginName(), obi.Spec.Source, obi.GetNamespace())
		return nil
	}

	if obi.DeletionTimestamp.IsZero() {
		if !utils.ContainFinalizers(obi.Finalizers, finalizer) {
			o := obi.DeepCopy()
			o.Finalizers = utils.AddFinalizers(o.Finalizers, finalizer)
			klog.V(10).Infof("Add finalizers for %s, finalizers: %v\n", key, o.Finalizers)
			o.Status.Phase = v1alpha1.ConditionFetching
			if _, err = c.observabilityClientSet.ArbiterV1alpha1().ObservabilityIndicants(namespace).
				Update(context.Background(), o, meta.UpdateOptions{}); err != nil {
				return err
			}
			return fmt.Errorf(utils.AddFinalizerErrorMsg)
		}

		klog.V(10).Infof("the finalizer already exists, execute the start logic")
		if routineCtx, ok := c.ctxs[key]; ok {
			routineCtx.cancelFunc()
			delete(c.ctxs, key)
		}
		ctx, cancel := context.WithCancel(c.baseCtx)
		c.ctxs[key] = cancelGoroutine{ctx: ctx, cancelFunc: cancel}

		if startFetchRoutine {
			klog.Infof("Start routine to fetch data for %s...\n", obi.GetName())
			go c.fetchRouteV1Alpha1(ctx, obi.DeepCopy())
		}

		c.recorder.Event(obi, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
		return nil
	}

	klog.Infof("SyncHandler: delete finalizers and cancel goroutine for %s\n", key)
	if utils.ContainFinalizers(obi.Finalizers, finalizer) {
		o := obi.DeepCopy()
		o.Finalizers = utils.RemoveFinalizer(o.Finalizers, finalizer)
		klog.V(10).Infof("Remove finalizers for %s, finalizers: %v\n", key, o.Finalizers)
		o.Status.Phase = v1alpha1.ConditionFetchDataDone
		if _, err = c.observabilityClientSet.ArbiterV1alpha1().ObservabilityIndicants(namespace).
			Update(context.Background(), o, meta.UpdateOptions{}); err != nil {
			return err
		}
	}

	if cancelRoutine, ok := c.ctxs[key]; ok {
		klog.Infof("Stop %s fetch routine", key)
		cancelRoutine.cancelFunc()
		delete(c.ctxs, key)
	}
	return nil
}

func (c *controller) getRuntimeObject(
	gvr schema.GroupVersionResource,
	instance *v1alpha1.ObservabilityIndicant) (*unstructured.Unstructured, error) {
	targetRef := instance.Spec.TargetRef

	klog.V(10).Infof("getRuntimeObject by %s\n", gvr.String())
	unstructuredResource := &unstructured.Unstructured{}
	var err error
	resourceInterface := c.dynamicClient.Resource(gvr).Namespace(targetRef.Namespace)
	if targetRef.Name != "" {
		unstructuredResource, err = resourceInterface.Get(context.TODO(), targetRef.Name, meta.GetOptions{})
		if err != nil {
			klog.Errorf("%s get resource %s by name %s error: %s\n", utils.ErrorLogPrefix, gvr, targetRef.Name, err)
			return unstructuredResource, err
		}
	} else if len(targetRef.Labels) > 0 {
		labelSelector := labels.FormatLabels(targetRef.Labels)
		rs, err := resourceInterface.List(context.TODO(), meta.ListOptions{
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

func (c *controller) fetchRouteV1Alpha1(ctx context.Context, instance *v1alpha1.ObservabilityIndicant) {
	method := "controller.fetchRoute"
	// TODO(0xff-dev): now only support metric.
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	targetRef := instance.Spec.TargetRef
	resourcePlural, err := utils.ResourcePlural(c.kubeClientSet, targetRef.Group, targetRef.Version, targetRef.Kind)
	if err != nil {
		klog.Errorf("%s convert kind to resource error: %s\n", utils.ErrorLogPrefix, err)
		c.recorder.Eventf(instance, corev1.EventTypeWarning, FindResourceErrorEventReason, "convert kind to resource error: %s", err)
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    targetRef.Group,
		Version:  targetRef.Version,
		Resource: resourcePlural,
	}

	klog.V(10).Infof("fetchRoute targetRef name: %s, targetRef labels: %v\n", targetRef.Name, targetRef.Labels)

	do1 := func() func() {
		var (
			unstructuredResource *unstructured.Unstructured
			metricName2Query     map[string]string
			firstCall            = true
		)
		return func() {
			tmpFirstCall := firstCall
			firstCall = false
			if tmpFirstCall || (targetRef.Name == "" && len(targetRef.Labels) > 0) {
				unstructuredResource, err = c.getRuntimeObject(gvr, instance)
				if err != nil {
					return
				}
				metricName2Query, err = c.renderQuery(instance, unstructuredResource.DeepCopy())
				if err != nil {
					klog.Errorf("%s Do renderQuery error: %s\n", utils.ErrorLogPrefix, err)
					return
				}
			}
			if unstructuredResource == nil {
				klog.Errorf("%s runtime object is nil", utils.ErrorLogPrefix)
				return
			}
			allMetricSuccess := make([]func(*v1alpha1.ObservabilityIndicant) error, 0)
			for metricName, args := range instance.Spec.Metric.Metrics {
				now := time.Now()

				//duration := time.Duration(instance.Spec.Metric.TimeRangeSeconds) * time.Second
				duration := time.Duration(instance.Spec.Metric.MetricIntervalSeconds) * time.Second
				obiReq := obi.GetMetricsRequest{
					ResourceNames: []string{unstructuredResource.GetName()},
					StartTime:     now.Add(-duration).UnixMilli(),
					EndTime:       now.UnixMilli(),
					Unit:          args.Unit,
					Kind:          unstructuredResource.GetKind(),
					Namespace:     targetRef.Namespace,
					MetricName:    metricName,
					Aggregation:   args.Aggregations,
				}

				if args.Query != "" {
					klog.V(5).Infof("this is a query based request")
					obiReq.Query = metricName2Query[metricName]
				}

				klog.V(10).Infof("fetch body: %s\n", obiReq.String())

				result, errType, err := c.fetcher.Fetch(subCtx, &obiReq)
				if err != nil {
					// NOTE: In the case of context interruption, the error should not be reflected in the conditions.
					if errType != fetch.FetchCtxDone {
						utilruntime.HandleError(err)
						c.recorder.Eventf(instance, corev1.EventTypeWarning,
							FetchDataErrorEventReason,
							"instance %s get metric data error: %s", instance.GetName(), err)
						_ = c.updateStatus(instance.GetNamespace(), instance.GetName(), c.setCondition(err.Error(),
							v1alpha1.ConditionFailure, v1alpha1.FetchDataError))
					}
					break
				}
				allMetricSuccess = append(allMetricSuccess, c.addMetric(metricName, result))
			}

			klog.V(10).Infof("allMetricSuccess match instance Metrics: %v\n", len(allMetricSuccess) == len(instance.Spec.Metric.Metrics))

			if len(allMetricSuccess) == len(instance.Spec.Metric.Metrics) {
				klog.Infof("%s Update instance %s status\n", method, instance.Name)
				allMetricSuccess = append(allMetricSuccess, c.setCondition("",
					v1alpha1.ConditionFetchDataDone, v1alpha1.FetchDataSuccess))
				if err := c.updateStatus(instance.GetNamespace(), instance.GetName(), allMetricSuccess...); err != nil {
					utilruntime.HandleError(fmt.Errorf("%s update instance %s status error: %s", utils.ErrorLogPrefix, instance.GetName(), err))
				}
				return
			}
		}
	}

	do := do1()
	do()

	for {
		select {
		case <-ctx.Done():
			klog.Infof("%s context deadline", method)
			return
		case <-time.After(time.Second * time.Duration(instance.Spec.Metric.MetricIntervalSeconds)):
			do()
		}
	}
}

func (c *controller) updateStatus(
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

	_, err = c.observabilityClientSet.ArbiterV1alpha1().ObservabilityIndicants(namespace).Update(context.TODO(), instance, meta.UpdateOptions{})
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
			metrics[metricName][0].Records = append(metrics[metricName][0].Records, (*data).Records...)
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
