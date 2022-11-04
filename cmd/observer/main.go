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
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/dynamic"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	"github.com/kube-arbiter/arbiter/pkg/fetch"
	"github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
	"github.com/kube-arbiter/arbiter/pkg/generated/informers/externalversions"
	"github.com/kube-arbiter/arbiter/pkg/observer"
	"github.com/kube-arbiter/arbiter/pkg/utils"
)

var (
	endpoint     = flag.String("endpoint", "/var/run/observer.sock", "sock file path")
	kubeconfig   = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	fetchTimeout = flag.Int("fetchtimeout", 120, "fetch metric data timeout second")
	namespace    = flag.String("namespace", "", `Which namespaces the operator should watch. 
			The default is to obtain from environment variables, multiple namespaces are separated by commas. default is all namespace`)

	shutdownSignals      = []os.Signal{os.Interrupt, syscall.SIGTERM}
	onlyOneSignalHandler = make(chan struct{})
)

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func main() {
	klog.Infof("Version: %#v\n", version.Get())
	klog.InitFlags(nil)
	flag.Parse()

	stopCh := SetupSignalHandler()

	var (
		cfg *rest.Config
		err error
	)
	if *kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Fatalf("%s building kubeconfig: %s", utils.ErrorLogPrefix, err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("%s building kubernetes clientset: %s", utils.ErrorLogPrefix, err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("%s building kubernetes dynamic clientset error: %s", utils.ErrorLogPrefix, err)
	}

	obiClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("%s building obiservability indicant clientset: %s", utils.ErrorLogPrefix, err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	obiInformerFactory := externalversions.NewSharedInformerFactory(obiClient, time.Second*30)

	conn, err := grpc.Dial(*endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return net.Dial("unix", s)
		}))

	if err != nil {
		klog.Fatalf("%s dial %s error: %s", utils.ErrorLogPrefix, *endpoint, err.Error())
	}
	fetcher, err := fetch.NewFetcher(conn, time.Duration(*fetchTimeout*int(time.Second)))
	if err != nil {
		klog.Fatalf("%s dial %s error: %s", utils.ErrorLogPrefix, *endpoint, err.Error())
	}
	c := observer.NewController(
		*namespace,
		kubeClient,
		dynamicClient,
		obiClient,
		kubeInformerFactory.Core().V1().Pods(),
		obiInformerFactory.Arbiter().V1alpha1().ObservabilityIndicants(),
		fetcher)

	kubeInformerFactory.Start(stopCh)
	obiInformerFactory.Start(stopCh)

	if err = c.Run(stopCh); err != nil {
		klog.Fatalf("%s running controller: %s", utils.ErrorLogPrefix, err.Error())
	}
}
