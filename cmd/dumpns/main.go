// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns"
	apitypes "github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
	"github.com/thediveo/lxkns/cmd/internal/pkg/task"
	"github.com/thediveo/lxkns/cmd/internal/pkg/turtles"
	"github.com/thediveo/lxkns/discover"
)

// dumpns emits the namespace and process discovery results as JSON. It takes
// JSON formatting options into account, such as not indenting output, or using
// tabs or a specific number of spaces for indentation.
func dumpns(cmd *cobra.Command, _ []string) error {
	containerizer := turtles.Containerizer(context.Background(), cmd)
	allns := discover.Namespaces(
		discover.WithStandardDiscovery(),
		discover.WithContainerizer(containerizer),
		discover.WithPIDMapper(), // recommended when using WithContainerizer.
		task.FromTasks(cmd),
	)
	var jsondata []byte
	var err error
	if compact, _ := cmd.PersistentFlags().GetBool("compact"); compact {
		// Compact JSON output without spaces and newlines.
		jsondata, err = json.Marshal(apitypes.NewDiscoveryResult(apitypes.WithResult(allns)))
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
			apitypes.NewDiscoveryResult(apitypes.WithResult(allns)), "", indent)
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
		Use:     "dumpns",
		Short:   "dumpns outputs discovered namespaces and processes as JSON",
		Version: lxkns.SemVersion,
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return cli.BeforeCommand(cmd)
		},
		RunE: dumpns,
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().BoolP(
		"compact", "c", false,
		"compact instead of pretty-printed output")
	rootCmd.PersistentFlags().BoolP(
		"tab", "t", false,
		"use tabs for indentation instead of two spaces")
	rootCmd.PersistentFlags().UintP(
		"indent", "i", 2,
		"use the given number of spaces (no more than 8) for indentation")

	cli.AddFlags(rootCmd)
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
