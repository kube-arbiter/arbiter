#!/usr/bin/env bash
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/..
ENABLE_DAEMON=${ENABLE_DAEMON:-true}
LOG_DIR=${LOG_DIR:-"/tmp"}
OUT_DIR=${OUT_DIR:-"/tmp/arbiter"}
TIMEOUT=${TIMEOUT:-300}s

function kind_up_cluster {
	# when update kind version, please change this file and github action file.
	# https://github.com/kubernetes-sigs/kind/releases
	case $K8S_VERSION in
	v1.18 | v1.18.20)
		go install sigs.k8s.io/kind@v0.15.0
		mv $HOME/go/bin/kind /usr/local/bin/kind
		kind version
		kind_image="kindest/node:v1.18.20@sha256:61c9e1698c1cb19c3b1d8151a9135b379657aee23c59bde4a8d87923fcb43a91"
		;;
	v1.19 | v1.19.16)
		go install sigs.k8s.io/kind@v0.17.0
		mv $HOME/go/bin/kind /usr/local/bin/kind
		kind version
		kind_image="kindest/node:v1.19.16@sha256:707469aac7e6805e52c3bde2a8a8050ce2b15decff60db6c5077ba9975d28b98"
		;;
	v1.20 | v1.20.15)
		go install sigs.k8s.io/kind@v0.17.0
		mv $HOME/go/bin/kind /usr/local/bin/kind
		kind version
		kind_image="kindest/node:v1.20.15@sha256:d67de8f84143adebe80a07672f370365ec7d23f93dc86866f0e29fa29ce026fe"
		;;
	v1.21 | v1.21.14)
		kind_image="kindest/node:v1.21.14@sha256:f9b4d3d1112f24a7254d2ee296f177f628f9b4c1b32f0006567af11b91c1f301"
		;;
	v1.22 | v1.22.13)
		kind_image="kindest/node:v1.22.13@sha256:4904eda4d6e64b402169797805b8ec01f50133960ad6c19af45173a27eadf959"
		;;
	v1.23 | v1.23.10)
		kind_image="kindest/node:v1.23.10@sha256:f047448af6a656fae7bc909e2fab360c18c487ef3edc93f06d78cdfd864b2d12"
		;;
	v1.24 | v1.24.4)
		kind_image="kindest/node:v1.24.4@sha256:adfaebada924a26c2c9308edd53c6e33b3d4e453782c0063dc0028bdebaddf98"
		;;
	v1.25 | v1.25.0)
		kind_image="kindest/node:v1.25.0@sha256:428aaa17ec82ccde0131cb2d1ca6547d13cf5fdabcc0bbecf749baa935387cbf"
		;;
	*)
		echo $K8S_VERSION "is not support"
		exit 1
		;;
	esac
	echo "kind to create cluster"
	kind create cluster --name arbiter-e2e --config=$ROOT/tests/kind-with-registry.yaml --image $kind_image
	# connect the registry to the cluster network if not already connected
	if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "kind-registry")" = 'null' ]; then
		docker network connect "kind" "kind-registry"
	fi
	echo "kind cluster created done."
}

function up_registry {
	echo "Running registry ..."
	# create registry container unless it already exists
	if [ "$(docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)" != 'true' ]; then
		docker run \
			-d --restart=always -p "127.0.0.1:5001:5000" --name "kind-registry" \
			registry:2
	fi
	echo "registry started."
}

# registry_load_image arg is image name
function registry_load_image {
	local image=$1
	docker push "${image}"
	echo "registry load image done."
}

# clean up
function cleanup {
	echo "Cleaning up..."
	echo "1. kind delete cluster arbiter-e2e"
	kind delete cluster --name arbiter-e2e
	echo "2. docker stop and rm kind-registry"
	docker stop kind-registry && docker rm kind-registry && docker network rm kind
	echo "3. delete tmp file"
	rm -rf ${OUT_DIR}
	echo "cleanup done."
}

if [[ ${ENABLE_DAEMON} == false ]]; then
	trap cleanup EXIT
else
	trap cleanup ERR
	trap cleanup INT
fi

function build_image {
	echo "building image..."
	make -C "${ROOT}" image WHAT=scheduler IMAGE_NAME=scheduler:e2e
	echo "build image done."
}

function check_control_plane_ready {
	echo "wait the node control-plane ready..."
	kubectl wait --timeout=${TIMEOUT} --for=condition=Ready node/arbiter-e2e-control-plane
}

function change_default_scheduler_config {
	docker exec arbiter-e2e-control-plane /bin/bash -c "cp /etc/kubernetes/kube-scheduler-new.yaml /etc/kubernetes/manifests/kube-scheduler.yaml"
	kubectl wait --timeout=${TIMEOUT} --for=condition=Ready -n kube-system pod/kube-scheduler-arbiter-e2e-control-plane
}

function remove_config_to_tmp {
	# resolve mac or win docker limitation
	cat >/tmp/kube-scheduler-new.yaml <<EOF
apiVersion: v1
kind: Pod
metadata:
  labels:
    component: kube-scheduler
    tier: control-plane
  name: kube-scheduler
  namespace: kube-system
spec:
  containers:
    - command:
        - kube-scheduler
        - --authentication-kubeconfig=/etc/kubernetes/scheduler.conf
        - --authorization-kubeconfig=/etc/kubernetes/scheduler.conf
        - --bind-address=127.0.0.1
        - --kubeconfig=/etc/kubernetes/scheduler.conf
        - --config=/etc/kubernetes/kube-scheduler-arbiter-config.yaml
        - --v=5
      image: localhost:5001/arbiter.k8s.com.cn/scheduler:latest
      imagePullPolicy: IfNotPresent
      livenessProbe:
        failureThreshold: 8
        httpGet:
          host: 127.0.0.1
          path: /healthz
          port: 10259
          scheme: HTTPS
        initialDelaySeconds: 10
        periodSeconds: 10
        timeoutSeconds: 15
      name: kube-scheduler
      resources:
        requests:
          cpu: 100m
      startupProbe:
        failureThreshold: 24
        httpGet:
          host: 127.0.0.1
          path: /healthz
          port: 10259
          scheme: HTTPS
        initialDelaySeconds: 10
        periodSeconds: 10
        timeoutSeconds: 15
      volumeMounts:
        - mountPath: /etc/kubernetes/scheduler.conf
          name: kubeconfig
          readOnly: true
        - mountPath: /etc/kubernetes/kube-scheduler-arbiter-config.yaml
          name: config
          readOnly: true
  hostNetwork: true
  priorityClassName: system-node-critical
  securityContext:
    seccompProfile:
      type: RuntimeDefault
  volumes:
    - hostPath:
        path: /etc/kubernetes/scheduler.conf
        type: FileOrCreate
      name: kubeconfig
    - hostPath:
        path: /etc/kubernetes/kube-scheduler-arbiter-config.yaml
        type: FileOrCreate
      name: config
EOF
	cp $ROOT/manifests/install/scheduler/kube-scheduler-arbiter-config.yaml /tmp
}

function install_scheduler_crd {
	kubectl apply -f $ROOT/manifests/crds/
	kubectl apply -f $ROOT/manifests/install/scheduler/scheduler-rbac.yaml
	kubectl apply -f $ROOT/manifests/install/scheduler/score.yaml
}

function remove_master_taint {
	kubectl taint nodes arbiter-e2e-control-plane node-role.kubernetes.io/master- 2>/dev/null || true
	kubectl taint nodes arbiter-e2e-control-plane node-role.kubernetes.io/control-plane- 2>/dev/null || true
}

# main
echo "Run arbiter e2e test for k8s: "$K8S_VERSION
cleanup
set -eE
build_image
remove_config_to_tmp
up_registry
kind_up_cluster
check_control_plane_ready
docker tag scheduler:e2e localhost:5001/arbiter.k8s.com.cn/scheduler:latest
registry_load_image localhost:5001/arbiter.k8s.com.cn/scheduler:latest
kind export kubeconfig --name arbiter-e2e --kubeconfig $HOME/.kube/config
install_scheduler_crd
change_default_scheduler_config
remove_master_taint
echo "Local kind cluster k8s:"$K8S_VERSION" is running. scheduler is changed to arbiter scheduler now."
