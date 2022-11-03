# Arbiter Helm Chart

## Installation

Quick start to deploy arbiter using Helm.

*Note: This installation setup and configure arbiter-scheduler as [a second scheduler](https://kubernetes.io/docs/tasks/extend-kubernetes/configure-multiple-schedulers/).*

### Prerequisites

- [Helm](https://helm.sh/docs/intro/quickstart/#install-helm)

#### Install chart using Helm v3.0+

```shell
$ git clone https://github.com/kube-arbiter/arbiter;
$ cd arbiter/charts;
# If namespace does not exist.
$ kubectl create ns arbiter-test
$ helm -narbiter-test install arbiter arbiter/
```

#### Verify pods are running properly.

```shell
$ kubectl get po -narbiter-test
NAME                                          READY   STATUS    RESTARTS   AGE
arbiter-scheduler                             1/1     Running   0          60s
executor-74c64d7769-gw7zc                     2/2     Running   0          55s
observer-plugins-75ccf6499f-rvv7x             2/2     Running   0          55s
```

### Configuration

The following table lists the configurable parameters of the as-a-second-scheduler chart and their default values.

| Parameter                                   | Description                                 | Default                                                          |
| ------------------------------------------- | ------------------------------------------- | ---------------------------------------------------------------- |
| `observer.resources`                        | observer controller resources.              | default request cpu is `500m`, default request memory is `64Mi`. |
| `observer.serviceAccountName`               | service account name                        | `observability`                                                  |
| `observer.image.observerImage`              | observer controller image                   | `kubearbiter/observer:v0.1.0`                                    |
| `observer.image.serverImage`                | the observer plugins image                  | `kubearbiter/observer-default-plugins:v0.1.0`                    |
| `observer.image.pullPolicy`                 | image pull policy                           | `IfNotPresent`                                                   |
| `observer.nameOverride`                     | Deployment name                             | `observer-plugins`                                               |
| `observer.address`                          | access address of prometheus in the cluster | `http://prometheus-server.kube-system.svc.cluster.local`         |
| `executor.nameOverride`                     | deployment name for tagging services        | `executor`                                                       |
| `executor.resources`                        | tagging services resources.                 | default request cpu is `10m`, default request memory is `64Mi`   |
| `executor.env`                              | Environment Variables                       | `[]`                                                             |
| `executor.image.resourceTaggingPluginImage` | tagging service image                       | `kubearbiter/resource-tagging-plugin:v0.1.0`                     |
| `executor.image.executorImage`              | execturo controller iamge                   | `kubearbiter/executor:v0.1.0`                                    |
| `executor.image.pullPolicy`                 | image pull policy                           | `IfNotPresent`                                                   |
| `scheduler.nameOverride`                    | scheduler plugin name                       | `arbiter-scheduler`                                              |
| `scheduler.serviceAccountName`              | scheduler plugin serviceAccount name        | `scheduler`                                                      |
| `scheduler.image`                           | image name (don't contain image's tag)      | `kubearbiter/scheduler`                                          |
| `scheduler.tag`                             | image tag                                   | `v0.1.0`                                                         |
| `scheduler.pullPolicy`                      | image pull policy                           | `IfNotPresent`                                                   |
| `scheduler.configMapName`                   | scheduler plugin config content             | `kube-scheduler-configuration-cm`                                |
| `scheduler.resource`                        | scheduler plugin resource                   | default request cpu is `100m`                                    |
| `kubeVersionOverride`                       | kubernetes version                          | Get by default via helm                                          |
