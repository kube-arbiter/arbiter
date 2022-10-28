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

package utils

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var (
	redColor       = []byte{27, 91, 51, 49, 109}
	reset          = []byte{27, 91, 48, 109}
	ErrorLogPrefix = fmt.Sprintf("%s%s%s", redColor, "[Error]", reset)
)

const (
	AddFinalizerErrorMsg = "update finalizer"
	HardGetNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

func ContainFinalizers(finalizers []string, f string) bool {
	if finalizers == nil {
		return false
	}
	for idx := 0; idx < len(finalizers); idx++ {
		if finalizers[idx] == f {
			return true
		}
	}
	return false
}

func IsAddFinalizerError(err error) bool {
	return err.Error() == AddFinalizerErrorMsg
}

func AddFinalizers(finalizers []string, f string) []string {
	for _, finalizer := range finalizers {
		if finalizer == f {
			return finalizers
		}
	}
	return append(finalizers, f)
}

func RemoveFinalizer(finalizers []string, f string) []string {
	idx := 0
	for i := 0; i < len(finalizers); i++ {
		if finalizers[i] == f {
			continue
		}
		finalizers[idx] = finalizers[i]
		idx++
	}
	return finalizers[:idx]
}

func CommonOpt(t, s []string) []string {
	ans := make([]string, 0)
	for ti := 0; ti < len(t); ti++ {
		for si := 0; si < len(s); si++ {
			if t[ti] == s[si] {
				ans = append(ans, t[ti])
			}
		}
	}
	return ans
}

// RenderQuery like helm. Render query using golang's template.
func RenderQuery(templateQuery string, data any, output io.Writer) error {
	templateRender := template.Must(template.New("query").Parse(templateQuery))

	return templateRender.Execute(output, data)
}

func ResourcePlural(client kubernetes.Interface, group, version, kind string) (string, error) {
	gv := schema.GroupVersion{Group: group, Version: version}
	resources, err := client.Discovery().ServerResourcesForGroupVersion(gv.String())
	if err != nil {
		klog.Errorf("%s discovery groupversion error: %s\n", ErrorLogPrefix, err)
		return "", err
	}

	for _, r := range resources.APIResources {
		if r.Kind == kind && !strings.Contains(r.Name, "/") {
			klog.V(4).Infof("Resource %s namespaced: %v\n", r.Name, r.Namespaced)
			return r.Name, nil
		}
	}
	return "", fmt.Errorf("not found resource plural by %s", gv.String())
}

func PodNamespace() string {
	f, err := os.Open(HardGetNamespacePath)
	if err != nil {
		klog.V(4).Infof("%s try get pod namespace by path %s error: %s", ErrorLogPrefix, HardGetNamespacePath, err)
		return ""
	}
	defer f.Close()

	podNamespace, err := io.ReadAll(f)
	if err != nil {
		klog.V(4).Infof("%s try read namespace from %s error: %s", ErrorLogPrefix, HardGetNamespacePath, err)
		return ""
	}

	return string(podNamespace)
}
