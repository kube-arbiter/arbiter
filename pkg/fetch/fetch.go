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
	FetchOK        FetchErrorType = 0
	FetchCtxDone   FetchErrorType = 1
	FetchTimeOut   FetchErrorType = 2
	FetchRPC       FetchErrorType = 3
	FetchDoNothing FetchErrorType = 4
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
	GetPluginName() string
	Health() bool
}

type fetcher struct {
	rpcCli        obi.ServerClient
	timeout       time.Duration
	capabilities  map[string]*obi.MetricInfo
	outPluginName string
}

// verify fetcher impl Fetcher
var _ Fetcher = &fetcher{}

func NewFetcher(conn *grpc.ClientConn, fetchTimeout time.Duration) Fetcher {
	f := &fetcher{rpcCli: obi.NewServerClient(conn), timeout: fetchTimeout}
	capabilitiers, err := f.rpcCli.PluginCapabilities(context.Background(), &obi.PluginCapabilitiesRequest{})
	if err != nil {
		klog.Infof("NewFetcher: try get capabilities error: %s\n", err)
		return f
	}

	f.capabilities = capabilitiers.MetricInfo
	r, err := f.rpcCli.GetPluginName(context.Background(), &obi.GetPluginNameRequest{})
	if err != nil {
		klog.Infof("NewFetcher: try get plugin name error: %s\n", err)
		return f
	}
	klog.Infof("NewFetcher plugin name: %s\n", r.Name)
	f.outPluginName = r.Name
	return f
}

func (f *fetcher) Health() bool {
	return true
}

func (f *fetcher) GetPluginName() string {
	return f.outPluginName
}

func (f *fetcher) Fetch(ctx context.Context, req *obi.GetMetricsRequest) (*v1alpha1.ObservabilityIndicantStatusMetricInfo, FetchErrorType, error) {
	method := "fetcher.Fetch"

	if len(req.ResourceNames) == 0 {
		return nil, FetchDoNothing, nil
	}

	klog.V(5).Infof("Starting to fetch metrics for resource: %s\n", req.ResourceNames)
	fetchResult := make(chan error)

	ans := &v1alpha1.ObservabilityIndicantStatusMetricInfo{}

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
