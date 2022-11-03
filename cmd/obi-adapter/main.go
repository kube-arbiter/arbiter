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

package main

import (
	"flag"
	"fmt"
	"os"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"

	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	customexternalmetrics "sigs.k8s.io/custom-metrics-apiserver/pkg/apiserver"
	basecmd "sigs.k8s.io/custom-metrics-apiserver/pkg/cmd"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/metrics-server/pkg/api"

	emProvider "github.com/kube-arbiter/arbiter/pkg/adapter/provider"
	generatedopenapi "github.com/kube-arbiter/arbiter/pkg/generated/api/generated/openapi"
	"github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
)

type ArbiterAdapter struct {
	basecmd.AdapterBase

	// AdapterConfigFile points to the file containing the arbiter configuration.
	AdapterConfigFile string
	/*// MetricsRelistInterval is the interval at which to relist the set of available metrics
	MetricsRelistInterval time.Duration
	// MetricsMaxAge is the period to query available metrics for
	MetricsMaxAge time.Duration*/
}

func (cmd *ArbiterAdapter) addFlags() {
	cmd.Flags().StringVar(&cmd.AdapterConfigFile, "config", cmd.AdapterConfigFile,
		"Configuration file containing details of how to transform between OBI metrics "+
			"and custom metrics API resources")
}

func (cmd *ArbiterAdapter) makeProvider(stopCh <-chan struct{}) (provider.ExternalMetricsProvider, error) {
	mapper, err := cmd.RESTMapper()
	if err != nil {
		return nil, fmt.Errorf("unable to construct RESTMapper: %v", err)
	}
	// Kubernetes client
	restConfig, err := cmd.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to construct Kubernetes rest client: %v", err)
	}
	kubeClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct Kubernetes versioned client: %v", err)
	}
	// construct the provider and start it
	emProvider := emProvider.NewArbiterProvider(mapper, kubeClient)
	return emProvider, nil
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	// set up flags
	cmd := &ArbiterAdapter{}
	cmd.Name = "obi-metrics-adapter"

	cmd.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(generatedopenapi.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(api.Scheme, customexternalmetrics.Scheme))
	cmd.OpenAPIConfig.Info.Title = "obi-metrics-adapter"
	cmd.OpenAPIConfig.Info.Version = "0.2.0"

	cmd.addFlags()
	// make sure we get klog flags
	local := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	logs.AddGoFlags(local)
	cmd.Flags().AddGoFlagSet(local)
	if err := cmd.Flags().Parse(os.Args); err != nil {
		klog.Fatalf("unable to parse flags: %v", err)
	}

	// stop channel closed on SIGTERM and SIGINT
	stopCh := genericapiserver.SetupSignalHandler()

	// construct the provider
	emProvider, err := cmd.makeProvider(stopCh)
	if err != nil {
		klog.Fatalf("unable to construct custom metrics provider: %v", err)
	}

	// attach the provider to the server, if it's needed
	if emProvider != nil {
		cmd.WithExternalMetrics(emProvider)
	}
	// run the server
	if err := cmd.Run(stopCh); err != nil {
		klog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}
