<h1>
Arbiter: An extendable scheduling and scaling tool for Kubernetes
</h1>

[![License](https://img.shields.io/github/license/kube-arbiter/arbiter.svg?color=4EB1BA&style=flat-square)](https://opensource.org/licenses/Apache-2.0)
[![GitHub release](https://img.shields.io/github/v/release/kube-arbiter/arbiter.svg?style=flat-square)](https://github.com/kube-arbiter/arbiter/releases/latest)
[![CI](https://img.shields.io/github/workflow/status/kube-arbiter/arbiter/e2e?label=CI&logo=github&style=flat-square)](https://github.com/kube-arbiter/arbiter/actions/workflows/e2e.yml)
[![Go Report Card](https://goreportcard.com/badge/kube-arbiter/arbiter?style=flat-square)](https://goreportcard.com/report/github.com/kube-arbiter/arbiter)
[![codecov](https://img.shields.io/codecov/c/github/kube-arbiter/arbiter?logo=codecov&style=flat-square)](https://codecov.io/github/kube-arbiter/arbiter)
[![PRs Welcome](https://badgen.net/badge/PRs/welcome/green?icon=https://api.iconify.design/octicon:git-pull-request.svg?color=white&style=flat-square)](CONTRIBUTING.md)

Arbiter is an extendable scheduling and scaling tool built on Kubernetes. It’ll aggregate various types of data and take them into account when managing, scheduling or scaling the applications in the cluster. It can facilitate K8S users to understand and manage resources deployed in the cluster, and then use this tool to improve the resource utilization and runtime efficiency of the applications.

The input can be metrics from your monitoring system, logs from your logging system or traces from your tracing system. Then arbiter can abstract the data to a unified format，so it can be consumed by scheduler, autoscaler or other controllers. Arbiter can also add labels to your applications based on how the administrator defines the classification and the data it collects. User can understand the characteristics of each application and the overall placement across the cluster, and then tune the factors that can improve it.

### Quick Start

- Refer to [Install and Run Arbiter](http://arbiter.k8s.com.cn/docs/Quick%20Start/install) for quick start.

- Refer to [helm installation](./charts/arbiter/README.md)

## Documentation

See our documentation on [Arbiter Site](http://arbiter.k8s.com.cn) for more details.

## More features

1. Get the cluster metrics for resource usage, such as cluster capacity、system resource usage，reserved resource and actual resource usage
2. Usage trend of various resources
3. Grafana template to show cluster metrics
4. Scheduling based on actual resource usage of node
5. Classify the pod/deployment/statefulset/service based on the resource sensitivity and the baseline defined by the administrator
6. Support grouping of nodes for resource isolation
7. Oversold configuration of pod and node for temporary resource feed

## Contribute to Arbiter

If you want to contribute to Arbiter, refer to [contribute guide](CONTRIBUTING.md).

## Roadmap

You can get what we're doing and plan to do here.

#### v0.1 - done

1. Define what's OBI(Observability Indicator) that can convert metrics/logging/tracing to an unified indicator, so we can use OBI for scheduler and scaler.
2. Introduce a context to let user extend scheduler easily using Javascript, and let user write Javascript to implement schedule logic.
3. Create a plugin framework that can integrate with different observability tools as input and generate OBI output, such as metric-server, prometheus etc...
4. Create a command line tool called 'abctl' to get OBI for K8S resources
5. Create sample use cases using current framework to show how it can work, including:

* Schedule Pod based on actual resource usage on nodes
* Schedule Pod based on node labels using different schedule policy
* Tag the pod/service for different resource sensitivity
* Schedule based on pod and node names/labels without define node affinity
* Reserve node resource in the scheduler extension
* Provide executor to demo how to use OBI to support various use cases.

#### v0.2 - working on now

1. Use kube-scheduler-simulator to evaluate the effect of different schedule policy
2. Support logging indicator in the observer plugins
3. Enhance the executor framework, so the user can easily define actions based on the results of OBI.
4. Allow to use Kubernetes scheduler policy dynamically in scheduler extension
5. Add more samples to generate OBI to introduce popular indicators, so the user can use directly

#### v0.3 - Future plans

1. Support tracing indicator in the observer plugins
2. Integrate with more monitoring/logging/tracing tools for more coverage
3. Improve the timeliness of data for better scheduling or scaling
4. Use time serial database to show the data trend for better visualization
5. Integrate with existing Scheduler, HPA and VPA from open source community, so the user can use it out-of-the-box
6. Add plugin management support to add/remove/update observer/executor/scheduler plugins.

## Support

If you need support, start with the troubleshooting guide, or create github [issues](https://github.com/kube-arbiter/arbiter/issues/new)
