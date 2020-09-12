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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/cmd/internal/pkg/cli"
)

// dump emits the namespace and process discovery results as JSON. It takes
// formatting options into account, such as not indenting output, or using
// tabs or a specific number of spaces for indentation.
func dump(cmd *cobra.Command, _ []string) error {
	allns := lxkns.Discover(lxkns.FullDiscovery)
	var j []byte
	var err error
	if compact, _ := cmd.PersistentFlags().GetBool("compact"); compact {
		// Compact JSON output without spaces and newlines.
		j, err = json.Marshal((*types.DiscoveryResult)(allns))
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
			indent = strings.Repeat(" ", int(spaces))
		}
		j, err = json.MarshalIndent((*types.DiscoveryResult)(allns), "", indent)
	}
	if err != nil {
		return err
	}
	fmt.Println(string(j))
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
		RunE:    dump,
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
	// For some discovery methods this app must be forked and re-executed; the
	// call to reexec.CheckAction() will automatically handle this situation
	// and then never return when in re-execution.
	reexec.CheckAction()
	// Otherwise, this is cobra boilerplate documentation, except for the
	// missing call to fmt.Println(err) which in the original boilerplate is
	// just plain wrong: it renders the error message twice, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
