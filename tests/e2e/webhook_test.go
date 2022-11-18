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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("webhook", Label("webhook"), func() {

	Describe("Check the resource of pod", Label("base", "quick"), func() {
		It("start to check the resource of pod", func() {
			var output string
			Expect(countChecker(testpodResourceCommand(), &output)).Error().Should(Succeed(), func() {
				fmt.Println("check the resource of pod successfully")
			})

			Expect(output).To(Equal(`{"cpu":"50m","memory":"629145Ki"}`), func() string {
				return output
			})
		})
	})
})
