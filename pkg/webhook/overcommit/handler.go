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

package overcommit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kube-arbiter/arbiter/pkg/apis/v1alpha1"
	"github.com/kube-arbiter/arbiter/pkg/generated/clientset/versioned"
)

var (
	OvercommitName string
)

type PatchEntry struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var body []byte

	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expert application/json", strings.ReplaceAll(contentType, "\r", ""))
		return
	}

	reviewResponse := Mutate(r.Context(), body)

	ar := v1.AdmissionReview{
		Response: reviewResponse,
	}

	resp, err := json.Marshal(ar)
	if err != nil {
		klog.Errorf("Cannot encode response: %v", err)
		http.Error(w, fmt.Sprintf("cannot encode response: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Cannot write response: %v", err)
		http.Error(w, fmt.Sprintf("cannot write response: %v", err), http.StatusInternalServerError)
		return
	}
}

func Mutate(ctx context.Context, data []byte) *v1.AdmissionResponse {
	response := &v1.AdmissionResponse{}
	// allow by default
	response.Allowed = true

	klog.V(4).Infof("Body: %s", strings.ReplaceAll(string(data), "\r", ""))
	ar := v1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		klog.Error(err)
		response.Result = &metav1.Status{
			Message: err.Error(),
		}
		return response
	}

	response.UID = ar.Request.UID

	var pod corev1.Pod
	if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
		klog.Errorf("Cannot unmarshal raw object: %v", err)
		klog.Errorf("Object: %v", strings.ReplaceAll(string(ar.Request.Object.Raw), "\r", ""))
		response.Result = &metav1.Status{
			Message: err.Error(),
		}

		return response
	}

	cfg, err := config.GetConfig()
	if err != nil {
		klog.Errorf("Cannot load config: %v", err)
		response.Result = &metav1.Status{
			Message: err.Error(),
		}
		return response
	}

	cs, _ := versioned.NewForConfig(cfg)
	overCommit, err := cs.ArbiterV1alpha1().OverCommits().Get(ctx, OvercommitName, metav1.GetOptions{})

	if err != nil {
		klog.Errorf("Get overcommit resource error: %v", err)
		response.Result = &metav1.Status{
			Message: err.Error(),
		}
		return response
	}

	patch := PodPatch(&pod, overCommit)
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		klog.Errorf("Error marshaling pod update: %v", err)
		return &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	response.Patch = patchBytes
	response.PatchType = func() *v1.PatchType { pt := v1.PatchTypeJSONPatch; return &pt }()

	return response
}

func PodPatch(pod *corev1.Pod, overCommit *v1alpha1.OverCommit) []PatchEntry {
	ratioCPU, _ := strconv.ParseFloat(overCommit.Spec.CPU.Ratio, 64)
	ratioMemory, _ := strconv.ParseFloat(overCommit.Spec.Memory.Ratio, 64)

	initContainers := patchContainers(pod.Spec.InitContainers, overCommit, ratioCPU, ratioMemory)
	containers := patchContainers(pod.Spec.Containers, overCommit, ratioCPU, ratioMemory)
	patch := make([]PatchEntry, 0)

	if len(initContainers) > 0 {
		patch = append(patch, PatchEntry{
			Op:    "replace",
			Path:  "/spec/initContainers",
			Value: initContainers,
		})
	}

	if len(containers) > 0 {
		patch = append(patch, PatchEntry{
			Op:    "replace",
			Path:  "/spec/containers",
			Value: containers,
		})
	}

	return patch
}

// Patch the containers using overcommit configuration
func patchContainers(oldContainers []corev1.Container, overCommit *v1alpha1.OverCommit, ratioCPU, ratioMemory float64) []corev1.Container {
	// Get the cpu/memory limit from overcommit if it has
	defaultLimitCPU := resource.MustParse("2")
	if overCommit.Spec.CPU.Limit != "" {
		defaultLimitCPU = resource.MustParse(overCommit.Spec.CPU.Limit)
	}
	defaultLimitMemory := resource.MustParse("4Gi")
	if overCommit.Spec.Memory.Limit != "" {
		defaultLimitMemory = resource.MustParse(overCommit.Spec.Memory.Limit)
	}
	containers := make([]corev1.Container, 0)
	for _, c := range oldContainers {
		// Use default limits if the original one is nil and overcommit has the limit configured
		if c.Resources.Limits == nil && overCommit.Spec.CPU.Limit != "" {
			c.Resources.Limits[corev1.ResourceCPU] = defaultLimitCPU
		}
		if c.Resources.Limits == nil && overCommit.Spec.Memory.Limit != "" {
			c.Resources.Limits[corev1.ResourceMemory] = defaultLimitMemory
		}
		// Calcuate the cpu/memory requests based on the limits
		cpuLimit := c.Resources.Limits[corev1.ResourceCPU]
		memoryLimit := c.Resources.Limits[corev1.ResourceMemory]
		cpuRequest := c.Resources.Requests[corev1.ResourceCPU]
		memoryRequest := c.Resources.Requests[corev1.ResourceMemory]

		// Allow user to define lower request than the configured overcommit value
		overcommitCPURequest := cpuRequest
		if float64(cpuRequest.MilliValue()) > float64(cpuLimit.MilliValue())*ratioCPU {
			overcommitCPURequest = resource.MustParse(fmt.Sprintf("%vm", int64(float64(cpuLimit.MilliValue())*ratioCPU)))
		}
		overcommitMemoryRequest := memoryRequest
		if float64(memoryRequest.MilliValue()) > float64(memoryLimit.MilliValue())*ratioMemory {
			overcommitMemoryRequest = resource.MustParse(fmt.Sprintf("%vMi", int64(float64(memoryLimit.Value())*ratioMemory)/1024/1024))
		}
		c.Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    overcommitCPURequest,
			corev1.ResourceMemory: overcommitMemoryRequest,
		}

		containers = append(containers, c)
	}
	return containers
}
