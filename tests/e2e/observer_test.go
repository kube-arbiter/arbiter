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

var _ = Describe("Observer", Label("observer"), func() {

	Describe("Check the number and data of obi", Label("base", "quick"), func() {
		It("start to check the number of obi", func() {
			var output string
			Expect(countChecker(obiCountCommand(), &output)).Error().Should(Succeed(), func() {
				fmt.Println("check the number of obi successfully")
			})

			Expect(output == "11").Should(BeTrue())
		})

		It("start to check the data of obi", func() {
			var output string
			Expect(countChecker(obiDataCommand(), &output)).Error().Should(Succeed())

			Expect(output == "11").Should(BeTrue())
		})

	})
})

var _ = Describe("Executor", Label("executor"), func() {
	Describe("Check the number and data of ObservabilityActionPolicy", Label("base", "quick"), func() {
		It("start to check the number of ObservabilityActionPolicy", func() {
			var output string
			Expect(countChecker(policyCountCommand(), &output)).Error().Should(Succeed(), func() {
				fmt.Println("check the number of ObservabilityActionPolicy successfully")
			})

			Expect(output == "8").Should(BeTrue())
		})

		It("start to check data of ObservabilityActionPolicy", func() {
			commands := podNodeLabelsCommand()
			var output string
			for _, cmd := range commands {
				Expect(countChecker(cmd, &output)).Error().Should(Succeed())
				Expect(output == "1").Should(BeTrue())

			}
		})
	})
})
