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

package engines

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	_ "github.com/thediveo/lxkns/cmd/internal/pkg/engines/test/broken"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("logging", Ordered, func() {

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

	It("ignores all engines", func(ctx context.Context) {
		rootCmd.SetArgs([]string{"foo", "--" + NoEnginesFlagName})
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(Containerizer(ctx, rootCmd, false)).To(BeNil())
	})

	It("informs the user in case of engine issues", func(ctx context.Context) {
		rootCmd.SetArgs([]string{"foo", "--" + KeepGoingFlagName + "=false"})
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(Containerizer(ctx, rootCmd, false)).Error().
			To(MatchError(ContainSubstring("broken engine plugin")))
	})

})
