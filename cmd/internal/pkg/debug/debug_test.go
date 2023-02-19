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

package debug

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/cmd/internal/test/getstdout"
	"github.com/thediveo/lxkns/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func slurp(f func()) func() {
	return func() {
		defer func(w io.Writer) { logrus.StandardLogger().Out = w }(logrus.StandardLogger().Out) // extremely readable :)
		logrus.StandardLogger().Out = os.Stderr
		f()
	}
}

// This unit test was writting during a unit test frenzy, triggered by zero
// trust. ha ha. HAHA. HAHAHAHAHA. ARGH!
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

	It("defaults to logging only fatalities", func() {
		rootCmd.SetArgs([]string{"foo"})
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(getstdout.Stdouterr(slurp(func() { log.Debugf("debug") }))).To(BeEmpty())
		Expect(getstdout.Stdouterr(slurp(func() { log.Infof("info") }))).To(BeEmpty())
	})

	It("logs information", func() {
		rootCmd.SetArgs([]string{"foo", "--" + LogFlagName})
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(getstdout.Stdouterr(slurp(func() { log.Debugf("debug") }))).To(BeEmpty())
		Expect(getstdout.Stdouterr(slurp(func() { log.Infof("info") }))).To(MatchRegexp(`INFO.*info`))
	})

	It("shows debug information", func() {
		rootCmd.SetArgs([]string{"foo", "--" + DebugFlagName})
		var err error
		out := getstdout.Stdouterr(slurp(func() {
			err = rootCmd.Execute()
		}))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(MatchRegexp(`DEBU.*debug logging enabled`))
		Expect(getstdout.Stdouterr(slurp(func() { log.Debugf("debug") }))).To(MatchRegexp(`DEBU.*debug`))
		Expect(getstdout.Stdouterr(slurp(func() { log.Infof("info") }))).To(MatchRegexp(`INFO.*info`))
	})

})
