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

package e2e_test

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	DeployNamespace = "default"
	TimeOutSecond   = 60
)

var _ = Describe("Scheduler e2e test", Label("scheduler"), func() {

	Describe("schedule some pod", Label("base", "quick"), func() {
		It("schedule a simple busybox pod", func() {
			const (
				DeployName = "test-default"
				Deploy     = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-default
  name: test-default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: test-default
  template:
    metadata:
      labels:
        app: test-default
    spec:
      containers:
      - command:
        - /bin/sh
        - -ec
        - sleep 1000
        image: busybox
        name: busybox
`
			)
			DeferCleanup(func() {
				Expect(DeleteDeploy(DeployName, DeployNamespace, TimeOutSecond)).Should(Succeed())
			})
			By("1. create deploy")
			Expect(
				CreateByYaml(Deploy, TimeOutSecond)).
				Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			By("2. make sure scheduler success")
			Expect(
				GetPodNodeName(DeployName, DeployNamespace, "", TimeOutSecond)).
				ShouldNot(BeZero(),
					DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			By("3. delete pod")
			Expect(
				DeletePod(DeployName, DeployNamespace, "", TimeOutSecond, false)).
				Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			By("4. wait pod reschedule done")
			Expect(
				GetPodNodeName(DeployName, DeployNamespace, "", TimeOutSecond)).
				ShouldNot(BeZero(),
					DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
		})
		It("schedule a simple busybox pod with nodeSelector", func() {
			const (
				DeployName  = "test-node-selector"
				MasterLabel = "kubernetes.io/hostname=arbiter-e2e-control-plane"
				Deploy      = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-node-selector
  name: test-node-selector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-node-selector
  template:
    metadata:
      labels:
        app: test-node-selector
    spec:
      nodeSelector:
        kubernetes.io/hostname: arbiter-e2e-control-plane
      containers:
      - command:
        - /bin/sh
        - -ec
        - sleep 1000
        image: busybox
        name: busybox
`
			)
			DeferCleanup(func() {
				Expect(DeleteDeploy(DeployName, DeployNamespace, TimeOutSecond)).Should(Succeed())
			})
			By("1. get node master name")
			masterName, err := GetNodeNameByLabel(MasterLabel, TimeOutSecond)
			Expect(err).Error().Should(Succeed())
			By("2. create deploy to master")
			Expect(
				CreateByYaml(Deploy, TimeOutSecond)).
				Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			By("3. make sure scheduler success")
			nodeName, err := GetPodNodeName(DeployName, DeployNamespace, "", TimeOutSecond)
			Expect(nodeName).Should(Equal(masterName),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
		})
	})
	Describe("only use score crd", Label("base", "quick"), func() {
		It("schedule a simple busybox pod", func() {
			const (
				DeployName = "test-default-score"
				Deploy     = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-default-score
  name: test-default-score
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-default-score
  template:
    metadata:
      labels:
        app: test-default-score
    spec:
      containers:
      - command:
        - /bin/sh
        - -ec
        - sleep 1000
        image: busybox
        name: busybox
`
				ScoreYaml = `apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: Score
metadata:
  name: show-demo
  namespace: kube-system
spec:
  logic: |
    function score() {
        var podLabel = pod.raw.metadata.labels;
        if (podLabel.app !== '%s') {
            return 0;
        }
        if (node.raw.metadata.name == '%s') {
            return 100;
        }
        return 0;
    }
`
			)
			DeferCleanup(func() {
				Expect(DeleteDeploy(DeployName, DeployNamespace, TimeOutSecond)).Should(Succeed())
				Expect(DeleteByYaml(ScoreYaml, TimeOutSecond)).Error().Should(Succeed())
			})
			By("1. create deploy")
			Expect(
				CreateByYaml(Deploy, TimeOutSecond)).
				Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			By("2. make sure scheduler success and get node name")
			nodeName, err := GetPodNodeName(DeployName, DeployNamespace, "", TimeOutSecond)
			Expect(nodeName).ShouldNot(BeZero(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			By("3. set score crd to schedule pod to another node")
			allNodeNames, err := GetNodeNameByLabel("", TimeOutSecond)
			Expect(err).Error().Should(Succeed(), DescribeNode("", TimeOutSecond))
			anotherNode := ""
			for _, i := range strings.Split(allNodeNames, " ") {
				if i != nodeName {
					anotherNode = i
					break
				}
			}
			Expect(anotherNode).ShouldNot(BeZero(),
				DescribeNode("", TimeOutSecond))
			Expect(CreateByYaml(fmt.Sprintf(ScoreYaml, DeployName, anotherNode), TimeOutSecond)).Error().Should(Succeed())
			By("3. delete pod")
			Expect(
				DeletePod(DeployName, DeployNamespace, "", TimeOutSecond, false)).
				Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			By("4. pod reschedule to wanted node")
			nodeName, err = GetPodNodeName(DeployName, DeployNamespace, "", TimeOutSecond)
			Expect(nodeName).Should(Equal(anotherNode),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployName, DeployNamespace, "", TimeOutSecond))
		})
	})
	Describe("schedule pod by node real cost", Label("base", "quick"), Serial, func() {
		It("schedule with obi get data from prometheus", func() {
			const (
				DeployCostCPUName = "test-cost-cpu-load"
				DeployCostCPU     = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-cost-cpu-load
  name: test-cost-cpu-load
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-cost-cpu-load
  template:
    metadata:
      labels:
        app: test-cost-cpu-load
    spec:
      containers:
      - command:
        - /consume-cpu/consume-cpu
        - --millicores=1000
        - --duration-sec=6000
        image: kubearbiter/resource-consumer:1.10
        name: consumer
        resources:
          requests:
            cpu: 100m
`
				DeployBusyBoxMulName = "test-busybox-mul"
				DeployBusyBoxMul     = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-busybox-mul
  name: test-busybox-mul
spec:
  replicas: %d
  selector:
    matchLabels:
      app: test-busybox-mul
  template:
    metadata:
      labels:
        app: test-busybox-mul
    spec:
      containers:
      - command:
        - /bin/sh
        - -ec
        - sleep 1000
        image: busybox
        name: busybox
`
				ScoreYaml = `apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: Score
metadata:
  name: show-demo
  namespace: kube-system
spec:
  logic: |
    const NodeCPUOBI = new Map([['arbiter-e2e-control-plane', 'default-prometheus-node-cpu-0'], ['arbiter-e2e-worker', 'default-prometheus-node-cpu-1'],]);
    const NodeMemOBI = new Map([['arbiter-e2e-control-plane', 'default-prometheus-node-mem-0'], ['arbiter-e2e-worker', 'default-prometheus-node-mem-1'],]);

    function getPodCpuMemReq() {
        const DefaultCPUReq = 100; // 0.1 core
        const DefaultMemReq = 200 * 1024 * 1024; // 200MB
        var podContainer = pod.raw.spec.containers;
        if (podContainer == undefined) {
            return [DefaultCPUReq, DefaultMemReq];
        }
        var cpuReq = 0;
        var memReq = 0;
        for (var i = 0; i < podContainer.length; i++) {
            var resources = podContainer[i].resources;
            if (resources.requests == undefined) {
                cpuReq += DefaultCPUReq;
                memReq += DefaultMemReq;
                continue
            }
            cpuReq += cpuParser(resources.requests.cpu);
            memReq += memParser(resources.requests.memory);
        }
        var podInitContainers = pod.raw.spec.initContainers;
        if (podInitContainers == undefined) {
            return [cpuReq, memReq];
        }
        var initCPUReq = 0;
        var initMemReq = 0;
        for (var i = 0; i < podInitContainers.length; i++) {
            var resources = podInitContainers[i].resources;
            if (resources.requests == undefined) {
                initCPUReq = DefaultCPUReq;
                initMemReq = DefaultMemReq;
            } else {
                initCPUReq = cpuParser(resources.requests.cpu);
            }
            if (initCPUReq > cpuReq) {
                cpuReq = initCPUReq;
            }
            if (initMemReq > memReq) {
                memReq = initMemReq;
            }
        }
        return [cpuReq, memReq];
    }

    function cpuParser(input) {
        const milliMatch = input.match(/^([0-9]+)m$/);
        if (milliMatch) {
            return milliMatch[1];
        }

        return parseFloat(input) * 1000;
    }

    function memParser(input) {
        const memoryMultipliers = {
            k: 1000, M: 1000 ** 2, G: 1000 ** 3, Ki: 1024, Mi: 1024 ** 2, Gi: 1024 ** 3,
        };
        const unitMatch = input.match(/^([0-9]+)([A-Za-z]{1,2})$/);
        if (unitMatch) {
            return parseInt(unitMatch[1], 10) * memoryMultipliers[unitMatch[2]];
        }

        return parseInt(input, 10);
    }

    function score() {
        // Feel free to modify this score function to suit your needs.
        // This score function replaces the default score function in the scheduling framework.
        // It inputs the pod and node to be scheduled, and outputs a number (usually 0 to 100).
        // The higher the number, the more the pod tends to be scheduled to this node.
        // The current example shows the scoring based on the actual cpu usage of the node.
        var req = getPodCpuMemReq();
        var podCPUReq = req[0];
        var podMemReq = req[1];
        var nodeName = node.raw.metadata.name;
        var capacity = node.raw.status.allocatable;
        var cpuCap = cpuParser(capacity.cpu);
        var memCap = memParser(capacity.memory);
        var cpuUsed = node.cpuReq;
        var memUsed = node.memReq;
        var cpuReal = node.obi[NodeCPUOBI.get(nodeName)].metric.cpu;
        if (cpuReal == undefined || cpuReal.avg == undefined) {
            console.error('[arbiter-js] cant find node cpu metric', nodeName);
        } else {
            cpuUsed = cpuReal.avg;  // if has metric, use metric instead
        }
        var memReal = node.obi[NodeMemOBI.get(nodeName)].metric.memory;
        if (memReal == undefined || memReal.avg == undefined) {
            console.error('[arbiter-js] cant find node mem metric', nodeName);
        } else {
            memUsed = memReal.avg;  // if has metric, use metric instead
        }
        console.log('[arbiter-js] cpuUsed', cpuUsed);
        // LeastAllocated
        var cpuScore = (cpuCap - cpuUsed - podCPUReq) / cpuCap;
        console.log('[arbiter-js] cpuScore:', cpuScore, 'nodeName', nodeName, 'cpuCap', cpuCap, 'cpuUsed', cpuUsed, 'podCPUReq', podCPUReq);
        var memScore = (memCap - memUsed - podMemReq) / memCap;
        console.log('[arbiter-js] memScore:', memScore, 'nodeName', nodeName, 'memCap', memCap, 'memUsed', memUsed, 'podMemReq', podMemReq);
        return (cpuScore + memScore) / 2 * 100;
    }
`
				OBITemplate = `
apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: ObservabilityIndicant
metadata:
  name: prometheus-node-cpu-%d
  labels:
    test: node-cpu-load-aware
spec:
  metric:
    historyLimit: 1
    metricIntervalSeconds: 30
    metrics:
      cpu:
        aggregations: []
        description: cpu
        query: sum(irate(container_cpu_usage_seconds_total{instance="{{.metadata.name}}"}[5m])) /2 *1000
        unit: m
    timeRangeSeconds: 600
  source: prometheus
  targetRef:
    group: ""
    index: %d
    kind: Node
    labels:
      "data-test": "data-test"
    version: v1
status:
  conditions: []
  phase: ""
  metrics: {}
---
apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: ObservabilityIndicant
metadata:
  name: prometheus-node-mem-%d
spec:
  metric:
    historyLimit: 1
    metricIntervalSeconds: 30
    metrics:
      memory:
        aggregations: []
        description: memory
        query: sum(node_memory_MemTotal_bytes{node="{{.metadata.name}}"} - node_memory_MemAvailable_bytes{node="{{.metadata.name}}"})
        unit: byte
    timeRangeSeconds: 600
  source: prometheus
  targetRef:
    group: ""
    index: %d
    kind: Node
    labels:
      "data-test": "data-test"
    name: ""
    namespace: ""
    version: v1
status:
  conditions: []
  phase: ""
  metrics: {}
`
			)
			var nodesNum int
			DeferCleanup(func() {
				Expect(DeleteDeploy(DeployCostCPUName, DeployNamespace, TimeOutSecond)).Should(Succeed())
				Expect(DeleteDeploy(DeployBusyBoxMulName, DeployNamespace, TimeOutSecond)).Should(Succeed())
				Expect(DeleteByYaml(ScoreYaml, TimeOutSecond)).Error().Should(Succeed())
				for i := 0; i < nodesNum; i++ {
					Expect(DeleteByYaml(fmt.Sprintf(OBITemplate, i, i, i, i), TimeOutSecond)).Error().Should(Succeed())
				}
			})
			By("1. Create a pod on a node that consumes almost all cpu, with cpu request of 100m")
			Expect(
				CreateByYaml(DeployCostCPU, TimeOutSecond)).
				Error().Should(Succeed(),
				DescribePod(DeployCostCPUName, DeployNamespace, "", TimeOutSecond))
			CostCPUNodeName, err := GetPodNodeName(DeployCostCPUName, DeployNamespace, "", TimeOutSecond)
			Expect(CostCPUNodeName).ShouldNot(BeZero(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))

			By("2. Get total node numbers")
			allNodeNames, err := GetNodeNameByLabel("", TimeOutSecond)
			Expect(allNodeNames).ShouldNot(BeZero(),
				DescribeNode("", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribeNode("", TimeOutSecond))
			nodesNum = len(strings.Split(allNodeNames, " "))

			By("3. Create node OBI to get node metrics")
			for i := 0; i < nodesNum; i++ {
				Expect(CreateByYaml(fmt.Sprintf(OBITemplate, i, i, i, i), TimeOutSecond)).Error().Should(Succeed())
			}
			Eventually(
				func() (string, error) {
					return GetOBIRecords(DeployNamespace, "test=node-cpu-load-aware", TimeOutSecond)
				}).
				WithTimeout(5 * TimeOutSecond * time.Second).WithPolling(10 * time.Second).ShouldNot(BeZero())

			By("4. Create a busybox deploy with replicas = 2 * nodeNums")
			Expect(
				CreateByYaml(fmt.Sprintf(DeployBusyBoxMul, nodesNum*2), TimeOutSecond)).
				Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))

			By("5. Make sure busybox be distributed in all nodes.")
			nodeName, err := GetPodNodeName(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond)
			Expect(nodeName).ShouldNot(BeZero(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			nodeNameKey := make(map[string]bool, nodesNum)
			for _, i := range strings.Split(nodeName, " ") {
				nodeNameKey[i] = true
			}
			Expect(len(nodeNameKey)).Should(Equal(nodesNum),
				DescribeNode("", TimeOutSecond))

			By("6. Enable cpu load-aware scheduling")
			Expect(CreateByYaml(ScoreYaml, TimeOutSecond)).Error().Should(Succeed())

			By("7. wait 600s to make prometheus obi update data", func() { time.Sleep(600 * time.Second) })

			By("8. Delete busybox all pod to make them reschedule")
			Expect(
				DeletePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond, false)).
				Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))

			By("9. busybox pod will not schedule in cost-cpu pod's node")
			newNodeName, err := GetPodNodeName(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond)
			Expect(newNodeName).ShouldNot(BeZero(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			allPodsReScheduleToLowCPUNode := true
			for _, i := range strings.Split(newNodeName, " ") {
				if i == CostCPUNodeName {
					allPodsReScheduleToLowCPUNode = false
				}
			}
			Expect(allPodsReScheduleToLowCPUNode).Should(BeTrue(),
				"%s\n\n%s\n\n%s\n\n%s",
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond),
				ShowOBI(DeployNamespace, "test=node-cpu-load-aware", TimeOutSecond),
				TopNode(TimeOutSecond),
				TopPod(TimeOutSecond),
			)
		})
		It("schedule with obi get data from metrics-server", func() {
			const (
				DeployCostCPUName = "test-cost-cpu-load-ms"
				DeployCostCPU     = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-cost-cpu-load-ms
  name: test-cost-cpu-load-ms
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-cost-cpu-load-ms
  template:
    metadata:
      labels:
        app: test-cost-cpu-load-ms
    spec:
      containers:
      - command:
        - /consume-cpu/consume-cpu
        - --millicores=1000
        - --duration-sec=6000
        image: kubearbiter/resource-consumer:1.10
        name: consumer
        resources:
          requests:
            cpu: 100m
`
				DeployBusyBoxMulName = "test-busybox-mul-ms"
				DeployBusyBoxMul     = `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-busybox-mul-ms
  name: test-busybox-mul-ms
spec:
  replicas: %d
  selector:
    matchLabels:
      app: test-busybox-mul-ms
  template:
    metadata:
      labels:
        app: test-busybox-mul-ms
    spec:
      containers:
      - command:
        - /bin/sh
        - -ec
        - sleep 1000
        image: busybox
        name: busybox
`
				ScoreYaml = `apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: Score
metadata:
  name: show-demo
  namespace: kube-system
spec:
  logic: |
    const NodeCPUOBI = new Map([['arbiter-e2e-control-plane', 'default-metrics-server-node-cpu-0'], ['arbiter-e2e-worker', 'default-metrics-server-node-cpu-1'],]);
    const NodeMemOBI = new Map([['arbiter-e2e-control-plane', 'default-metrics-server-node-mem-0'], ['arbiter-e2e-worker', 'default-metrics-server-node-mem-1'],]);

    function getPodCpuMemReq() {
        const DefaultCPUReq = 100; // 0.1 core
        const DefaultMemReq = 200 * 1024 * 1024; // 200MB
        var podContainer = pod.raw.spec.containers;
        if (podContainer == undefined) {
            return [DefaultCPUReq, DefaultMemReq];
        }
        var cpuReq = 0;
        var memReq = 0;
        for (var i = 0; i < podContainer.length; i++) {
            var resources = podContainer[i].resources;
            if (resources.requests == undefined) {
                cpuReq += DefaultCPUReq;
                memReq += DefaultMemReq;
                continue
            }
            cpuReq += cpuParser(resources.requests.cpu);
            memReq += memParser(resources.requests.memory);
        }
        var podInitContainers = pod.raw.spec.initContainers;
        if (podInitContainers == undefined) {
            return [cpuReq, memReq];
        }
        var initCPUReq = 0;
        var initMemReq = 0;
        for (var i = 0; i < podInitContainers.length; i++) {
            var resources = podInitContainers[i].resources;
            if (resources.requests == undefined) {
                initCPUReq = DefaultCPUReq;
                initMemReq = DefaultMemReq;
            } else {
                initCPUReq = cpuParser(resources.requests.cpu);
            }
            if (initCPUReq > cpuReq) {
                cpuReq = initCPUReq;
            }
            if (initMemReq > memReq) {
                memReq = initMemReq;
            }
        }
        return [cpuReq, memReq];
    }

    function cpuParser(input) {
        const milliMatch = input.match(/^([0-9]+)m$/);
        if (milliMatch) {
            return milliMatch[1];
        }

        return parseFloat(input) * 1000;
    }

    function memParser(input) {
        const memoryMultipliers = {
            k: 1000, M: 1000 ** 2, G: 1000 ** 3, Ki: 1024, Mi: 1024 ** 2, Gi: 1024 ** 3,
        };
        const unitMatch = input.match(/^([0-9]+)([A-Za-z]{1,2})$/);
        if (unitMatch) {
            return parseInt(unitMatch[1], 10) * memoryMultipliers[unitMatch[2]];
        }

        return parseInt(input, 10);
    }

    function score() {
        // Feel free to modify this score function to suit your needs.
        // This score function replaces the default score function in the scheduling framework.
        // It inputs the pod and node to be scheduled, and outputs a number (usually 0 to 100).
        // The higher the number, the more the pod tends to be scheduled to this node.
        // The current example shows the scoring based on the actual cpu usage of the node.
        var req = getPodCpuMemReq();
        var podCPUReq = req[0];
        var podMemReq = req[1];
        var nodeName = node.raw.metadata.name;
        var capacity = node.raw.status.allocatable;
        var cpuCap = cpuParser(capacity.cpu);
        var memCap = memParser(capacity.memory);
        var cpuUsed = node.cpuReq;
        var memUsed = node.memReq;
        var cpuReal = node.obi[NodeCPUOBI.get(nodeName)].metric.cpu;
        if (cpuReal == undefined || cpuReal.avg == undefined) {
            console.error('[arbiter-js] cant find node cpu metric', nodeName);
        } else {
            cpuUsed = cpuReal.avg;  // if has metric, use metric instead
        }
        var memReal = node.obi[NodeMemOBI.get(nodeName)].metric.memory;
        if (memReal == undefined || memReal.avg == undefined) {
            console.error('[arbiter-js] cant find node mem metric', nodeName);
        } else {
            memUsed = memReal.avg;  // if has metric, use metric instead
        }
        console.log('[arbiter-js] cpuUsed', cpuUsed);
        // LeastAllocated
        var cpuScore = (cpuCap - cpuUsed - podCPUReq) / cpuCap;
        console.log('[arbiter-js] cpuScore:', cpuScore, 'nodeName', nodeName, 'cpuCap', cpuCap, 'cpuUsed', cpuUsed, 'podCPUReq', podCPUReq);
        var memScore = (memCap - memUsed - podMemReq) / memCap;
        console.log('[arbiter-js] memScore:', memScore, 'nodeName', nodeName, 'memCap', memCap, 'memUsed', memUsed, 'podMemReq', podMemReq);
        return (cpuScore + memScore) / 2 * 100;
    }
`
				OBITemplate = `
apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: ObservabilityIndicant
metadata:
  name: metrics-server-node-cpu-%d
  labels:
    test: node-cpu-load-aware-ms
spec:
  metric:
    historyLimit: 1
    metricIntervalSeconds: 30
    metrics:
      cpu:
        aggregations:
        - time
        description: ""
        query: ""
        unit: 'm'
    timeRangeSeconds: 600
  source: metrics-server
  targetRef:
    group: ""
    index: %d
    kind: Node
    labels:
      data-test: data-test
    name: ""
    namespace: ""
    version: v1
status:
  conditions: []
  phase: ""
  metrics: {}
---
apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: ObservabilityIndicant
metadata:
  name: metrics-server-node-mem-%d
  labels:
    test: node-cpu-load-aware-ms
spec:
  metric:
    historyLimit: 1
    metricIntervalSeconds: 30
    metrics:
      memory:
        aggregations:
        - time
        description: ""
        query: ""
        unit: "byte"
    timeRangeSeconds: 600
  source: metrics-server
  targetRef:
    group: ""
    index: %d
    kind: Node
    labels:
      data-test: data-test
    name: ""
    namespace: ""
    version: v1
status:
  conditions: []
  phase: ""
  metrics: {}
`
			)
			var nodesNum int
			DeferCleanup(func() {
				Expect(DeleteDeploy(DeployCostCPUName, DeployNamespace, TimeOutSecond)).Should(Succeed())
				Expect(DeleteDeploy(DeployBusyBoxMulName, DeployNamespace, TimeOutSecond)).Should(Succeed())
				Expect(DeleteByYaml(ScoreYaml, TimeOutSecond)).Error().Should(Succeed())
				for i := 0; i < nodesNum; i++ {
					Expect(DeleteByYaml(fmt.Sprintf(OBITemplate, i, i, i, i), TimeOutSecond)).Error().Should(Succeed())
				}
			})
			By("1. Create a pod on a node that consumes almost all cpu, with cpu request of 100m")
			Expect(
				CreateByYaml(DeployCostCPU, TimeOutSecond)).
				Error().Should(Succeed(),
				DescribePod(DeployCostCPUName, DeployNamespace, "", TimeOutSecond))
			CostCPUNodeName, err := GetPodNodeName(DeployCostCPUName, DeployNamespace, "", TimeOutSecond)
			Expect(CostCPUNodeName).ShouldNot(BeZero(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))

			By("2. Get total node numbers")
			allNodeNames, err := GetNodeNameByLabel("", TimeOutSecond)
			Expect(allNodeNames).ShouldNot(BeZero(),
				DescribeNode("", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribeNode("", TimeOutSecond))
			nodesNum = len(strings.Split(allNodeNames, " "))

			By("3. Create node OBI to get node metrics")
			for i := 0; i < nodesNum; i++ {
				Expect(CreateByYaml(fmt.Sprintf(OBITemplate, i, i, i, i), TimeOutSecond)).Error().Should(Succeed())
			}
			Eventually(
				func() (string, error) {
					return GetOBIRecords(DeployNamespace, "test=node-cpu-load-aware-ms", TimeOutSecond)
				}).
				WithTimeout(5 * TimeOutSecond * time.Second).WithPolling(10 * time.Second).ShouldNot(BeZero())

			By("4. Create a busybox deploy with replicas = 2 * nodeNums")
			Expect(
				CreateByYaml(fmt.Sprintf(DeployBusyBoxMul, nodesNum*2), TimeOutSecond)).
				Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))

			By("5. Make sure busybox be distributed in all nodes.")
			nodeName, err := GetPodNodeName(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond)
			Expect(nodeName).ShouldNot(BeZero(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			nodeNameKey := make(map[string]bool, nodesNum)
			for _, i := range strings.Split(nodeName, " ") {
				nodeNameKey[i] = true
			}
			Expect(len(nodeNameKey)).Should(Equal(nodesNum),
				DescribeNode("", TimeOutSecond))

			By("6. Enable cpu load-aware scheduling")
			Expect(CreateByYaml(ScoreYaml, TimeOutSecond)).Error().Should(Succeed())

			By("7. wait 600s to make metrics-server and obi get data", func() { time.Sleep(600 * time.Second) })

			By("8. Delete busybox all pod to make them reschedule")
			Expect(
				DeletePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond, false)).
				Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))

			By("9. busybox pod will not schedule in cost-cpu pod's node")
			newNodeName, err := GetPodNodeName(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond)
			Expect(newNodeName).ShouldNot(BeZero(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			Expect(err).Error().Should(Succeed(),
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond))
			allPodsReScheduleToLowCPUNode := true
			for _, i := range strings.Split(newNodeName, " ") {
				if i == CostCPUNodeName {
					allPodsReScheduleToLowCPUNode = false
				}
			}
			Expect(allPodsReScheduleToLowCPUNode).Should(BeTrue(),
				"%s\n\n%s\n\n%s\n\n%s",
				DescribePod(DeployBusyBoxMulName, DeployNamespace, "", TimeOutSecond),
				ShowOBI(DeployNamespace, "test=node-cpu-load-aware-ms", TimeOutSecond),
				TopNode(TimeOutSecond),
				TopPod(TimeOutSecond),
			)
		})
	})
})
