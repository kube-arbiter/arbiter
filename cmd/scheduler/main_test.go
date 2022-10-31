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
	"net"
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
	"k8s.io/kubernetes/pkg/scheduler/apis/config/testing/defaults"

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
apiVersion: kubescheduler.config.k8s.io/v1beta2
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

	testcases := []struct {
		name            string
		flags           []string
		registryOptions []app.Option
		wantPlugins     map[string]*kubeschedulerconfig.Plugins
	}{
		{
			name: "default config",
			flags: []string{
				"--kubeconfig", configKubeconfig,
			},
			wantPlugins: map[string]*kubeschedulerconfig.Plugins{
				"default-scheduler": defaults.ExpandedPluginsV1beta3,
			},
		},
		{
			name:            "single profile config - Arbiter",
			flags:           []string{"--config", ExtendProfilesConfig},
			registryOptions: []app.Option{app.WithPlugin(scheduler.Name, scheduler.New)},
			wantPlugins: map[string]*kubeschedulerconfig.Plugins{
				"default-scheduler": {
					QueueSort:  defaults.PluginsV1beta2.QueueSort,
					PreFilter:  defaults.PluginsV1beta2.PreFilter,
					Filter:     defaults.PluginsV1beta2.Filter,
					Bind:       defaults.PluginsV1beta2.Bind,
					PostFilter: defaults.PluginsV1beta2.PostFilter,
					Score:      kubeschedulerconfig.PluginSet{Enabled: append(defaults.PluginsV1beta2.Score.Enabled, kubeschedulerconfig.Plugin{Name: scheduler.Name, Weight: 100})},
					PreScore:   defaults.PluginsV1beta2.PreScore,
					Reserve:    defaults.PluginsV1beta2.Reserve,
					PreBind:    defaults.PluginsV1beta2.PreBind,
				},
			},
		},
	}

	makeListener := func(t *testing.T) net.Listener {
		t.Helper()
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Fatal(err)
		}
		return l
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.PanicOnError)
			opts := options.NewOptions()

			nfs := opts.Flags
			for _, f := range nfs.FlagSets {
				fs.AddFlagSet(f)
			}
			if err := fs.Parse(tc.flags); err != nil {
				t.Fatal(err)
			}

			// use listeners instead of static ports so parallel test runs don't conflict
			opts.SecureServing.Listener = makeListener(t)
			defer opts.SecureServing.Listener.Close()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_, sched, err := app.Setup(ctx, opts, tc.registryOptions...)
			if err != nil {
				t.Fatal(err)
			}

			gotPlugins := make(map[string]*kubeschedulerconfig.Plugins)
			for n, p := range sched.Profiles {
				gotPlugins[n] = p.ListPlugins()
			}

			if diff := cmp.Diff(tc.wantPlugins, gotPlugins); diff != "" {
				t.Errorf("unexpected plugins diff (-want, +got): %s", diff)
			}
		})
	}
}
