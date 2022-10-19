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
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	leaseDuration    time.Duration
	renewDeadline    time.Duration
	cacheSyncTimeout time.Duration
)

func CustomizeControllerOptions(options *ctrl.Options) *ctrl.Options {
	options.LeaseDuration = &leaseDuration
	options.RenewDeadline = &renewDeadline
	options.Controller.CacheSyncTimeout = &cacheSyncTimeout
	return options
}

func init() {
	RegisterFlag(func(set *flag.FlagSet) {
		set.DurationVar(&leaseDuration, "lease-duration", time.Second*80, "Lease Duration")
		set.DurationVar(&renewDeadline, "renew-deadline", time.Second*60, "Renew Deadline")
		set.DurationVar(&cacheSyncTimeout, "cache-sync-timeout", time.Minute*10, "Cache Sync Timeout")
	})
}
