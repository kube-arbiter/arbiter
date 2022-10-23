#!/usr/bin/env bash

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/..
ENABLE_DAEMON=${ENABLE_DAEMON:-true}
LOG_DIR=${LOG_DIR:-"/tmp"}
OUT_DIR=${OUT_DIR:-"/tmp/arbiter"}
TIMEOUT=${TIMEOUT:-300}s
ARBITER_NS="arbiter-system"

# registry_load_image arg is image name
function registry_load_image {
	local image=$1
	docker push "${image}"
	echo "registry load image done."
}

function build_image {
	echo "building image..."
	make -C "${ROOT}" image WHAT=observer GOARCH=amd64 ARCH=amd64 IMAGE_NAME=observer:e2e
	echo "build observer image done."
	make -C "${ROOT}" image WHAT=executor GOARCH=amd64 ARCH=amd64 IMAGE_NAME=executor:e2e
	echo "build executor image done."
}

function install_prometheus_by_helm {
	echo "prometheus install start..."
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm upgrade --wait --install prometheus prometheus-community/prometheus --namespace kube-system --set prometheus.service.type=NodePort --set prometheus.service.nodePort=30900
	echo "prometheus install finish!"
}

function install_metrics_server_by_helm {
	echo "metrics-server install start..."
	helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
	helm upgrade --install metrics-server metrics-server/metrics-server
	# https://github.com/kubernetes-sigs/metrics-server/issues/917#issuecomment-986732226
	kubectl scale --replicas=0 --timeout=${TIMEOUT} -n default deployment/metrics-server
	kubectl patch deploy -n default metrics-server --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'
	kubectl scale --replicas=1 --timeout=${TIMEOUT} -n default deployment/metrics-server
	kubectl wait deployment --timeout=${TIMEOUT} -n default metrics-server --for condition=Available=True
	echo "metrics-server install finish!"
}

function install_crd {
	# prepare-k8s already installed.
	# kubectl apply -f $ROOT/manifests/crds/
	# install metric, prometheus.
	kubectl create ns ${ARBITER_NS}

	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/install/observer/observer-rbac.yaml
	# install observer metric-server
	# use kubearbiter/observer-metric-server:dev for e2e test
	cat $ROOT/manifests/install/observer/observer-metric-server.yaml | sed -e 's/kubearbiter\/observer:.*/localhost:5001\/arbiter.k8s.com.cn\/observer:e2e/g' -e 's/kubearbiter\/observer-metric-server:.*/kubearbiter\/observer-metric-server:dev/g' | kubectl -n ${ARBITER_NS} apply -f -
	# install observer prometheus
	# use kubearbiter/observer-promtheus:dev for e2e test
	cat $ROOT/manifests/install/observer/observer-prometheus.yaml | sed -e 's/kubearbiter\/observer:.*/localhost:5001\/arbiter.k8s.com.cn\/observer:e2e/g' -e 's/kubearbiter\/observer-prometheus-server:.*/kubearbiter\/observer-prometheus-server:dev/g' | kubectl -n ${ARBITER_NS} apply -f -

	kubectl wait deployment --timeout=${TIMEOUT} -n ${ARBITER_NS} observer-metric-server --for condition=Available=True
	echo "observer-metric-server install finish..."
	kubectl wait deployment --timeout=${TIMEOUT} -n ${ARBITER_NS} observer-prometheus --for condition=Available=True
	echo "observer-prometheus install finish..."
	# install executor rbac
	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/install/executor/executor-rbac.yaml
	# install executor
	# use kubearbiter/executor-resource-tagger:dev for e2e test
	cat $ROOT/manifests/install/executor/executor-resource-tagger.yaml | sed -e 's/kubearbiter\/executor:.*/localhost:5001\/arbiter.k8s.com.cn\/executor:e2e/g' -e 's/kubearbiter\/executor-resource-tagger:.*/kubearbiter\/executor-resource-tagger:dev/g' | kubectl -n ${ARBITER_NS} apply -f -
	kubectl wait deployment --timeout=${TIMEOUT} -n ${ARBITER_NS} executor-resource-tagger --for condition=Available=True
	echo "executor install finish..."
}

function test_obi {
	# install mem-cost deployment
	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/example/mem-cost-deploy.yaml
	kubectl wait deployment -n ${ARBITER_NS} mem-cost --for condition=Available=True --timeout=60s
	echo "deploy mem-cost install successfully in ${ARBITER_NS} namespace."

	echo "deploy metric, prometheus obi"
	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/example/observer/metric
	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/example/observer/prometheus

	echo "wait 240s for data to fill"
	sleep 240
	for i in $(seq 1 6); do
		[ $(kubectl get obi -n ${ARBITER_NS} | grep -v 'NAME' | awk '{print NR}' | tail -n1) -eq 11 ] && s=0 && break || s=$? && sleep 10
	done
	(exit $s)

	if [ $? -ne 0 ]; then
		exit 1
	fi

	echo "testing obi have data"
	for name in metric-server-pod-cpu metric-server-pod-mem prometheus-pod-cpu prometheus-pod-mem metric-server-node-cpu metric-server-node-mem prometheus-node-cpu prometheus-node-mem prometheus-cluster-schedulable-cpu prometheus-max-available-cpu prometheus-rawdata-node-unschedule; do
		for i in $(seq 1 2); do
			[ $(kubectl get obi ${name} -n ${ARBITER_NS} -oyaml | grep 'timestamp' | awk '{print NR}' | tail -n1) -ge 2 ] && s=0 && break || s=$? && sleep 10
		done
		if [ $s -ne 0 ]; then
			kubectl get obi -n ${ARBITER_NS} -owide
			echo $(kubectl get obi -n ${ARBITER_NS} -oyaml | grep 'records' | awk '{print NR}' | tail -n1)
			kubectl get obi -n ${ARBITER_NS} -oyaml
			kubectl get nodes --show-labels
			kubectl get po -n ${ARBITER_NS} --show-labels
			exit $s
		fi
	done

	echo "deploy policies wait 30s"
	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/example/executor/resource-tagger
	sleep 30

	for i in $(seq 1 6); do
		[ $(kubectl get ObservabilityActionPolicy -n ${ARBITER_NS} | grep -v "NAME" | awk '{print NR}' | tail -n1) -eq 8 ] && s=0 && break || s=$? && sleep 10
	done
	(exit $s)
	if [ $? -ne 0 ]; then
		exit 1
	fi

	for pod_label in metric-server-pod-cpu metric-server-pod-mem prometheus-pod-cpu prometheus-pod-mem; do
		pod_count=$(kubectl get po -n ${ARBITER_NS} -l${pod_label} | grep -v 'NAME' | awk '{print NR}' | tail -n1)
		if [ $pod_count -ne 1 ]; then
			echo "expect one pod by label $pod_label"
			exit 1
		fi
	done
	echo "target pod has been labeled"

	for node_label in metric-server-node-cpu metric-server-node-mem prometheus-node-cpu prometheus-node-mem; do
		node_count=$(kubectl get node -l${node_label} | grep -v 'NAME' | awk '{print NR}' | tail -n1)
		if [ $node_count -ne 1 ]; then
			echo "expect one node by label $node_label"
			exit 1
		fi
	done
	echo "target node has benn labeled"
	echo "obi test done"
}

function test_abctl() {
	echo "build abctl command"
	goos=$(go env GOOS)
	goarch=$(go env GOARCH)
	make -C "${ROOT}" binary WHAT=abctl GOARCH=$goarch GOOS=$goos

	echo "expect 11 obi"
	obi_count=$($ROOT/_output/bin/$goos/$goarch/abctl -n ${ARBITER_NS} get -mcpu | grep -v 'NAME' | awk '{print NR}' | tail -n1)
	if [ $obi_count -ne 11 ]; then
		echo "get $obi_count obi, not 11"
		exit 1
	fi
	echo "test abctl successfully"
}

set -eE
install_prometheus_by_helm
install_metrics_server_by_helm

build_image

docker tag observer:e2e localhost:5001/arbiter.k8s.com.cn/observer:e2e
registry_load_image localhost:5001/arbiter.k8s.com.cn/observer:e2e

docker tag executor:e2e localhost:5001/arbiter.k8s.com.cn/executor:e2e
registry_load_image localhost:5001/arbiter.k8s.com.cn/executor:e2e

install_crd

test_obi
test_abctl
