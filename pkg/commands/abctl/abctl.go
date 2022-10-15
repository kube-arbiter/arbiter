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

package abctl

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/util/homedir"

	"github.com/kube-arbiter/arbiter/pkg/commands/get"
	"github.com/kube-arbiter/arbiter/pkg/commands/version"
)

type Options struct {
	ConfigFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
	Arguments []string
	Plugins   string
}

var defaultConfigFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)

func NewDefaultCliCommandWitArgs() *cobra.Command {
	op := Options{
		IOStreams:   genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
		Arguments:   os.Args,
		ConfigFlags: defaultConfigFlags,
	}
	return NewDefaultCliCommand(op)
}

func NewDefaultCliCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abctl [flags]",
		Short: `The abctl command is used to make an aggregate presentation of observabilityindicant resources.`,
		Long:  "The abctl command is used to make an aggregate presentation of observabilityindicant resources.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	homeDir := homedir.HomeDir()
	defaultNamespace := "default"
	if homeDir != "" {
		homeDir = filepath.Join(homeDir, ".kube", "config")
	}
	opts.ConfigFlags.KubeConfig = &homeDir
	opts.ConfigFlags.Namespace = &defaultNamespace
	opts.ConfigFlags.AddFlags(cmd.PersistentFlags())

	cmd.AddCommand(get.NewGetCommand(get.Options{IOStreams: opts.IOStreams, ConfigFlags: opts.ConfigFlags}))
	cmd.AddCommand(version.NewVersionCommand())
	return cmd
}
