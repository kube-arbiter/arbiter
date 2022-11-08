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

package e2e_test

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	DefaultNamespace = "default"
	Kubectl          = "kubectl"
)

func CreateByYaml(yaml string, timeoutSecond int) (outStr string, err error) {
	args := []string{"apply", "-f", "-"}
	return BaseCmd("", yaml, timeoutSecond, args...)
}

func DeleteByYaml(yaml string, timeoutSecond int) (outStr string, err error) {
	args := []string{"delete", "-f", "-"}
	return BaseCmd("", yaml, timeoutSecond, args...)
}

func DeletePod(deployName, namespace, label string, timeoutSecond int, wait bool) (ok bool, err error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	if label == "" {
		label = "app=" + deployName
	}
	args := []string{"delete", "pod", "-n", namespace, "-l", label}
	if wait {
		args = append(args, "--wait=true")
	} else {
		args = append(args, "--force")
	}
	_, err = BaseCmd("", "", timeoutSecond, args...)
	if err != nil {
		return false, fmt.Errorf("delete pod[%s] error:[%s]", strings.Join(args, " "), err)
	}
	return true, nil
}

func GetPodNodeName(deployName, namespace, label string, timeoutSecond int) (nodeName string, err error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	if label == "" {
		label = "app=" + deployName
	}
	// 1. wait deploy progress (create or rollout) done.
	args := []string{"wait", "--for=condition=Progressing", fmt.Sprintf("deployment/%s", deployName)}
	if _, err = BaseCmd("", "", timeoutSecond, args...); err != nil {
		return "", err
	}
	// 2. wait deploy available (can serve) done.
	args = []string{"wait", "--for=condition=Available", fmt.Sprintf("deployment/%s", deployName)}
	if _, err = BaseCmd("", "", timeoutSecond, args...); err != nil {
		return "", err
	}
	// 3. get pod node name
	args = []string{"get", "pod", "-n", namespace, "-l", label, "-o", `jsonpath="{.items[*].spec.nodeName}"`}
	return BaseCmd("", "", timeoutSecond, args...)
}

func GetNodeNameByLabel(label string, timeoutSecond int) (nodeName string, err error) {
	return GetByJSONPath("node", "", label, "{.items[*].metadata.name}", timeoutSecond)
}

func GetOBIRecords(namespace, label string, timeoutSecond int) (nodeName string, err error) {
	return GetByJSONPath("obi", namespace, label, "{.items[*].status.metrics.*[0].records}", timeoutSecond)
}

func GetByJSONPath(resource, namespace, label, jsonpath string, timeoutSecond int) (name string, err error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	args := []string{"get", resource, "-n", namespace, "-l", label, "-o", fmt.Sprintf(`jsonpath="%s"`, jsonpath)}
	return BaseCmd("", "", timeoutSecond, args...)
}
func GetYaml(resource, namespace, label string, timeoutSecond int) (out string, err error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	args := []string{"get", resource, "-n", namespace, "-l", label, "-o", "yaml"}
	return BaseCmd("", "", timeoutSecond, args...)
}

func DescribePod(deployName, namespace, label string, timeoutSecond int) string {
	return describe(deployName, namespace, label, "pod", timeoutSecond)
}

func ShowOBI(namespace, label string, timeoutSecond int) string {
	out, _ := GetYaml("obi", namespace, label, timeoutSecond)
	return out
}

func DescribeNode(label string, timeoutSecond int) string {
	return describe("", "", label, "node", timeoutSecond)
}

func describe(deployName, namespace, label, resource string, timeoutSecond int) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	if label == "" && resource == "pod" {
		label = "app=" + deployName
	}
	args := []string{"describe", resource, "-n", namespace, "-l", label}
	out, err := BaseCmd("", "", timeoutSecond, args...)
	if err != nil {
		return fmt.Sprintf("Describe %s[%s] error:[%s]", resource, strings.Join(args, " "), err)
	}
	return fmt.Sprintf("Describe %s of %s[%s] in ns:[%s]:\n\n %s", resource, deployName, label, namespace, out)
}

func TopNode(timeoutSecond int) string {
	return top("node", timeoutSecond)
}

func TopPod(timeoutSecond int) string {
	return top("pod", timeoutSecond)
}

func top(resource string, timeoutSecond int) string {
	args := []string{"top", resource}
	if resource == "pod" {
		args = append(args, "-A")
	}
	out, err := BaseCmd("", "", timeoutSecond, args...)
	if err != nil {
		return fmt.Sprintf("kubectl top %s error:[%s]", resource, err)
	}
	return fmt.Sprintf("kubectl top %s:\n\n %s", resource, out)
}
func BaseCmd(cmdName, stdIn string, timeoutSecond int, args ...string) (out string, err error) {
	if cmdName == "" {
		cmdName = Kubectl
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecond)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdin = strings.NewReader(stdIn)
	var o []byte
	o, err = cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("exec timeout[%ds]: %s %s", timeoutSecond, cmdName, strings.Join(args, " "))
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(o), fmt.Errorf("exec[%s %s] get err:[stderr:%s, %s]", cmdName, strings.Join(args, " "), string(exitErr.Stderr), exitErr.String())
		}
		return string(o), fmt.Errorf("exec[%s %s] get err:[%s]", cmdName, strings.Join(args, " "), err)
	}
	return strings.Trim(string(o), `"'`), nil
}

func DeleteDeploy(name string, namespace string, timeoutSecond int) error {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	args := []string{"delete", "deploy", "-n", namespace, name, "--wait=true"}
	_, err := BaseCmd("", "", timeoutSecond, args...)
	if err != nil {
		return fmt.Errorf("delete pod[%s] error:[%s]", strings.Join(args, " "), err)
	}
	return nil
}

func obiCountCommand() string {
	return `kubectl -narbiter-system get obi |grep -v 'NAME'|wc -l`
}

func obiDataCommand() []string {
	obiNames := []string{"metric-server-pod-cpu", "metric-server-pod-mem", "prometheus-pod-cpu", "prometheus-pod-mem",
		"metric-server-node-cpu", "metric-server-node-mem", "prometheus-node-cpu", "prometheus-node-mem", "prometheus-cluster-schedulable-cpu",
		"prometheus-max-available-cpu", "prometheus-rawdata-node-unschedule"}
	outputTemplate := `kubectl get obi %s -n %s -oyaml | grep 'timestamp' | wc -l`
	output := make([]string, len(obiNames))
	for idx, name := range obiNames {
		output[idx] = fmt.Sprintf(outputTemplate, name, "arbiter-system")
	}

	return output
}

func policyCountCommand() string {
	return `kubectl get ObservabilityActionPolicy -narbiter-system|grep -v 'NAME'|wc -l`
}

func podNodeLabelsCommand() []string {
	podLabels := []string{"metric-server-pod-cpu", "metric-server-pod-mem", "prometheus-pod-cpu", "prometheus-pod-mem"}
	nodeLabels := []string{"metric-server-node-cpu", "metric-server-node-mem", "prometheus-node-cpu", "prometheus-node-mem"}
	commands := make([]string, 0)
	for _, label := range podLabels {
		commands = append(commands, fmt.Sprintf(`kubectl get po -narbiter-system -l%s|grep -v 'NAME'|wc -l`, label))
	}

	for _, label := range nodeLabels {
		commands = append(commands, fmt.Sprintf(`kubectl get node -l%s|grep -v 'NAME'|wc -l`, label))
	}

	return commands
}

func countChecker(cmd string, output *string) error {
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		return err
	}
	*output = strings.TrimSpace(string(out))

	return nil
}
