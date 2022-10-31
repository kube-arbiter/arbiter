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

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/helper"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	"github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
	informers "github.com/kube-arbiter/arbiter/pkg/generated/informers/externalversions"
	"github.com/kube-arbiter/arbiter/pkg/scheduler/manager"
)

const (
	Name           = "Arbiter"
	LogPrefix      = "[arbiter] "
	DebugLogic     = `console.log("[arbiter]", "pod:", JSON.stringify(pod), "node:", JSON.stringify(node));`
	PredictedRatio = 0.6
	PreLogic       = `var node = JSON.parse(JSON.stringify(node));
node.raw.status.capacity.cpus = nodeCpus;
node.raw.status.capacity.memory = nodeMemory;

`
)

var (
	ErrNoScoreFunction = errors.New("no score function found")
)

type Arbiter struct {
	frameworkHandler framework.Handle
	manager          manager.Manager1
}

var _ framework.PostBindPlugin = &Arbiter{}
var _ framework.ScorePlugin = &Arbiter{}
var _ framework.ScoreExtensions = &Arbiter{}

func (ex *Arbiter) NormalizeScore(ctx context.Context, state *framework.CycleState, p *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	klog.V(10).Infoln(LogPrefix + "NormalizeScore")
	return helper.DefaultNormalizeScore(framework.MaxNodeScore, false, scores)
}

func (ex *Arbiter) backToDefaultScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (score int64, newState *framework.Status) {
	klog.V(2).Infoln(LogPrefix + "back to default score")
	return 0, nil
}

func (ex *Arbiter) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (score int64, newState *framework.Status) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				klog.ErrorS(err, LogPrefix+"catch a panic", "pod", klog.KObj(pod), "node", nodeName)
			} else {
				klog.Errorf(LogPrefix+"catch a panic %#v pod:%#v, nodeName:%s", r, klog.KObj(pod), nodeName)
			}
			score, newState = ex.backToDefaultScore(ctx, state, pod, nodeName)
		}
	}()
	return ex.score(ctx, state, pod, nodeName)
}

func (ex *Arbiter) score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (score int64, newState *framework.Status) {
	var err error
	klog.V(10).InfoS(LogPrefix+"Score", "pod", klog.KObj(pod), "node", nodeName)
	logic := ex.manager.GetScoreSpec(ctx)
	if strings.TrimSpace(logic) == "" {
		klog.Infoln(LogPrefix + "no logic")
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	klog.V(10).InfoS(LogPrefix+"ScoreLogic", "pod", klog.KObj(pod), "node", nodeName, "logicStr", logic)

	nodeInfo, err := ex.frameworkHandler.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("getting node %q from Snapshot: %w", nodeName, err))
	}

	registry := new(require.Registry)
	vm := goja.New()
	registry.Enable(vm)
	console.Enable(vm)
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	podMetric, err := ex.manager.GetPodMetric(ctx, pod)
	if err != nil {
		klog.Warningf(LogPrefix+"GetPodMetric failed,pod:%#v use default value instead", klog.KObj(pod))
	}
	node := nodeInfo.Node()
	nodeMetric, err := ex.manager.GetNodeMetric(ctx, node)
	if err != nil {
		klog.Warningf(LogPrefix+"GetNodeMetric failed,node:%#v use default value instead", klog.KObj(node))
	}
	podWithMetric := GetDefaultPodMetric(pod, podMetric)

	/*
		same with node
	*/
	pt, err := json.Marshal(podWithMetric)
	if err != nil {
		klog.ErrorS(err, LogPrefix+"pod json.Marshal error", "pod", klog.KObj(podWithMetric.Pod))
		klog.V(5).ErrorS(err, LogPrefix+"pod json.Marshal error", "podWithMetric", podWithMetric)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	var po map[string]interface{}
	if err = json.Unmarshal(pt, &po); err != nil {
		klog.ErrorS(err, LogPrefix+"pod json.Unmarshal error", "pod", klog.KObj(podWithMetric.Pod))
		klog.V(5).ErrorS(err, LogPrefix+"pod json.Marshal error", "podWithMetric", podWithMetric)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}

	err = vm.Set("pod", po)
	if err != nil {
		klog.ErrorS(err, LogPrefix+"js vm set pod get err", "logic", logic, "pod", klog.KObj(podWithMetric.Pod))
		klog.V(5).ErrorS(err, LogPrefix+"js vm set pod get err", "logic", logic, "podWithMetric", podWithMetric)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	nodeWithMetric := GetDefaultNodeMetric(node, nodeMetric)

	/*
		try to resolve 'node.Status.Capacity cant import' issue.
	*/
	t, err := json.Marshal(nodeWithMetric)
	if err != nil {
		klog.ErrorS(err, LogPrefix+"node json.Marshal error", "node", klog.KObj(nodeWithMetric.Node))
		klog.V(5).ErrorS(err, LogPrefix+"node json.Marshal error", "nodeWithMetric", nodeWithMetric)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	var no map[string]interface{}
	if err = json.Unmarshal(t, &no); err != nil {
		klog.ErrorS(err, LogPrefix+"node json.Unmarshal error", "node", klog.KObj(nodeWithMetric.Node))
		klog.V(5).ErrorS(err, LogPrefix+"node json.Unmarshal error", "nodeWithMetric", nodeWithMetric)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}

	err = vm.Set("node", no)
	if err != nil {
		klog.ErrorS(err, LogPrefix+"js vm set node get err", "logic", logic, "node", klog.KObj(nodeWithMetric.Node))
		klog.V(5).ErrorS(err, LogPrefix+"js vm set node get err", "logic", logic, "node Unmarshal", no)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	klog.Infoln(LogPrefix + "get js val finish")

	if klog.V(4).Enabled() {
		if _, err = vm.RunString(DebugLogic); err != nil {
			klog.ErrorS(err, LogPrefix+"run debug logic error")
		}
		klog.Infoln(LogPrefix + "debug logic finish")
	}

	if _, err = vm.RunString(logic); err != nil {
		klog.ErrorS(err, LogPrefix+"score js logic is not right", "logic", logic)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	klog.V(10).Infoln(LogPrefix + "run js logic finish")

	if v := vm.Get("score"); v == nil {
		klog.ErrorS(ErrNoScoreFunction, LogPrefix+"should write a function score(){...} in score crd, back to default score logic", "logic", logic)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	klog.V(10).Infoln(LogPrefix + "defined there is a score function in js")

	var fn func() float64
	if err = vm.ExportTo(vm.Get("score"), &fn); err != nil {
		klog.V(4).ErrorS(err, LogPrefix+"Score get result 0", "logic", logic)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	klog.V(10).Infoln(LogPrefix + "get score value finish")

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				klog.ErrorS(err, LogPrefix+"Score js logic get panic", "logic", logic, "pod", klog.KObj(podWithMetric.Pod), "node", nodeName)
				klog.V(5).ErrorS(err, LogPrefix+"Score js logic get panic", "logic", logic, "podWithMetric", podWithMetric, "node", nodeName)
			} else {
				klog.ErrorS(fmt.Errorf("panic"), LogPrefix+"Score js logic get panic", "logic", logic, "pod", klog.KObj(podWithMetric.Pod), "node", nodeName, "panic", r)
				klog.V(5).ErrorS(fmt.Errorf("panic"), LogPrefix+"Score js logic get panic", "logic", logic, "podWithMetric", podWithMetric, "node", nodeName, "panic", r)
			}
			score, newState = ex.backToDefaultScore(ctx, state, pod, nodeName)
		}
	}()
	score = int64(fn())
	klog.V(10).InfoS(LogPrefix+"all finish", "score", score, "nodeName", node.Name, "podName", pod.Name)
	return score, nil
}

func GetDefaultPodMetric(pod *v1.Pod, metric manager.MetricData) manager.PodWithMetric {
	if len(metric) != 0 {
		return manager.PodWithMetric{Pod: pod, Metric: metric}
	}
	var cpu, mem int64
	for _, container := range pod.Spec.Containers {
		if container.Resources.Limits.Cpu() != nil {
			cpu += container.Resources.Limits.Cpu().MilliValue()
		}
		if cpu == 0 {
			cpu = 100
		}
		if container.Resources.Limits.Memory() != nil {
			mem += container.Resources.Limits.Memory().Value()
		}
		if mem == 0 {
			mem = 100 * 1024 * 1024
		}
	}
	return manager.PodWithMetric{Pod: pod, Metric: manager.MetricData{
		"mem": &manager.FullMetrics{
			ObservabilityIndicantStatusMetricInfo: &v1alpha1.ObservabilityIndicantStatusMetricInfo{},
			Avg:                                   float64(mem) * PredictedRatio,
			Max:                                   float64(mem) * PredictedRatio,
			Min:                                   float64(mem) * PredictedRatio,
		},
		"cpu": &manager.FullMetrics{
			ObservabilityIndicantStatusMetricInfo: &v1alpha1.ObservabilityIndicantStatusMetricInfo{},
			Avg:                                   float64(cpu) * PredictedRatio,
			Max:                                   float64(cpu) * PredictedRatio,
			Min:                                   float64(cpu) * PredictedRatio,
		},
	}}
}

func GetDefaultNodeMetric(node *v1.Node, metric manager.MetricData) manager.NodeWithMetric {
	if len(metric) != 0 {
		klog.V(10).InfoS("node has obi metrics, use it.", "node", node)
		return manager.NodeWithMetric{Node: node, Metric: metric}
	}
	var cpu, mem int64 = 2000, 4 * 1024 * 1024 * 1024
	if node.Status.Capacity.Memory() != nil {
		mem = node.Status.Capacity.Memory().Value()
	}
	if node.Status.Capacity.Cpu() != nil {
		cpu = node.Status.Capacity.Cpu().MilliValue()
	}
	return manager.NodeWithMetric{Node: node, Metric: manager.MetricData{
		"mem": &manager.FullMetrics{
			ObservabilityIndicantStatusMetricInfo: &v1alpha1.ObservabilityIndicantStatusMetricInfo{StartTime: metav1.NewTime(time.Now().Add(time.Hour * -1)), EndTime: metav1.NewTime(time.Now().Add(time.Hour))},
			Avg:                                   float64(mem) * PredictedRatio,
			Max:                                   float64(mem) * PredictedRatio,
			Min:                                   float64(mem) * PredictedRatio,
		},
		"cpu": &manager.FullMetrics{
			ObservabilityIndicantStatusMetricInfo: &v1alpha1.ObservabilityIndicantStatusMetricInfo{StartTime: metav1.NewTime(time.Now().Add(time.Hour * -1)), EndTime: metav1.NewTime(time.Now().Add(time.Hour))},
			Avg:                                   float64(cpu) * PredictedRatio,
			Max:                                   float64(cpu) * PredictedRatio,
			Min:                                   float64(cpu) * PredictedRatio,
		},
	}}
}

func (ex *Arbiter) ScoreExtensions() framework.ScoreExtensions {
	klog.V(10).Infoln(LogPrefix + "ScoreExtensions")
	return nil
}

func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	klog.V(10).Infof(LogPrefix+"New Arbiter Init Start...%#v", obj)
	restConfig := handle.KubeConfig()
	client, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	schedulerInformer := informerFactory.Arbiter().V1alpha1().Schedulers()
	scoreInformer := informerFactory.Arbiter().V1alpha1().Scores()
	observabilityIndicantInformer := informerFactory.Arbiter().V1alpha1().ObservabilityIndicants()
	podInformer := handle.SharedInformerFactory().Core().V1().Pods()
	nodeInformer := handle.SharedInformerFactory().Core().V1().Nodes()
	replicaSetInformer := handle.SharedInformerFactory().Apps().V1().ReplicaSets()

	mgr := manager.NewManager(client, handle.SnapshotSharedLister(), podInformer, nodeInformer)
	plugin := &Arbiter{
		frameworkHandler: handle,
		manager:          mgr,
	}
	scoreInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    mgr.ScoreAdd,
		UpdateFunc: mgr.ScoreUpdate,
		DeleteFunc: mgr.ScoreDelete,
	})
	schedulerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    mgr.SchedulerAdd,
		UpdateFunc: mgr.SchedulerUpdate,
		DeleteFunc: mgr.SchedulerDelete,
	})
	observabilityIndicantInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    mgr.ObservabilityIndicantAdd,
		UpdateFunc: mgr.ObservabilityIndicantUpdate,
		DeleteFunc: mgr.ObservabilityIndicantDelete,
	})
	replicaSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    mgr.RsAdd,
		UpdateFunc: mgr.RsUpdate,
		DeleteFunc: mgr.RsDelete,
	})
	informerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), scoreInformer.Informer().HasSynced) {
		err := fmt.Errorf("WaitForCacheSync failed")
		klog.ErrorS(err, LogPrefix+"Cannot sync caches")
		return nil, err
	}
	klog.V(10).Infoln(LogPrefix + "New Arbiter Init Finish...")
	return plugin, nil
}

func (ex *Arbiter) Name() string {
	return Name
}

func (ex *Arbiter) PreFilterExtensions() framework.PreFilterExtensions {
	klog.V(10).Infoln(LogPrefix + "PreFilterExtensions")
	return nil
}

func (ex *Arbiter) PostBind(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeName string) {
	klog.V(10).InfoS(LogPrefix+"PostBind", "pod", klog.KObj(pod), "nodeName", nodeName)
}
