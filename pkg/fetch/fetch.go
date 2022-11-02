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

package fetch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	obi "github.com/kube-arbiter/arbiter/pkg/proto/lib/observer"
)

type FetchErrorType uint8

const (
	FetchOK                  FetchErrorType = 0
	FetchCtxDone             FetchErrorType = 1
	FetchTimeOut             FetchErrorType = 2
	FetchRPC                 FetchErrorType = 3
	FetchDoNothing           FetchErrorType = 4
	MaximumConnectRetryCount                = 10
)

type Resp struct {
	Unit        string
	Aggregation map[string]float64
}

type Request struct {
	// metric, log or trace
	From          string
	Req           *obi.GetMetricsRequest
	MetricName    string
	ResourceNames ResourceNameWithKind // kind
	Namespace     string
	Start, End    int64

	Query        string
	Unit         string
	Aggregations []string
}

type ResourceNameWithKind struct {
	Names []string
	Kind  string
}

func (r ResourceNameWithKind) String() string {
	return fmt.Sprintf("%s/%s", r.Kind, strings.Join(r.Names, ","))
}

type Fetcher interface {
	Fetch(context.Context, *obi.GetMetricsRequest) (*v1alpha1.ObservabilityIndicantStatusMetricInfo, FetchErrorType, error)
	GetPluginNames() map[string]struct{}
	Health() bool
}

type fetcher struct {
	rpcCli         obi.ServerClient
	timeout        time.Duration
	capabilities   map[string]*obi.PluginCapability
	outPluginNames []string
}

// verify fetcher impl Fetcher
var _ Fetcher = &fetcher{}

func NewFetcher(conn *grpc.ClientConn, fetchTimeout time.Duration) (Fetcher, error) {
	f := &fetcher{rpcCli: obi.NewServerClient(conn), timeout: fetchTimeout}
	// try to connect to observer plugin
	var capabilitiers *obi.PluginCapabilitiesResponse
	var err error
	for try := 0; try < MaximumConnectRetryCount; try++ {
		capabilitiers, err = f.rpcCli.PluginCapabilities(context.Background(), &obi.PluginCapabilitiesRequest{})
		if err == nil {
			break
		}
		klog.Errorf("NewFetcher: try to get capabilities error: %s\n", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return f, err
	}

	f.capabilities = capabilitiers.Capabilities
	r, err := f.rpcCli.GetPluginNames(context.Background(), &obi.GetPluginNameRequest{})
	if err != nil {
		klog.Infof("NewFetcher: try to get plugin name error: %s\n", err)
		return f, err
	}
	klog.Infof("NewFetcher plugin name: %v\n", r.Names)
	f.outPluginNames = r.Names
	return f, nil
}

func (f *fetcher) Health() bool {
	return true
}

func (f *fetcher) GetPluginNames() map[string]struct{} {
	pluginNames := make(map[string]struct{})
	for _, name := range f.outPluginNames {
		pluginNames[name] = struct{}{}
	}

	return pluginNames
}

func (f *fetcher) Fetch(ctx context.Context, req *obi.GetMetricsRequest) (*v1alpha1.ObservabilityIndicantStatusMetricInfo, FetchErrorType, error) {
	method := "fetcher.Fetch"

	klog.V(5).Infof("Starting to fetch metrics for resource: %s\n", req.ResourceNames)

	// If the logic does not go to get data from the fetchResult channel, it will cause a goroutine leak.
	// So unbuffered channels cannot be used here
	fetchResult := make(chan error, 1)

	ans := &v1alpha1.ObservabilityIndicantStatusMetricInfo{}
	for idx, pluginName := range f.outPluginNames {
		if req.Source == pluginName {
			break
		}
		if idx == len(f.outPluginNames)-1 {
			return ans, FetchDoNothing, fmt.Errorf("plugin %s not found", req.Source)
		}
	}

	go func(ans *v1alpha1.ObservabilityIndicantStatusMetricInfo) {
		metricResp, err := f.rpcCli.GetMetrics(ctx, req)
		if err != nil {
			klog.Errorf("\t %s try to get metric by resource(%v) or by query(%s) error: %s\n", method, req.ResourceNames, req.Query, err)
			fetchResult <- err
			return
		}
		if metricResp.ResourceName != "" {
			// such as pod name or node name
			// When the targetRef field supports label selectors, there will be different expressions
			// But now it is thought that those that support the query field will not return the resource_name field.
			ans.TargetItem = metricResp.ResourceName
		}

		ans.Unit = req.Unit
		if metricResp.Unit != "" {
			ans.Unit = metricResp.Unit
		}

		ans.Records = make([]v1alpha1.Record, len(metricResp.Records))
		for idx, record := range metricResp.Records {
			ans.Records[idx] = v1alpha1.Record{Timestamp: record.Timestamp, Value: record.Value}
		}

		ans.StartTime = metav1.NewTime(time.Unix(0, req.StartTime*int64(time.Millisecond)))
		ans.EndTime = metav1.NewTime(time.Unix(0, req.EndTime*int64(time.Millisecond)))

		fetchResult <- nil
	}(ans)

	select {
	case <-ctx.Done():
		return ans, FetchCtxDone, fmt.Errorf("the fetch context is broken")
	case <-time.After(f.timeout):
		return ans, FetchTimeOut, fmt.Errorf("fetch context timeout")
	case fetchErr := <-fetchResult:
		et := FetchOK
		if fetchErr != nil {
			et = FetchRPC
		}
		return ans, et, fetchErr
	}
}
