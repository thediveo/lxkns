// Copyright 2023 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package task

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This unit test was writting during a unit test frenzy, triggered by zero
// trust. ha ha. HAHA. HAHAHAHAHA. ARGH!
var _ = Describe("task CLI flag", Ordered, func() {

	var rootCmd *cobra.Command

	BeforeEach(func() {
		rootCmd = &cobra.Command{
			PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
				return cli.BeforeCommand(cmd)
			},
			RunE: func(*cobra.Command, []string) error { return nil },
		}
		cli.AddFlags(rootCmd)
	})

	It("disables task discovery", func() {
		rootCmd.SetArgs([]string{"foo", "--" + TaskFlagName + "=false"})
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(Enabled(rootCmd)).To(BeFalse())
		Expect(FromTasks(rootCmd)).To(BeNil())
	})

	It("defaults to task discovery", func() {
		rootCmd.SetArgs([]string{"foo"})
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(Enabled(rootCmd)).To(BeTrue())
		Expect(FromTasks(rootCmd)).NotTo(BeNil())
	})

})
