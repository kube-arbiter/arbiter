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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/helper"

	"github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
	informers "github.com/kube-arbiter/arbiter/pkg/generated/informers/externalversions"
	"github.com/kube-arbiter/arbiter/pkg/scheduler/manager"
)

const (
	Name       = "Arbiter"
	LogPrefix  = "[arbiter] "
	DebugLogic = `console.log("[arbiter]", "pod:", JSON.stringify(pod), "node:", JSON.stringify(node));`
)

var (
	ErrNoScoreFunction = errors.New("no score function found")
)

type Arbiter struct {
	frameworkHandler framework.Handle
	manager          manager.Manager
}

var _ framework.PostBindPlugin = &Arbiter{}
var _ framework.ScorePlugin = &Arbiter{}
var _ framework.ScoreExtensions = &Arbiter{}

func (ex *Arbiter) NormalizeScore(ctx context.Context, state *framework.CycleState, p *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	return helper.DefaultNormalizeScore(framework.MaxNodeScore, false, scores)
}

func (ex *Arbiter) backToDefaultScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (score int64, newState *framework.Status) {
	klog.V(2).Infoln(LogPrefix+"back to default score function", "pod", klog.KObj(pod), "node", nodeName)
	return 0, nil
}

func (ex *Arbiter) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (score int64, newState *framework.Status) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				klog.V(1).ErrorS(err, LogPrefix+"catch a panic", "pod", klog.KObj(pod), "node", nodeName)
			} else {
				klog.V(1).ErrorS(fmt.Errorf(LogPrefix+"catch a panic %#v", r), "pod", klog.KObj(pod), "node", nodeName)
			}
			score, newState = ex.backToDefaultScore(ctx, state, pod, nodeName)
		}
	}()
	scoreResults, totalWeight := ex.manager.GetScore(ctx, pod.GetNamespace())
	if totalWeight <= 0 {
		klog.V(1).ErrorS(errors.New("all scoreCR totalWeight <= 0"), LogPrefix+"all scoreCR totalWeight <=0", "pod", klog.KObj(pod), "node", nodeName)
		return ex.backToDefaultScore(ctx, state, pod, nodeName)
	}
	ex.frameworkHandler.Parallelizer().Until(ctx, len(scoreResults), func(piece int) {
		subCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		logic := scoreResults[piece].Logic
		nameKey := scoreResults[piece].NameKey
		scoreResults[piece].Result, scoreResults[piece].Err = ex.scoreOne(subCtx, state, pod, nodeName, logic, nameKey)
	})
	msg := strings.Builder{}
	for _, v := range scoreResults {
		score += v.Result * v.Weight
		_, _ = msg.WriteString(fmt.Sprintf("| scoreCR=%s weight=%d score=%d ", v.NameKey, v.Weight, v.Result))
		if v.Err != nil {
			_, _ = msg.WriteString(fmt.Sprintf("err=%s ", v.Err))
		}
		_, _ = msg.WriteString("|")
	}
	score /= totalWeight
	klog.V(1).InfoS(LogPrefix+"Score Result", "pod", klog.KObj(pod), "node", nodeName, "score", score, "scoreDetail", msg.String())
	if score < 0 || score > 100 {
		return 0, framework.AsStatus(errors.New(msg.String()))
	}
	return
}

func (ex *Arbiter) scoreOne(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName, logic, scoreKey string) (score int64, err error) {
	klog.V(5).InfoS(LogPrefix+"Score One", "pod", klog.KObj(pod), "node", nodeName)
	if strings.TrimSpace(logic) == "" {
		return 0, errors.New("no logic")
	}
	klog.V(5).InfoS(LogPrefix+"ScoreLogic", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "logicStr", logic)

	nodeInfo, err := ex.frameworkHandler.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, fmt.Errorf("getting node %q from Snapshot: %w", nodeName, err)
	}

	registry := new(require.Registry)
	vm := goja.New()
	registry.Enable(vm)
	console.Enable(vm)
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	podOBI, err := ex.manager.GetPodOBI(ctx, pod)
	if err != nil {
		klog.V(4).InfoS(LogPrefix+"GetPodOBI failed, use default value instead", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey)
	}
	node := nodeInfo.Node()
	nodeOBI, err := ex.manager.GetNodeOBI(ctx, node.Name)
	if err != nil {
		klog.V(4).InfoS(LogPrefix+"GetNodeOBI failed, use default value instead", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey)
	}
	podWithOBI := &manager.PodWithOBI{Pod: *pod, OBI: podOBI}

	/*
		same with node
	*/
	pt, err := json.Marshal(podWithOBI)
	if err != nil {
		if klog.V(5).Enabled() {
			klog.V(5).ErrorS(err, LogPrefix+"pod json.Marshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey)
		} else if klog.V(4).Enabled() {
			klog.V(4).ErrorS(err, LogPrefix+"pod json.Marshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "podWithOBI", podWithOBI)
		}
		return 0, err
	}
	var po map[string]interface{}
	if err = json.Unmarshal(pt, &po); err != nil {
		if klog.V(5).Enabled() {
			klog.V(5).ErrorS(err, LogPrefix+"pod json.Unmarshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey)
		} else if klog.V(4).Enabled() {
			klog.V(4).ErrorS(err, LogPrefix+"pod json.Unmarshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "podWithOBI", podWithOBI)
		}
		return 0, err
	}

	err = vm.Set("pod", po)
	if err != nil {
		if klog.V(5).Enabled() {
			klog.V(5).ErrorS(err, LogPrefix+"js vm set pod get err", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic)
		} else if klog.V(4).Enabled() {
			klog.V(4).ErrorS(err, LogPrefix+"js vm set pod get err", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic, "podWithOBI", podWithOBI)
		}
		return 0, err
	}
	nodeWithOBI := manager.NodeWithOBI{Node: *node, OBI: nodeOBI, CPUReq: nodeInfo.NonZeroRequested.MilliCPU, MemReq: nodeInfo.NonZeroRequested.Memory}

	/*
		try to resolve 'node.Status.Capacity cant import' issue.
	*/
	t, err := json.Marshal(nodeWithOBI)
	if err != nil {
		if klog.V(5).Enabled() {
			klog.V(5).ErrorS(err, LogPrefix+"node json.Marshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey)
		} else if klog.V(4).Enabled() {
			klog.V(4).ErrorS(err, LogPrefix+"node json.Marshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "nodeWithOBI", nodeWithOBI)
		}
		return 0, err
	}
	var no map[string]interface{}
	if err = json.Unmarshal(t, &no); err != nil {
		if klog.V(5).Enabled() {
			klog.V(5).ErrorS(err, LogPrefix+"node json.Unmarshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey)
		} else if klog.V(4).Enabled() {
			klog.V(4).ErrorS(err, LogPrefix+"node json.Unmarshal error", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "nodeWithOBI", nodeWithOBI)
		}
		return 0, err
	}

	err = vm.Set("node", no)
	if err != nil {
		if klog.V(5).Enabled() {
			klog.V(5).ErrorS(err, LogPrefix+"js vm set node get err", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey)
		} else if klog.V(4).Enabled() {
			klog.V(4).ErrorS(err, LogPrefix+"js vm set node get err", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic, "node Unmarshal", no)
		}
		return 0, err
	}
	klog.Infoln(LogPrefix+"get js val finish", "pod", klog.KObj(pod), "node", nodeName)

	if klog.V(5).Enabled() {
		if _, err = vm.RunString(DebugLogic); err != nil {
			klog.ErrorS(err, LogPrefix+"run debug logic error", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "debugLogic", DebugLogic)
		}
		klog.Infoln(LogPrefix+"debug logic finish", "pod", klog.KObj(pod), "node", nodeName, "debugLogic", DebugLogic, "scoreCR", scoreKey)
	}

	if _, err = vm.RunString(logic); err != nil {
		if klog.V(4).Enabled() {
			klog.V(4).ErrorS(err, LogPrefix+"score js logic is not right", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic, "podWithOBI", podWithOBI, "nodeWithOBI", nodeWithOBI)
		} else {
			klog.V(1).ErrorS(err, LogPrefix+"score js logic is not right", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic)
		}
		return 0, err
	}
	klog.V(5).Infoln(LogPrefix+"run js logic finish", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey)

	if v := vm.Get("score"); v == nil {
		if klog.V(4).Enabled() {
			klog.V(4).ErrorS(ErrNoScoreFunction, LogPrefix+"should write a function score(){...} in score crd, back to default score logic", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey)
		} else {
			klog.V(1).ErrorS(ErrNoScoreFunction, LogPrefix+"should write a function score(){...} in score crd, back to default score logic", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic)
		}
		return 0, err
	}
	klog.V(5).Infoln(LogPrefix+"defined there is a score function in js", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey)

	var fn func() float64
	if err = vm.ExportTo(vm.Get("score"), &fn); err != nil {
		klog.V(4).ErrorS(err, LogPrefix+"Score get error result", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic)
		return 0, err
	}
	klog.V(5).InfoS(LogPrefix+"get score value finish", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey)

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				if klog.V(4).Enabled() {
					klog.V(4).ErrorS(err, LogPrefix+"Score js logic get panic", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic, "podWithOBI", podWithOBI, "nodeWithOBI", nodeWithOBI)
				} else {
					klog.V(1).ErrorS(err, LogPrefix+"Score js logic get panic", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey)
				}
			} else {
				if klog.V(4).Enabled() {
					klog.V(4).ErrorS(fmt.Errorf("get panic:%v", r), LogPrefix+"Score js logic get panic", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey, "logic", logic, "podWithOBI", podWithOBI, "nodeWithOBI", nodeWithOBI)
				} else {
					klog.V(1).ErrorS(fmt.Errorf("get panic:%v", r), LogPrefix+"Score js logic get panic", "pod", klog.KObj(&podWithOBI.Pod), "node", nodeName, "scoreCR", scoreKey)
				}
			}
		}
	}()
	score = int64(fn())
	klog.V(5).InfoS(LogPrefix+"all finish", "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "score", score)
	if score < 0 || score > 100 {
		msg := fmt.Sprintf("ScoreCR:%s returns an invalid score %d, it should in the range of [%v, %v]", scoreKey, score, framework.MinNodeScore, framework.MaxNodeScore)
		klog.ErrorS(errors.New(msg), msg, "pod", klog.KObj(pod), "node", nodeName, "scoreCR", scoreKey, "score", score)
		return 0, errors.New(msg)
	}
	return score, nil
}

func (ex *Arbiter) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	klog.V(5).Infof(LogPrefix+"New Arbiter Init Start...%#v", obj)
	restConfig := handle.KubeConfig()
	client, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	scoreInformer := informerFactory.Arbiter().V1alpha1().Scores()
	observabilityIndicantInformer := informerFactory.Arbiter().V1alpha1().ObservabilityIndicants()
	podInformer := handle.SharedInformerFactory().Core().V1().Pods()
	nodeInformer := handle.SharedInformerFactory().Core().V1().Nodes()

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
	observabilityIndicantInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    mgr.ObservabilityIndicantAdd,
		UpdateFunc: mgr.ObservabilityIndicantUpdate,
		DeleteFunc: mgr.ObservabilityIndicantDelete,
	})
	informerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), scoreInformer.Informer().HasSynced) {
		err := fmt.Errorf("WaitForCacheSync failed")
		klog.ErrorS(err, LogPrefix+"Cannot sync caches")
		return nil, err
	}
	klog.V(5).Infoln(LogPrefix + "New Arbiter Init Finish...")
	return plugin, nil
}

func (ex *Arbiter) Name() string {
	return Name
}

func (ex *Arbiter) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (ex *Arbiter) PostBind(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeName string) {
	klog.V(5).InfoS(LogPrefix+"PostBind", "pod", klog.KObj(pod), "node", nodeName)
}
