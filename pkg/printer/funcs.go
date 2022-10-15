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
	"log"
	"plugin"
	"strconv"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
)

type AggregateFunction func(v1alpha1.ObservabilityIndicantStatusMetricInfo) string

const Precision = 0.000001 // equal

var defaultFuncs = map[string]AggregateFunction{
	"MAX": Max,
	"MIN": Min,
	"AVG": Avg,
}

func RegisterDefaultFuncs(plugins ...string) {
	for _, soFile := range plugins {
		p, err := plugin.Open(soFile)
		if err != nil {
			log.Fatal(err)
		}
		symbol, err := p.Lookup("Funcs")
		if err != nil {
			log.Fatal(err)
		}
		mapFunc, ok := symbol.(*map[string]AggregateFunction)
		if !ok {
			return
		}

		for name, fn := range *mapFunc {
			defaultFuncs[name] = fn
		}
	}
}

func MatchDefaultFunc(args ...string) map[string]AggregateFunction {
	ans := make(map[string]AggregateFunction)
	for _, arg := range args {
		if fn, ok := defaultFuncs[arg]; ok {
			ans[arg] = fn
		}
	}
	return ans
}

func Max(arg v1alpha1.ObservabilityIndicantStatusMetricInfo) string {
	if len(arg.Records) == 0 {
		return none
	}

	max, _ := strconv.ParseFloat(arg.Records[0].Value, 64)
	for idx := 1; idx < len(arg.Records); idx++ {
		f, _ := strconv.ParseFloat(arg.Records[idx].Value, 64)
		if f > max {
			max = f
		}
	}
	return fmt.Sprintf("%.1f%s", max, arg.Unit)
}

func Min(arg v1alpha1.ObservabilityIndicantStatusMetricInfo) string {
	if len(arg.Records) == 0 {
		return none
	}

	min, _ := strconv.ParseFloat(arg.Records[0].Value, 64)
	for idx := 1; idx < len(arg.Records); idx++ {
		f, _ := strconv.ParseFloat(arg.Records[idx].Value, 64)
		if f < min {
			min = f
		}
	}
	return fmt.Sprintf("%.1f%s", min, arg.Unit)
}

func Avg(arg v1alpha1.ObservabilityIndicantStatusMetricInfo) string {
	if len(arg.Records) == 0 {
		return none
	}

	avg := float64(0)
	for _, v := range arg.Records {
		f, _ := strconv.ParseFloat(v.Value, 64)
		avg += f
	}
	return fmt.Sprintf("%.1f%s", avg/float64(len(arg.Records)), arg.Unit)
}
