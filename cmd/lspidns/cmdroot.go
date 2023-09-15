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
	asciitree "github.com/thediveo/go-asciitree"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/cmd/internal/pkg/task"
	"github.com/thediveo/lxkns/cmd/internal/pkg/turtles"
	"github.com/thediveo/lxkns/discover"

	_ "github.com/thediveo/lxkns/cmd/internal/pkg/debug"
)

func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "lspidns",
		Short:   "lspidns shows the tree of PID namespaces",
		Version: lxkns.SemVersion,
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return cli.BeforeCommand(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			user, _ := cmd.PersistentFlags().GetBool("user")
			// Run a standard namespace discovery (comprehensive, but without
			// mount point discovery).
			cizer := turtles.Containerizer(context.Background(), cmd)
			allns := discover.Namespaces(
				discover.WithStandardDiscovery(),
				discover.WithContainerizer(cizer),
				discover.WithPIDMapper(), // recommended when using WithContainerizer.
				task.FromTasks(cmd),
			)
			fmt.Print(
				asciitree.Render(
					allns.PIDNSRoots,
					&PIDNSVisitor{
						ShowUserNS: user,
					},
					style.NamespaceStyler))
			return nil
		},
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().BoolP(
		"user", "u", false,
		"shows owner user namespaces")
	cli.AddFlags(rootCmd)
	return
}
