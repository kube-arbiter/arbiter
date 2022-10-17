# Install and Run Arbiter

- [Install and Run Arbiter](#install-and-run-arbiter)
      - [Create or Use existing Kubernetes Cluster](#create-or-use-existing-kubernetes-cluster)
      - [Install and run observer & executor](#install-and-run-observer--executor)
      - [Install and run scheduler](#install-and-run-scheduler)
      - [Uninstall](#uninstall)

#### Create or Use existing Kubernetes Cluster

Firstly you need to have a Kubernetes cluster, and a `kubectl` command-line tool must be configured to communicate with the cluster.

The Kubernetes version must equal to or greater than **v1.20.0**. To check the version, use `kubectl version --short`.

If you do not have a cluster yet, create one by using one of the following provision tools:

* [kind](https://kind.sigs.k8s.io/docs/)
* [kubeadm](https://kubernetes.io/docs/admin/kubeadm/)
* [minikube](https://minikube.sigs.k8s.io/)

#### Install and run observer & executor

1. Clone the code repository

```shell
git clone https://github.com/kube-arbiter/arbiter
```

2. Get required images

Container images are available in the Docker Hub. You can pull and push to your image registry or use them directly if your cluster can access docker hub.

```
docker pull kubearbiter/observer:v0.1.0
docker pull kubearbiter/observer-metric-server:v0.1.0
docker pull kubearbiter/observer-prometheus-server:v0.1.0

docker pull kubearbiter/executor:v0.1.0
docker pull kubearbiter/executor-resource-tagger:v0.1.0

docker pull kubearbiter/scheduler:v0.1.0
```

3. Install resources to run observer and executor with default plugins

```shell
cd arbiter
# create required CRDs
kubectl apply -f manifests/crds
# create namespace to run arbiter
kubectl create ns arbiter-system
# 1. install observer
kubectl -n arbiter-system apply -f manifests/install/observer/observer-rbac.yaml
# Run if use metric-server for metrics service
kubectl -n arbiter-system apply -f manifests/install/observer/observer-metric-server.yaml
# Run if use prometheus for metrics service
kubectl -n arbiter-system apply -f manifests/install/observer/observer-prometheus.yaml
# 2. install executor
kubectl -n arbiter-system apply -f manifests/install/executor
```

4. Run examples

Create a OBI to collect metric data. Here is a sample using the default metric-server of K8S, you probably need to update the value of kind, labels, name, namespace.

Here we use the example from [metric-server-pod-cpu](../manifests/example/observer/metric/metric-server-pod-cpu.yaml)

```yaml
apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: ObservabilityIndicant
metadata:
  name: metric-server-pod-cpu
spec:
  metric:
    metricIntervalSeconds: 30 # interval to pull data
    metrics:
      cpu:
        aggregations:
        - time
        description: cpu
        query: ""
        unit: 'm'
    timeRangeSeconds: 3600 # how long the data point will be collected
  source: metric-server
  targetRef:
    group: ""
    index: 0
    kind: Pod # support pod and node for now
    labels:
      app: mem-cost # match the label of pod/node, can be empty
    name: "" # match the name of pod/node, can be empty
    namespace: arbiter-system # namespace of the pod above, can be empty for node
    version: v1
```

Update manifests/example/observer/metric/metric-server-pod-cpu.yaml following the description above, and run:

```
kubectl apply -f manifests/example/observer/metric/metric-server-pod-cpu.yaml'
```

5. Check results

You can get the obi object and check the data in the status.

```yaml
# kubectl edit obi 
status:
  conditions:
  - lastHeartbeatTime: "2022-10-10T08:12:40Z"
    lastTransitionTime: "2022-10-10T08:12:40Z"
    reason: FetchDataDone
    type: FetchDataDone
  phase: Fetching
  data: # data collected from source
    metrics:
      cpu:
      - endTime: "2022-10-10T08:13:40Z"
        records:
        - timestamp: 1665385980000
          value: "0.034195"
        - timestamp: 1665386040000
          value: "0.033384"
        - timestamp: 1665386100000
          value: "0.034334"
        ...
```

More examples can be found under [observer examples](../manifests/example/observer). You can try to understand how it works better.

#### Install and run scheduler

1. Log into the master node
2. Backup original kube-scheduler.yaml

```
cp /etc/kubernetes/manifests/kube-scheduler.yaml /etc/kubernetes/kube-scheduler-backup.yaml
```

3. Install the scheduler

```
# install scheduler to replace exist one 
kubectl apply -f manifests/install/scheduler/scheduler-rbac.yaml
kubectl apply -f manifests/install/scheduler/scheduler.yaml
kubectl apply -f manifests/install/scheduler/score.yaml

# copy config file or create the file using the content on master node
cp manifests/install/scheduler/kube-scheduler-arbiter-config.yaml to master node path: /etc/kubernetes/kube-scheduler-arbiter-config.yaml
```

Update the kube-scheduler.yaml under /etc/kubernetes/manifests to use arbiter scheduler.

Generally, we need to make a couple of changes:

* pass in the composed scheduler-config file via argument `--config`
* replace the default Kubernetes scheduler image with arbiter scheduler image
* mount the scheduler config file to be readable when scheduler starting

Here is the changes you need to check and update:

```
    - --config=/etc/kubernetes/kube-scheduler-arbiter-config.yaml
    ...
    # use kubearbiter/scheduler:pre-v0.1.0 image if your K8S cluster is less or equal to v1.20
    image: kubearbiter/scheduler:v0.1.0
    volumeMounts:
    - mountPath: /etc/kubernetes/kube-scheduler-arbiter-config.yaml
      name: arbier-scheduler-config
      readOnly: true
    ...
  volumes:
  - hostPath:
      path: /etc/kubernetes/kube-scheduler-arbiter-config.yaml
      type: FileOrCreate
    name: arbier-scheduler-config
```

Make sure the kube-scheduler container is restarted after the file is updated.

4. Run a test and understand how obi affects the kube-scheduler

See [Scheduling by resource usage](http://arbiter.k8s.com.cn/docs/User%20Cases/schedule-by-real-usage)

#### Uninstall

1. Remove Arbiter

   ```
   $ kubectl delete ns arbiter-system
   ```
2. Recover `kube-scheduler.yaml` and delete `kube-scheduler-arbiter-config.yaml`

   - If the cluster is created by `kubeadm` or `minikube`, log into Master node:
     ```
     $ mv /etc/kubernetes/kube-scheduler.yaml /etc/kubernetes/manifests/
     $ rm /etc/kubernetes/kube-scheduler-arbiter-config.yaml
     ```
3. Check state of default scheduler

   ```
   $ kubectl get pod -n kube-system | grep kube-scheduler
   kube-scheduler-xxxx           1/1     Running   0          91s
   ```
