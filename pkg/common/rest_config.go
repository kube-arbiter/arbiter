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

package common

import (
	"flag"

	"k8s.io/client-go/rest"
)

var (
	apiQPS   float64
	apiBurst int
)

func GetAPIQPS() float32 {
	return float32(apiQPS)
}

func GetAPIBurst() int {
	return apiBurst
}

func CustomizeRestConfig(config *rest.Config) *rest.Config {
	config.QPS = GetAPIQPS()
	config.Burst = GetAPIBurst()
	return config
}

func init() {
	RegisterFlag(func(set *flag.FlagSet) {
		set.Float64Var(&apiQPS, "api-qps", 100, "API QPS")
		set.IntVar(&apiBurst, "api-burst", 100, "API Burst")
	})
}
