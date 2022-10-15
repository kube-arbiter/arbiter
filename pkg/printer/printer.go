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

package printer

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
)

type DefaultPrintFlags struct {
	Funcs      map[string]AggregateFunction
	MetricName string
	Writer     io.Writer
}

type defaultPrinter struct {
	// funcs
	funcs map[string]AggregateFunction

	// metricName cpu, memory eg.
	metricName string

	headers []string
	writer  io.Writer
}

// first version only support fields. no extension.
// METRICNAME like cpu, memory eg.
var fixedHeaders = []string{"NAME", "NAMESPACE", "KIND", "TARGETNAME", "METRICNAME"}

const (
	tab  = "\t"
	none = "<none>"
)

type Printer interface {
	// Print this command line tool is only used for this crd, so I won't do too much abstraction here.
	Print(*v1alpha1.ObservabilityIndicantList)
}

var _ Printer = &defaultPrinter{}

func NewDefaultPrinter(ops *DefaultPrintFlags) Printer {
	dp := &defaultPrinter{
		funcs:      make(map[string]AggregateFunction),
		headers:    make([]string, len(fixedHeaders)), // 固定是5
		metricName: ops.MetricName,
		writer:     os.Stdout,
	}
	copy(dp.headers, fixedHeaders)
	if ops.Writer != nil {
		dp.writer = ops.Writer
	}
	if len(ops.Funcs) > 0 {
		upFnName := make([]string, 0)
		for k, f := range ops.Funcs {
			dp.funcs[k] = f
			upFnName = append(upFnName, strings.ToUpper(k))
		}
		sort.Strings(upFnName)
		dp.headers = append(dp.headers, upFnName...)
	} else {
		dp.funcs["AVG"], dp.funcs["MAX"], dp.funcs["MIN"] = Avg, Max, Min
		dp.headers = append(dp.headers, "AVG", "MAX", "MIN")
	}

	return dp
}

func (dp *defaultPrinter) Print(instanceList *v1alpha1.ObservabilityIndicantList) {
	w := tabwriter.NewWriter(dp.writer, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, strings.Join(dp.headers, tab))

	for _, instance := range instanceList.Items {
		metrics := instance.Status.Metrics
		if metrics == nil {
			continue
		}

		datas := make([]string, len(dp.headers))
		datas[0], datas[1], datas[2], datas[3] = instance.GetName(), instance.GetNamespace(),
			instance.Spec.TargetRef.Kind, instance.Spec.TargetRef.Name
		datas[4] = strings.ToUpper(dp.metricName)

		if instance.Spec.TargetRef.Name == "" {
			datas[3] = none
			metrics := instance.Status.Metrics
			if len(metrics) != 0 {
				for _, v := range metrics {
					if len(v) != 0 {
						datas[3] = v[0].TargetItem
						break
					}
				}
			}
		}

		for idx := 5; idx < len(dp.headers); idx++ {
			datas[idx] = none
		}
		if metricValues, ok := metrics[dp.metricName]; ok {
			for idx := 5; idx < len(dp.headers); idx++ {
				if fn := dp.funcs[dp.headers[idx]]; fn != nil {
					datas[idx] = fn(metricValues[0])
				}
			}
		}
		fmt.Fprintln(w, strings.Join(datas, tab))
	}
	w.Flush()
}
