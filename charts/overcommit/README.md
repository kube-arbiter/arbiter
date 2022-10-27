# Webhook Helm Chart
*Note: You should deploy [cert-manager](https://cert-manager.io/docs/installation/) before install the chart
## Installation

Quick start to deploy webhook using Helm.

*Note: This installation setup and configure overcommit webhook

### Prerequisites

- [Helm](https://helm.sh/docs/intro/quickstart/#install-helm)

#### Install chart using Helm v3.0+

```shell
$ git clone https://github.com/kube-arbiter/arbiter;
$ cd arbiter/charts;
# If namespace does not exist.
$ kubectl create ns overcommit-test
$ helm -novercommit-test install overcommit-webhook overcommit/
```

#### Verify resource are running properly.

```shell
kubectl get po -novercommit-test
NAME                                  READY   STATUS    RESTARTS   AGE
overcommit-webhook-86877fbf9c-tmjlx   1/1     Running   0          11m
```

```shell
kubectl get overcommit
NAME         AGE
overcommit   6d15h
```
```shell
 kubectl get overcommit overcommit -oyaml
apiVersion: arbiter.k8s.com.cn/v1alpha1
kind: OverCommit
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"arbiter.k8s.com.cn/v1alpha1","kind":"OverCommit","metadata":{"annotations":{},"name":"overcommit"},"spec":{"cpu":{"limit":"100m","ratio":"0.5"},"memory":{"limit":"1024Mi","ratio":"0.6"},"namespaceList":["default"]}}
  creationTimestamp: "2022-11-02T12:10:21Z"
  generation: 1
  name: overcommit
  resourceVersion: "1816"
  uid: e082f9a5-d939-466e-a51d-342632bfa805
spec:
  cpu:
    limit: 100m
    ratio: "0.5"
  memory:
    limit: 1024Mi
    ratio: "0.6"
  namespaceList:
  - default

```

### Configuration

The following table lists the configurable parameters of overcommit

| field name       | desc         | value |
| ------------ | ------------ | ------ |
| memory.ratio | memory ratio | 0.5    |
| cpu.ratio    | cpu ratio    | 0.5    |
| memory.limit | memory limit   | 500Mi  |
| cpu.limit    | cpu limit      | 100m   |
| namespaceList   | effected namespaces  | array |
