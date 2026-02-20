// Copyright 2026 Harald Albrecht.
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
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	"github.com/thediveo/go-asciitree/v2"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"

	_ "github.com/thediveo/clippy/debug"
)

var allowedProcesses = []string{
	"systemd",
	"systemd-journald",

	"containerd",
	"containerd-shim",
	"dockerd",
	"docker-init",

	"sleep",

	"kthreadd",
	"cpuhp/.*",
	"ksoftirqd/.*",
	"oom_reaper",
	"rcu_.*",
	"upowerd",
}

func pruneAndDump(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cizer := turtles.Containerizer(ctx, &cobra.Command{})
	defer cizer.Close()

	allns := discover.Namespaces(
		discover.WithFullDiscovery(),
		discover.WithContainerizer(cizer),
		discover.WithPIDMapper(), // recommended when using WithContainerizer.
		discover.WithAffinityAndScheduling(),
		discover.WithTaskAffinityAndScheduling(),
	)

	pruneProcesses(allns, matchAny(allowedProcesses))
	sanitizeProcesses(allns)
	for line := range strings.Lines(asciitree.Render(
		[]*model.Process{allns.Processes[1], allns.Processes[2]},
		&processVisitor{},
		asciitree.LineTreeStyler,
	)) {
		slog.Info(line) // yeah, yuk
	}

	pruneNamespaces(allns)
	for typIndex, namespaceMap := range allns.Namespaces {
		typname, _ := model.NamespaceTypeNameByIndex(model.NamespaceTypeIndex(typIndex))
		if typname == "pid" || typname == "user" {
			for line := range strings.Lines(asciitree.Render(
				[]model.Hierarchy{allns.Processes[1].Namespaces[typIndex].(model.Hierarchy)},
				&hierarchyVisitor{},
				asciitree.LineTreeStyler,
			)) {
				slog.Info(line)
			}
			continue
		}
		for _, namespace := range namespaceMap {
			slog.Info(namespace.String())
		}
	}

	purgeMounts(allns)
	for nsid := range allns.Mounts {
		slog.Info("mounts\n", slog.String("namespace", allns.Namespaces[model.MountNS][nsid].String()))
	}

	sanitizeContainers(allns)

	var jsondata []byte
	var err error
	if compact, _ := cmd.PersistentFlags().GetBool("compact"); compact {
		// Compact JSON output without spaces and newlines.
		jsondata, err = json.Marshal(types.NewDiscoveryResult(types.WithResult(allns)))
	} else {
		// Pretty-printed JSON output, with either tabs or spaces for
		// indentation.
		var indent string
		if tab, _ := cmd.PersistentFlags().GetBool("tab"); tab {
			indent = "\t"
		} else {
			spaces, _ := cmd.PersistentFlags().GetUint("indent")
			if spaces > 8 {
				spaces = 8
			}
			// ...still wondering why Repeat(" ", -2) should be even accepted at
			// compile time and using uint instead of int?? 😕
			indent = strings.Repeat(" ", int(spaces))
		}
		jsondata, err = json.MarshalIndent(
			types.NewDiscoveryResult(types.WithResult(allns)), "", indent)
	}
	if err != nil {
		return err
	}
	fmt.Println(string(jsondata))

	return nil
}

// newRootCmd creates the root command with usage and version information, as
// well as the available CLI flags (including descriptions).
func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "pruneddiscovery",
		Short:   "pruneddiscovery discovers, prunes, and outputs the final result JSON",
		Version: lxkns.SemVersion,
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: pruneAndDump,
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().BoolP(
		"compact", "c", false,
		"compact instead of pretty-printed output")
	rootCmd.PersistentFlags().BoolP(
		"tab", "t", false,
		"use tabs for indentation instead of spaces")
	rootCmd.PersistentFlags().UintP(
		"indent", "i", 2,
		"use the given number of spaces (no more than 8) for indentation")

	clippy.AddFlags(rootCmd)
	return
}

func main() {
	// This is cobra boilerplate documentation, except for the missing call to
	// fmt.Println(err) which in the original boilerplate is just plain wrong:
	// it renders the error message twice, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
