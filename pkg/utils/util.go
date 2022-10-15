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

const AddFinalizerErrorMsg = "update finalizer"

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
