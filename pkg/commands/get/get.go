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

package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	"github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
	"github.com/kube-arbiter/arbiter/pkg/printer"
)

type Options struct {
	MetricName  string
	Name        string
	ConfigFlags *genericclioptions.ConfigFlags
	client      versioned.Interface
	genericclioptions.IOStreams
}

func NewGetCommand(option Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get ([NAME | -l label] | NAME ...) [flags]",
		Short: "get command is used to get the obi object",
		Long: `The get command is used to get the obi object. 
By default, the aggregated data of max, min and agv of the obi object will be displayed, using the following example

abctl get [obi-name] -n namespace -m cpu

abctl get -A -m memory -aMAX,MIN`,
		RunE: func(cmd *cobra.Command, args []string) error {
			restCfg, err := option.ConfigFlags.ToRESTConfig()
			if err != nil {
				fmt.Fprint(option.ErrOut, err)
				return err
			}
			option.client, err = versioned.NewForConfig(restCfg)
			if err != nil {
				fmt.Fprint(option.ErrOut, err)
				return err
			}

			allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")
			metricName, _ := cmd.Flags().GetString("metric-name")
			labelSelector, _ := cmd.Flags().GetString("selector")
			namespace, _ := cmd.Flags().GetString("namespace")
			plugins, _ := cmd.Flags().GetString("plugins")
			if len(plugins) > 0 {
				printer.RegisterDefaultFuncs(strings.Split(plugins, ":")...)
			}
			if allNamespaces {
				namespace = ""
			}
			printResources := &v1alpha1.ObservabilityIndicantList{Items: []v1alpha1.ObservabilityIndicant{}}
			if (len(args) == 0 && len(labelSelector) == 0) || len(labelSelector) > 0 {
				printResources, err = option.client.ArbiterV1alpha1().ObservabilityIndicants(namespace).List(context.TODO(), v1.ListOptions{
					LabelSelector: labelSelector,
				})
				if err != nil {
					fmt.Fprint(option.ErrOut, err)
					return err
				}
			}
			if len(args) > 0 {
				for _, name := range args {
					r, err := option.client.ArbiterV1alpha1().ObservabilityIndicants(namespace).Get(context.TODO(), name, v1.GetOptions{})
					if err != nil {
						fmt.Fprint(option.ErrOut, err)
						return err
					}
					printResources.Items = append(printResources.Items, *r)
				}
			}
			aggregationFuncs, _ := cmd.Flags().GetString("aggregations")
			filterFuns := printer.MatchDefaultFunc(strings.Split(aggregationFuncs, ",")...)

			if len(printResources.Items) == 0 {
				which := namespace
				if which == "" {
					which = "all"
				}
				fmt.Fprintf(option.Out, "No resources found in %s namespace.\n", which)
				return nil
			}
			writer := printer.NewDefaultPrinter(&printer.DefaultPrintFlags{
				Funcs:      filterFuns,
				MetricName: metricName,
				Writer:     option.Out,
			})
			writer.Print(printResources)
			return nil
		},
	}
	cmd.Flags().BoolP("all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace")
	cmd.Flags().StringP("plugins", "p", "", "register aggregation functions through go's plugin system. If the same name is used, the function introduced later will overwrite the previous function. E.g: -p=/tmp/range.so:/tmp/up.so")
	cmd.Flags().StringP("metric-name", "m", "", "resource metric name")
	cmd.Flags().StringP("aggregations", "a", "AVG,MAX,MIN", "select the aggregate function to display")
	cmd.Flags().StringP("selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	_ = cmd.MarkFlagRequired("metric-name")
	return cmd
}
