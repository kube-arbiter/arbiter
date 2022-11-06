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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"k8s.io/kubernetes/cmd/kube-scheduler/app/options"
	kubeschedulerconfig "k8s.io/kubernetes/pkg/scheduler/apis/config"

	"github.com/kube-arbiter/arbiter/pkg/scheduler"
)

func TestSetup(t *testing.T) {
	// temp dir
	tmpDir, err := os.MkdirTemp("", "scheduler")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// https server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"metadata": {"name": "test"}}`))
	}))
	defer server.Close()

	configKubeconfig := filepath.Join(tmpDir, "config.kubeconfig")
	if err := os.WriteFile(configKubeconfig, []byte(fmt.Sprintf(`
apiVersion: v1
kind: Config
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: %s
  name: default
contexts:
- context:
    cluster: default
    user: default
  name: default
current-context: default
users:
- name: default
  user:
    username: config
`, server.URL)), os.FileMode(0600)); err != nil {
		t.Fatal(err)
	}

	ExtendProfilesConfig := filepath.Join(tmpDir, "KubeSchedulerConfiguration.yaml")
	if err := os.WriteFile(ExtendProfilesConfig, []byte(fmt.Sprintf(`
apiVersion: kubescheduler.config.k8s.io/v1beta1
kind: KubeSchedulerConfiguration
clientConnection:
  kubeconfig: "%s"
profiles:
  - schedulerName: default-scheduler
    plugins:
      score:
        enabled:
          - name: Arbiter
            weight: 100
    pluginConfig:
    - name: Arbiter
      args:
        kubeConfigPath: "%s"
`, configKubeconfig, configKubeconfig)), os.FileMode(0600)); err != nil {
		t.Fatal(err)
	}

	defaultPlugins := map[string][]kubeschedulerconfig.Plugin{
		"QueueSortPlugin": {
			{Name: "PrioritySort"},
		},
		"PreFilterPlugin": {
			{Name: "NodeResourcesFit"},
			{Name: "NodePorts"},
			{Name: "PodTopologySpread"},
			{Name: "InterPodAffinity"},
			{Name: "VolumeBinding"},
		},
		"FilterPlugin": {
			{Name: "NodeUnschedulable"},
			{Name: "NodeName"},
			{Name: "TaintToleration"},
			{Name: "NodeAffinity"},
			{Name: "NodePorts"},
			{Name: "NodeResourcesFit"},
			{Name: "VolumeRestrictions"},
			{Name: "EBSLimits"},
			{Name: "GCEPDLimits"},
			{Name: "NodeVolumeLimits"},
			{Name: "AzureDiskLimits"},
			{Name: "VolumeBinding"},
			{Name: "VolumeZone"},
			{Name: "PodTopologySpread"},
			{Name: "InterPodAffinity"},
		},
		"PostFilterPlugin": {
			{Name: "DefaultPreemption"},
		},
		"PreScorePlugin": {
			{Name: "InterPodAffinity"},
			{Name: "PodTopologySpread"},
			{Name: "TaintToleration"},
		},
		"ScorePlugin": {
			{Name: "NodeResourcesBalancedAllocation", Weight: 1},
			{Name: "ImageLocality", Weight: 1},
			{Name: "InterPodAffinity", Weight: 1},
			{Name: "NodeResourcesLeastAllocated", Weight: 1},
			{Name: "NodeAffinity", Weight: 1},
			{Name: "NodePreferAvoidPods", Weight: 10000},
			{Name: "PodTopologySpread", Weight: 2},
			{Name: "TaintToleration", Weight: 1},
		},
		"BindPlugin":    {{Name: "DefaultBinder"}},
		"ReservePlugin": {{Name: "VolumeBinding"}},
		"PreBindPlugin": {{Name: "VolumeBinding"}},
	}

	testcases := []struct {
		name            string
		flags           []string
		registryOptions []app.Option
		wantPlugins     map[string]map[string][]kubeschedulerconfig.Plugin
	}{
		{
			name: "default config",
			flags: []string{
				"--kubeconfig", configKubeconfig,
			},
			wantPlugins: map[string]map[string][]kubeschedulerconfig.Plugin{
				"default-scheduler": defaultPlugins,
			},
		},
		{
			name:            "single profile config - Arbiter",
			flags:           []string{"--config", ExtendProfilesConfig},
			registryOptions: []app.Option{app.WithPlugin(scheduler.Name, scheduler.New)},
			wantPlugins: map[string]map[string][]kubeschedulerconfig.Plugin{
				"default-scheduler": {
					"QueueSortPlugin":  defaultPlugins["QueueSortPlugin"],
					"BindPlugin":       {{Name: "DefaultBinder"}},
					"PostFilterPlugin": {{Name: "DefaultPreemption"}},
					"ScorePlugin": {
						{Name: "NodeResourcesBalancedAllocation", Weight: 1},
						{Name: "ImageLocality", Weight: 1},
						{Name: "InterPodAffinity", Weight: 1},
						{Name: "NodeResourcesLeastAllocated", Weight: 1},
						{Name: "NodeAffinity", Weight: 1},
						{Name: "NodePreferAvoidPods", Weight: 10000},
						{Name: "PodTopologySpread", Weight: 2},
						{Name: "TaintToleration", Weight: 1},
						{Name: "Arbiter", Weight: 100},
					},
					"ReservePlugin": {{Name: "VolumeBinding"}},
					"PreBindPlugin": {{Name: "VolumeBinding"}},
					"FilterPlugin": {
						{Name: "NodeUnschedulable"}, {Name: "NodeName"}, {Name: "TaintToleration"},
						{Name: "NodeAffinity"}, {Name: "NodePorts"}, {Name: "NodeResourcesFit"},
						{Name: "VolumeRestrictions"}, {Name: "EBSLimits"}, {Name: "GCEPDLimits"},
						{Name: "NodeVolumeLimits"}, {Name: "AzureDiskLimits"}, {Name: "VolumeBinding"},
						{Name: "VolumeZone"}, {Name: "PodTopologySpread"}, {Name: "InterPodAffinity"},
					},
					"PreFilterPlugin": {
						{Name: "NodeResourcesFit"}, {Name: "NodePorts"}, {Name: "PodTopologySpread"},
						{Name: "InterPodAffinity"}, {Name: "VolumeBinding"},
					},
					"PreScorePlugin": {
						{Name: "InterPodAffinity"}, {Name: "PodTopologySpread"},
						{Name: "TaintToleration"},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			opts, err := options.NewOptions()
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range opts.Flags().FlagSets {
				fs.AddFlagSet(f)
			}
			if err := fs.Parse(tc.flags); err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cc, sched, err := app.Setup(ctx, opts, tc.registryOptions...)
			if err != nil {
				t.Fatal(err)
			}
			defer cc.SecureServing.Listener.Close()
			defer cc.InsecureServing.Listener.Close()

			gotPlugins := make(map[string]map[string][]kubeschedulerconfig.Plugin)
			for n, p := range sched.Profiles {
				gotPlugins[n] = p.ListPlugins()
			}

			if diff := cmp.Diff(tc.wantPlugins, gotPlugins); diff != "" {
				t.Errorf("unexpected plugins diff (-want, +got): %s", diff)
			}
		})
	}
}
