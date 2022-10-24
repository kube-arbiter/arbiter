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
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	if obi.Spec.Source != c.fetcher.GetPluginName() {
		// just return if the OBI source doesn't match with the plugin
		return nil
	}

	if len(c.watchNamespaces) == 0 {
		// watch all namespaces
		klog.Infof("SyncHandler: watching all namespaces")
	} else {
		_, ok := c.watchNamespaces[obi.GetNamespace()]
		if !ok {
			// just return if the watched namespace and OBI defined namespace isn't matched
			return nil
		}
	}
	klog.Infof("SyncHandler: handling '%s'\n", key)

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

func (c *controller) fetchRouteV1Alpha1(ctx context.Context, instance *v1alpha1.ObservabilityIndicant) {
	method := "controller.fetchRoute"
	// TODO(0xff-dev): now only support metric.
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var gvr schema.GroupVersionResource
	targetRef := instance.Spec.TargetRef

	if targetRef.Kind == "Node" || targetRef.Kind == "Pod" {
		// If it's node or pod, get the gvr object
		resourcePlural, err := utils.ResourcePlural(c.kubeClientSet, targetRef.Group, targetRef.Version, targetRef.Kind)
		if err != nil {
			klog.Errorf("%s convert kind to resource error: %s\n", utils.ErrorLogPrefix, err)
			c.recorder.Eventf(instance, corev1.EventTypeWarning, FindResourceErrorEventReason, "convert kind to resource error: %s", err)
			return
		}
		gvr = schema.GroupVersionResource{
			Group:    targetRef.Group,
			Version:  targetRef.Version,
			Resource: resourcePlural,
		}
		klog.V(10).Infof("fetchRoute targetRef name: %s, targetRef labels: %v\n", targetRef.Name, targetRef.Labels)
	}

	var (
		unstructuredResource *unstructured.Unstructured
		metricName2Query     map[string]string
		firstCall            = true
	)
	fetchObiData := func() {
		tmpFirstCall := firstCall
		firstCall = false
		// Get the runtime resource for 1st time if the OBI is based on label match
		if tmpFirstCall || (targetRef.Name == "" && len(targetRef.Labels) > 0) {
			var err error
			if !gvr.Empty() {
				unstructuredResource, err = c.getRuntimeObject(subCtx, gvr, instance)
				if err != nil {
					return
				}
			}
			if unstructuredResource != nil {
				metricName2Query, err = c.renderQuery(instance, unstructuredResource.DeepCopy())
			} else {
				metricName2Query, err = c.renderQuery(instance, &unstructured.Unstructured{Object: map[string]interface{}{}})
			}
			if err != nil {
				klog.Errorf("%s Do renderQuery error: %s\n", utils.ErrorLogPrefix, err)
				return
			}
		}
		var resourceNames []string
		var kind string
		if unstructuredResource != nil {
			resourceNames = []string{unstructuredResource.GetName()}
			kind = unstructuredResource.GetKind()
		}
		allMetricSuccess := make([]func(*v1alpha1.ObservabilityIndicant) error, 0)
		for metricName, args := range instance.Spec.Metric.Metrics {
			now := time.Now()

			duration := time.Duration(instance.Spec.Metric.MetricIntervalSeconds) * time.Second
			obiReq := obi.GetMetricsRequest{
				ResourceNames: resourceNames,
				StartTime:     now.Add(-duration).UnixMilli(),
				EndTime:       now.UnixMilli(),
				Unit:          args.Unit,
				Kind:          kind,
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
					_ = c.updateStatus(subCtx, instance.GetNamespace(), instance.GetName(), c.setCondition(err.Error(),
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
			if err := c.updateStatus(subCtx, instance.GetNamespace(), instance.GetName(), allMetricSuccess...); err != nil {
				utilruntime.HandleError(fmt.Errorf("%s update instance %s status error: %s", utils.ErrorLogPrefix, instance.GetName(), err))
			}
			return
		}
	}

	// Fetch for 1st time
	fetchObiData()

	for {
		select {
		case <-ctx.Done():
			klog.Infof("%s context deadline", method)
			return
		case <-time.After(time.Second * time.Duration(instance.Spec.Metric.MetricIntervalSeconds)):
			fetchObiData()
		}
	}
}
