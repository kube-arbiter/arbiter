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
	make -C "${ROOT}" image WHAT=webhook GOARCH=amd64 ARCH=amd64 IMAGE_NAME=webhook:e2e
	echo "build webhook image done."
}

function install_prometheus_by_helm {
	echo "prometheus install start..."
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm upgrade --wait --install prometheus prometheus-community/prometheus --namespace kube-system --set prometheus.service.type=NodePort --set prometheus.service.nodePort=30900
	echo "prometheus install finish!"
}

function install_certmanager_by_helm {
	echo "cert manager install start..."
	if [ $(echo $K8S_VERSION | awk -F '.' '{print $2}') -gt 20 ]; then
		CertMangerVersion="v1.10.0"
	else
		CertMangerVersion="v1.8.2"
	fi
	helm repo add jetstack https://charts.jetstack.io
	helm upgrade --wait --install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --version $CertMangerVersion --set installCRDs=true
	echo "cert manager install finish!"
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
	cat $ROOT/manifests/install/observer/observer.yaml | sed -e 's/kubearbiter\/observer:.*/localhost:5001\/arbiter.k8s.com.cn\/observer:e2e/g' -e 's/kubearbiter\/observer-default-plugins:.*/kubearbiter\/observer-default-plugins:dev/g' | kubectl -n ${ARBITER_NS} apply -f -

	kubectl wait deployment --timeout=${TIMEOUT} -n ${ARBITER_NS} observer-server --for condition=Available=True
	echo "observer-server install finish..."

	# install executor rbac
	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/install/executor/executor-rbac.yaml
	# install executor
	# use kubearbiter/executor:dev for e2e test
	cat $ROOT/manifests/install/executor/executor.yaml | sed -e 's/kubearbiter\/executor:.*/localhost:5001\/arbiter.k8s.com.cn\/executor:e2e/g' -e 's/kubearbiter\/executor-default-plugins:.*/kubearbiter\/executor-default-plugins:dev/g' | kubectl -n ${ARBITER_NS} apply -f -
	kubectl wait deployment --timeout=${TIMEOUT} -n ${ARBITER_NS} executor-server --for condition=Available=True
	echo "executor install finish..."
}

function install_webhook {
	kubectl apply -f $ROOT/manifests/install/overcommit/rbac.yaml
	# kubectl apply -f $ROOT/manifests/install/overcommit/cert-manager.yaml
	# kubectl wait deployment -n cert-manager --timeout=${TIMEOUT} cert-manager --for condition=Available=True
	# kubectl wait deployment -n cert-manager --timeout=${TIMEOUT} cert-manager-cainjector --for condition=Available=True
	# kubectl wait deployment -n cert-manager --timeout=${TIMEOUT} cert-manager-webhook --for condition=Available=True
	kubectl apply -f $ROOT/manifests/install/overcommit/overcommit.yaml
	kubectl apply -f $ROOT/manifests/install/overcommit/webhook.yaml
	kubectl apply -f $ROOT/manifests/install/overcommit/cert.yaml
	cat $ROOT/manifests/install/overcommit/overcommit-deployment.yaml | sed -e 's/kubearbiter\/webhook:.*/localhost:5001\/arbiter.k8s.com.cn\/webhook:e2e/g' | kubectl apply -f -
	kubectl wait deployment --timeout=${TIMEOUT} overcommit-webhook --for condition=Available=True
	cat $ROOT/manifests/install/overcommit/test.yaml | sed -e 's/kubearbiter\/webhook:.*/localhost:5001\/arbiter.k8s.com.cn\/webhook:e2e/g' | kubectl apply -f -
	kubectl wait deployment --timeout=${TIMEOUT} test-deployment --for condition=Available=True
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

		# check data valid
		sum=$(kubectl get obi ${name} -n ${ARBITER_NS} -o jsonpath="{.status.metrics.*[*].records.*.value}" | awk 'BEGIN{sum=0} {for(i=1;i<=NF;i++)sum+=$i} END{print(sum)}')
		if [ $sum -eq 0 ]; then
			# The value might not be a number and summed, but it maybe not wrong as the value maybe original raw data from prometheus.
			sum=$(kubectl get obi ${name} -n ${ARBITER_NS} -o jsonpath="{.status.metrics.*[*].records.*.value}" | grep "value" | wc -l)
		fi

		if [ $s -ne 0 ] || [ $sum -eq 0 ]; then
			kubectl get obi -n ${ARBITER_NS} -owide
			echo $(kubectl get obi -n ${ARBITER_NS} -oyaml | grep 'records' | awk '{print NR}' | tail -n1)
			kubectl get obi -n ${ARBITER_NS} -oyaml
			kubectl get nodes --show-labels
			kubectl get po -n ${ARBITER_NS} --show-labels
			exit $s
		fi
	done

	echo "deploy policies wait 30s"
	kubectl -n ${ARBITER_NS} apply -f $ROOT/manifests/example/executor
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
install_certmanager_by_helm

build_image

docker tag observer:e2e localhost:5001/arbiter.k8s.com.cn/observer:e2e
registry_load_image localhost:5001/arbiter.k8s.com.cn/observer:e2e

docker tag executor:e2e localhost:5001/arbiter.k8s.com.cn/executor:e2e
registry_load_image localhost:5001/arbiter.k8s.com.cn/executor:e2e

docker tag webhook:e2e localhost:5001/arbiter.k8s.com.cn/webhook:e2e
registry_load_image localhost:5001/arbiter.k8s.com.cn/webhook:e2e

install_crd
install_webhook

test_obi
test_abctl
