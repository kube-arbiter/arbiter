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
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"

	gocache "github.com/patrickmn/go-cache"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	informerv1 "k8s.io/client-go/informers/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	schedv1alpha1 "github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	clientset "github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
)

const (
	DefaultName      = "default"
	ManagerLogPrefix = "[Arbiter-Manager] "
)

var (
	ErrNotFoundInCache = errors.New("not Found In Memory Cache")
	ErrTypeAssertion   = errors.New("type assertion err")
	ErrNoData          = errors.New("obi have no data")
)

type Manager1 interface {
	GetScoreSpec(ctx context.Context) string
	GetPodMetric(ctx context.Context, pod *v1.Pod) (metric MetricData, err error)
	GetNodeMetric(ctx context.Context, node *v1.Node) (metric MetricData, err error)
}

type Manager struct {
	client clientset.Interface

	// TODO(Abirdcfly): should benchmark gocache
	podMetric  *gocache.Cache
	nodeMetric *gocache.Cache
	scheduler  *gocache.Cache
	score      *gocache.Cache
	rsToDeploy *gocache.Cache

	// snapshotSharedLister is pod shared list
	snapshotSharedLister framework.SharedLister
	// podLister is pod lister
	podLister listerv1.PodLister

	sync.RWMutex
	nodeLister listerv1.NodeLister
}

func (mgr *Manager) GetPodMetric(ctx context.Context, pod *v1.Pod) (metric MetricData, err error) {
	// TODO(Abirdcfly): pod metric support will in v0.2.0
	return
}

func (mgr *Manager) GetNodeMetric(ctx context.Context, node *v1.Node) (metric MetricData, err error) {
	data, ok := mgr.nodeMetric.Get(metav1.NamespaceDefault + "/" + node.Name)
	if !ok {
		klog.V(4).ErrorS(ErrNotFoundInCache, "Failed to get nodeMetric", "nodeName", node.Name)
		return nil, ErrNotFoundInCache
	}
	metric, ok = data.(MetricData)
	if !ok {
		klog.V(4).ErrorS(ErrTypeAssertion, "Failed to get nodeMetric", "data", data)
		return nil, ErrTypeAssertion
	}
	if len(metric) == 0 {
		klog.V(4).ErrorS(ErrNoData, "Cache has key, but no data", "data", data)
		return nil, ErrNoData
	}
	return
}

func NewManager(client clientset.Interface, snapshotSharedLister framework.SharedLister, podInformer informerv1.PodInformer, nodeInformer informerv1.NodeInformer) *Manager {
	pgMgr := &Manager{
		client:               client,
		podMetric:            gocache.New(gocache.NoExpiration, gocache.NoExpiration),
		nodeMetric:           gocache.New(gocache.NoExpiration, gocache.NoExpiration),
		scheduler:            gocache.New(gocache.NoExpiration, gocache.NoExpiration),
		score:                gocache.New(gocache.NoExpiration, gocache.NoExpiration),
		rsToDeploy:           gocache.New(gocache.NoExpiration, gocache.NoExpiration),
		snapshotSharedLister: snapshotSharedLister,
		podLister:            podInformer.Lister(),
		nodeLister:           nodeInformer.Lister(),
		RWMutex:              sync.RWMutex{},
	}
	return pgMgr
}

func (mgr *Manager) GetScoreSpec(ctx context.Context) (logic string) {
	scoreName, ok := mgr.scheduler.Get("Score")
	if !ok {
		klog.ErrorS(ErrNotFoundInCache, "Failed to get scheduler.score")
		return
	}
	if scoreName == nil {
		return
	}
	scoreNameStr, ok := scoreName.(string)
	if !ok {
		klog.ErrorS(ErrTypeAssertion, "Failed to get scoreNameStr")
		return
	}
	scoreSpecLogic, ok := mgr.score.Get(scoreNameStr)
	if !ok {
		klog.ErrorS(ErrNotFoundInCache, "Failed to get score.spec.logic", "Score.Name", scoreNameStr)
		return
	}
	logic, ok = scoreSpecLogic.(string)
	if !ok {
		klog.ErrorS(ErrTypeAssertion, "Failed to get score.spec.logic", "Score.Name", scoreNameStr)
		return
	}
	return
}

func (mgr *Manager) ScoreAdd(obj interface{}) {
	klog.V(10).Infof("[%s] get new Score", ManagerLogPrefix)
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	ns, _, _ := cache.SplitMetaNamespaceKey(key)
	if ns != SchedulerNamespace() {
		return
	}
	score, ok := obj.(*schedv1alpha1.Score)
	if !ok {
		klog.ErrorS(ErrTypeAssertion, "Failed to get score")
		return
	}
	mgr.score.Set(score.Name, score.Spec.Logic, gocache.NoExpiration)
}

func (mgr *Manager) ScoreUpdate(old interface{}, new interface{}) {
	klog.V(10).Infof("[%s] get update Score", ManagerLogPrefix)
	mgr.ScoreAdd(new)
}

func (mgr *Manager) ScoreDelete(obj interface{}) {
	klog.V(10).Infof("[%s] get delete Score", ManagerLogPrefix)
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	ns, _, _ := cache.SplitMetaNamespaceKey(key)
	if ns != SchedulerNamespace() {
		return
	}
	score, ok := obj.(*schedv1alpha1.Score)
	if !ok {
		klog.ErrorS(ErrTypeAssertion, "Failed to get score")
		return
	}
	mgr.score.Delete(score.Name)
}

func (mgr *Manager) SchedulerAdd(obj interface{}) {
	klog.V(10).Infof("[%s] get new Scheduler", ManagerLogPrefix)
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	ns, name, _ := cache.SplitMetaNamespaceKey(key)
	if name != DefaultName || ns != SchedulerNamespace() {
		return
	}
	score, ok := obj.(*schedv1alpha1.Scheduler)
	if !ok {
		klog.ErrorS(ErrTypeAssertion, "Failed to get scheduler")
		return
	}
	mgr.scheduler.Set("Score", score.Spec.Score, gocache.NoExpiration)
}

func (mgr *Manager) SchedulerUpdate(old interface{}, new interface{}) {
	klog.V(10).Infoln(ManagerLogPrefix + "get update Scheduler")
	mgr.ScoreAdd(new)
}

func (mgr *Manager) SchedulerDelete(obj interface{}) {
	klog.V(10).Infoln(ManagerLogPrefix + "get delete Scheduler")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.ErrorS(err, ManagerLogPrefix+"Failed to obj in cache when delete", "obj", obj)
		runtime.HandleError(err)
		return
	}
	ns, name, _ := cache.SplitMetaNamespaceKey(key)
	if name != DefaultName || ns != SchedulerNamespace() {
		return
	}
	mgr.scheduler.Flush()
}

func (mgr *Manager) ObservabilityIndicantAdd(obj interface{}) {
	klog.V(10).Infoln(ManagerLogPrefix + "get new ObservabilityIndicant")
	_, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.V(5).ErrorS(err, ManagerLogPrefix+"Failed to obj in cache when add", "obj", obj)
		runtime.HandleError(err)
		return
	}
	obi, ok := obj.(*schedv1alpha1.ObservabilityIndicant)
	if !ok {
		klog.V(5).ErrorS(ErrTypeAssertion, ManagerLogPrefix+"Failed to get observability indicant", "obj", obj)
		return
	}
	expiration := gocache.NoExpiration
	if len(obi.Status.Metrics) == 0 {
		klog.V(5).ErrorS(ErrNoData, ManagerLogPrefix+"obi have no data", "obi", klog.KObj(obi))
		return
	}
	var cacheName *gocache.Cache
	switch {
	case IsResourceNode(obi.Spec.TargetRef):
		cacheName = mgr.nodeMetric
	case IsResourcePod(obi.Spec.TargetRef):
		cacheName = mgr.podMetric
	default:
		klog.V(5).ErrorS(ErrNotFoundInCache, ManagerLogPrefix+"Failed to get cacheName", "TargetRef", obi.Spec.TargetRef)
		return
	}
	cacheKey := getMetricCacheKey(obi)
	/*
		{
			"metrics": {
				"cpu": [
					{
						"endTime": "2022-09-01T10:36:57Z",
						"records": [
							{
								"timestamp": 1662024960000,
								"value": "0.470097"
							},
							{
								"timestamp": 1662025020000,
								"value": "0.466142"
							}
						],
						"startTime": "2022-09-01T09:36:57Z",
						"targetItem": "cpu-cost-1-7db7f54cdd-fnh8k",
						"unit": "C"
					}
				]
		    }
		}
	*/
	var metricData MetricData
	if data, ok := cacheName.Get(cacheKey); ok {
		metricData, ok = data.(MetricData)
		if !ok {
			klog.V(5).ErrorS(errors.New("get data err"), ManagerLogPrefix+"get data err")
			metricData = make(map[string]*FullMetrics)
		}
	} else {
		metricData = make(map[string]*FullMetrics)
	}
	metrics := obi.Status.Metrics
	hasRecords := false
	for metricType, metricInfo := range metrics {
		// metricType cpu mem ...
		if _, exist := (metricData)[metricType]; !exist {
			(metricData)[metricType] = new(FullMetrics)
		}
		v := (metricData)[metricType]
		if len(metricInfo) == 0 {
			continue
		}
		v.ObservabilityIndicantStatusMetricInfo = metricInfo[0].DeepCopy()
		if len(v.ObservabilityIndicantStatusMetricInfo.Records) == 0 {
			continue
		}
		hasRecords = true
		v.Max, v.Min, v.Avg = 0, 0, 0
		var i int
		var sum float64
		for _, r := range v.Records {
			val, err := strconv.ParseFloat(r.Value, 64)
			if err != nil {
				klog.V(5).ErrorS(err, ManagerLogPrefix+"Failed to parse float", "Value", r.Value, "obi", klog.KObj(obi))
				continue
			}
			if val > v.Max {
				v.Max = val
			}
			if val < v.Min {
				v.Min = val
			}
			sum += val
			i++
		}
		v.Avg = sum / float64(i)
	}
	if !hasRecords {
		klog.V(5).ErrorS(ErrNoData, ManagerLogPrefix+"has no records", "obi", klog.KObj(obi))
		return
	}
	cacheName.Set(cacheKey, metricData, expiration)
}

func (mgr *Manager) ObservabilityIndicantUpdate(old interface{}, new interface{}) {
	klog.V(10).Infoln(ManagerLogPrefix + "get update ObservabilityIndicant")
	mgr.ObservabilityIndicantAdd(new)
}

func (mgr *Manager) ObservabilityIndicantDelete(obj interface{}) {
	klog.V(10).Infoln(ManagerLogPrefix + "get delete ObservabilityIndicant")
	obi, ok := obj.(*schedv1alpha1.ObservabilityIndicant)
	if !ok {
		klog.ErrorS(errors.New("cant convert to observability indicant"), ManagerLogPrefix+"cant convert to observability indicant", "obj", obj)
		return
	}
	cacheKey := getMetricCacheKey(obi)
	// TODO(Abirdcfly): optimize deletion
	mgr.podMetric.Delete(cacheKey)
}

func (mgr *Manager) RsAdd(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	rs, ok := obj.(*appsv1.ReplicaSet)
	if !ok {
		klog.V(5).ErrorS(errors.New("cant convert to replica set"), "cant convert to replica set", "obj", obj)
		return
	}
	ownerReferences := rs.OwnerReferences
	if len(ownerReferences) == 0 {
		return
	}
	deployName := ""
	for _, ownerReference := range ownerReferences {
		if ownerReference.Kind == "Deployment" {
			deployName = ownerReference.Name
			break
		}
	}
	if deployName == "" {
		return
	}
	mgr.rsToDeploy.Set(key, deployName, gocache.NoExpiration)
}

func (mgr *Manager) RsUpdate(old interface{}, new interface{}) {
	mgr.RsAdd(new)
}

func (mgr *Manager) RsDelete(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	mgr.rsToDeploy.Delete(key)
}

func getMetricCacheKey(obi *schedv1alpha1.ObservabilityIndicant) string {
	ns := obi.Spec.TargetRef.Namespace
	if ns == "" {
		ns = metav1.NamespaceDefault
	}
	name := obi.Spec.TargetRef.Name
	for _, v := range obi.Status.Metrics {
		if len(v) == 0 {
			continue
		}
		name = v[0].TargetItem
	}
	return ns + "/" + name
}

func IsResourceNode(o schedv1alpha1.ObservabilityIndicantSpecTargetRef) bool {
	return o.Kind == "Node" && o.Group == v1.GroupName && o.Version == "v1"
}

func IsResourcePod(o schedv1alpha1.ObservabilityIndicantSpecTargetRef) bool {
	return o.Kind == "Pod" && o.Group == v1.GroupName && o.Version == "v1"
}

func SchedulerNamespace() string {
	// Assumes have set the POD_NAMESPACE environment variable using the downward API.
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns
	}
	// Fall back to the namespace associated with the service account token, if available
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return "kube-system"
}
