// lsallns -- lists *all* namespaces.

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
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/klo"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/model"
)

// NamespaceRow stores information about a single namespace, to be printed
// as a single row.
type NamespaceRow struct {
	ID       uint64
	Type     string
	PID      int
	ProcName string
	Comment  string
}

func dumpresult(result *lxkns.DiscoveryResult) error {
	// Prepare output list from the discovery results. For this, we iterate
	// over all types of namespaces, because the discovery results contain the
	// namespaces organized by type of namespace.
	list := []NamespaceRow{}
	for nsidx := range result.Namespaces {
		for _, ns := range result.SortedNamespaces(model.NamespaceTypeIndex(nsidx)) {
			item := NamespaceRow{
				ID:   uint64(ns.ID().Ino),
				Type: ns.Type().Name(),
			}
			// Try to be consistent by always showing the "most senior"
			// process joined to a particular namespace. And yes, namespaces
			// might be kind of a Last Kingdom ;)
			if proc := ns.Ealdorman(); proc != nil {
				item.PID = int(proc.PID)
				item.ProcName = proc.Name
				item.Comment = proc.Controlgroup
			} else if ns.Ref() != "" {
				item.Comment = "bound:" + ns.Ref()
			}
			list = append(list, item)
		}
	}
	// For outputting a neat table, which is even sorted, we rely on "klo",
	// the "kubectl-like outputter". The DefaultColumnSpec specifies the table
	// headers in the form of "<Headertext>:{<JSON-Path-Expression>}".
	prn, err := klo.PrinterFromFlag("",
		&klo.Specs{DefaultColumnSpec: "NAMESPACE:{.ID},TYPE:{.Type},PID:{.PID},PROCESS:{.ProcName},COMMENT:{.Comment}"})
	if err != nil {
		return err
	}
	table, err := klo.NewSortingPrinter("{.ID}", prn)
	if err != nil {
		return err
	}
	table.Fprint(os.Stdout, list)
	return nil
}

func lsallns(cmd *cobra.Command, _ []string) error {
	var result *lxkns.DiscoveryResult
	if input, _ := cmd.PersistentFlags().GetString("input"); input != "" {
		var r io.Reader
		if input == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(input)
			if err != nil {
				return err
			}
			defer f.Close()
			r = f
		}
		// TODO: use correct discovery result method.
		dr := (*types.DiscoveryResult)(&lxkns.DiscoveryResult{
			Namespaces: *model.NewAllNamespaces(),
			Processes:  model.ProcessTable{},
		})
		if err := json.NewDecoder(r).Decode(dr); err != nil {
			return fmt.Errorf("cannot decode discovery results, %w", err)
		}
		result = (*lxkns.DiscoveryResult)(dr)
	} else {
		// Run a full namespace discovery.
		result = lxkns.Discover(lxkns.FullDiscovery)
	}
	return dumpresult(result)
}

// newRootCmd creates the root command with usage and version information, as
// well as the available CLI flags (including descriptions).
func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "lsallns",
		Short:   "lsallns lists *all* namespaces ;)",
		Version: lxkns.SemVersion,
		Args:    cobra.NoArgs,
		RunE:    lsallns,
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().StringP(
		"input", "i", "",
		"reads discovery information from JSON file or '-' stdin")
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
