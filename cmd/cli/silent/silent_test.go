// Copyright 2020 Harald Albrecht.
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

package silent

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/safe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("keeping shtumm", func() {

	BeforeEach(func() {
		oldDefault := slog.Default()
		DeferCleanup(func() { slog.SetDefault(oldDefault) })
	})

	It("defaults to informing", func() {
		cmd := cobra.Command{
			RunE: func(*cobra.Command, []string) error {
				slog.Debug("**DEBUG**")
				slog.Info("**INFO**")
				slog.Error("**ERROR**")
				return nil
			},
			PreRunE: func(cmd *cobra.Command, _ []string) error {
				return clippy.BeforeCommand(cmd)
			},
		}
		cmd.SetArgs([]string{})
		clippy.AddFlags(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(&cmd, &out)
		Expect(cmd.Execute()).To(Succeed())
		output := out.String()
		Expect(output).To(ContainSubstring("**ERROR**"))
		Expect(output).To(ContainSubstring("**INFO**"))
		Expect(output).NotTo(ContainSubstring("**DEBUG**"))
	})

	It("keeps silent", func() {
		cmd := cobra.Command{
			RunE: func(*cobra.Command, []string) error {
				slog.Debug("**DEBUG**")
				slog.Info("**INFO**")
				slog.Error("**ERROR**")
				return nil
			},
			PreRunE: func(cmd *cobra.Command, _ []string) error {
				return clippy.BeforeCommand(cmd)
			},
		}
		cmd.SetArgs([]string{"--" + SilentFlagName})
		clippy.AddFlags(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(&cmd, &out)
		Expect(cmd.Execute()).To(Succeed())
		output := out.String()
		Expect(output).To(ContainSubstring("**ERROR**"))
		Expect(output).NotTo(ContainSubstring("**INFO**"))
		Expect(output).NotTo(ContainSubstring("**DEBUG**"))
	})

	It("informs when told to not be silent", func() {
		cmd := cobra.Command{
			RunE: func(*cobra.Command, []string) error {
				slog.Debug("**DEBUG**")
				slog.Info("**INFO**")
				slog.Error("**ERROR**")
				return nil
			},
			PreRunE: func(cmd *cobra.Command, _ []string) error {
				return clippy.BeforeCommand(cmd)
			},
		}
		cmd.SetArgs([]string{"--" + SilentFlagName + "=false"})
		clippy.AddFlags(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(&cmd, &out)
		Expect(cmd.Execute()).To(Succeed())
		output := out.String()
		Expect(output).To(ContainSubstring("**ERROR**"))
		Expect(output).To(ContainSubstring("**INFO**"))
		Expect(output).NotTo(ContainSubstring("**DEBUG**"))
	})

	It("logs details", func() {
		cmd := cobra.Command{
			RunE: func(*cobra.Command, []string) error {
				slog.Debug("**DEBUG**")
				slog.Info("**INFO**")
				slog.Error("**ERROR**")
				return nil
			},
			PreRunE: func(cmd *cobra.Command, _ []string) error {
				return clippy.BeforeCommand(cmd)
			},
		}
		cmd.SetArgs([]string{"--debug"})
		clippy.AddFlags(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(&cmd, &out)
		Expect(cmd.Execute()).To(Succeed())
		output := out.String()
		Expect(output).To(ContainSubstring("**ERROR**"))
		Expect(output).To(ContainSubstring("**INFO**"))
		Expect(output).To(ContainSubstring("**DEBUG**"))
	})

	It("prefers silence", func() {
		cmd := cobra.Command{
			RunE: func(*cobra.Command, []string) error {
				slog.Debug("**DEBUG**")
				slog.Info("**INFO**")
				slog.Error("**ERROR**")
				return nil
			},
			PreRunE: func(cmd *cobra.Command, _ []string) error {
				return clippy.BeforeCommand(cmd)
			},
		}
		cmd.SetArgs([]string{})
		clippy.AddFlags(&cmd)
		PreferSilence(&cmd)

		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(&cmd, &out)
		Expect(cmd.Execute()).To(Succeed())
		output := out.String()
		Expect(output).To(ContainSubstring("**ERROR**"))
		Expect(output).NotTo(ContainSubstring("**INFO**"))
		Expect(output).NotTo(ContainSubstring("**DEBUG**"))
	})

})
