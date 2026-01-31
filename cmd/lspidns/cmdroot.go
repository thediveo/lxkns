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

package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	"github.com/thediveo/go-asciitree/v2"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/cli/icon"
	"github.com/thediveo/lxkns/cmd/cli/reflabel"
	"github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/style"
	"github.com/thediveo/lxkns/cmd/cli/task"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/discover"

	_ "github.com/thediveo/clippy/debug"
)

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "lspidns",
		Short:   "lspidns shows the tree of PID namespaces",
		Version: lxkns.SemVersion,
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: lspidnscmd,
	}
	silent.PreferSilence(rootCmd)
	// Sets up the flags.
	rootCmd.PersistentFlags().BoolP(
		"user", "u", false,
		"shows owner user namespaces")
	clippy.AddFlags(rootCmd)
	return
}

func lspidnscmd(cmd *cobra.Command, _ []string) error {
	user, _ := cmd.PersistentFlags().GetBool("user")
	// Run a standard namespace discovery (comprehensive, but without
	// mount point discovery).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cizer := turtles.Containerizer(ctx, cmd)
	defer cizer.Close()
	allns := discover.Namespaces(
		discover.WithStandardDiscovery(),
		discover.WithContainerizer(cizer),
		discover.WithPIDMapper(), // recommended when using WithContainerizer.
		task.DiscoveryOption(cmd),
	)
	_, err := fmt.Fprint(cmd.OutOrStdout(),
		asciitree.Render(
			allns.PIDNSRoots,
			&PIDNSVisitor{
				ShowUserNamespaces:      user,
				NamespaceIcon:           icon.NamespaceIcon(cmd),
				NamespaceReferenceLabel: reflabel.NamespaceReferenceLabel(cmd),
			},
			style.NamespaceStyler))
	return err
}
