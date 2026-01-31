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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/spf13/cobra"
	"github.com/thediveo/klo"
	"github.com/thediveo/lxkns"
	apitypes "github.com/thediveo/lxkns/api/types"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

// NamespaceRow stores information about a single namespace, to be printed
// as a single row.
type NamespaceRow struct {
	ID            uint64
	Type          string
	PID           int
	ContainerName string
	ProcName      string
	Comment       string
}

// rowItem returns the row information about the specified namespace.
func rowItem(ns model.Namespace) NamespaceRow {
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
		if proc.Container != nil {
			item.ContainerName = proc.Container.Name
		}
		item.Comment = "cgroup:" + proc.CpuCgroup
		return item
	}
	if looseThreads := ns.LooseThreads(); len(looseThreads) > 0 {
		looseThreads = slices.Clone(looseThreads)
		slices.SortFunc(looseThreads,
			func(task1, task2 *model.Task) int {
				switch v := task1.Starttime - task2.Starttime; {
				case v > ^uint64(0)>>1:
					return -1
				case v == 0:
					return 0
				default:
					return 1
				}
			})
		task := looseThreads[0]
		item.PID = int(task.TID)
		item.ProcName = "[" + task.Name + "]"
		if task.Process.Container != nil {
			item.ContainerName = task.Process.Container.Name
		}
		item.Comment = "cgroup:" + task.CpuCgroup
		return item
	}
	if ref := ns.Ref(); len(ref) != 0 {
		item.Comment = "bound:" + ns.Ref().String()
		return item
	}
	item.Comment = "???"
	return item
}

// dumpresult takes discovery results, extracts the required fields, and then
// dumps the extracted data to stdout in a neat ASCII table.
func dumpresult(result *discover.Result) error {
	// Prepare output list from the discovery results. For this, we iterate
	// over all types of namespaces, because the discovery results contain the
	// namespaces organized by type of namespace.
	list := []NamespaceRow{}
	for nsidx := range result.Namespaces {
		for _, namespace := range result.SortedNamespaces(model.NamespaceTypeIndex(nsidx)) {
			list = append(list, rowItem(namespace))
		}
	}
	// For outputting a neat table, which is even sorted, we rely on "klo",
	// the "kubectl-like outputter". The DefaultColumnSpec specifies the table
	// headers in the form of "<Headertext>:{<JSON-Path-Expression>}".
	prn, err := klo.PrinterFromFlag("",
		&klo.Specs{DefaultColumnSpec: "NAMESPACE:{.ID},TYPE:{.Type},CONTAINER:{.ContainerName},PID:{.PID},PROCESS/[TASK]:{.ProcName},COMMENT:{.Comment}"})
	if err != nil {
		return err
	}
	table, err := klo.NewSortingPrinter("{.ID}", prn)
	if err != nil {
		return err
	}
	return table.Fprint(os.Stdout, list)
}

// lsallns works on the given CLI flags to decide whether to run its own
// Linux-kernel namespaces discovery or to load existing results in JSON format
// from a file (or stdin). It then dumps the discovery results in a neat ASCII
// table to stdout.
func lsallns(cmd *cobra.Command, _ []string) error {
	var result *discover.Result
	if input, _ := cmd.PersistentFlags().GetString("input"); input != "" {
		var r io.Reader
		if input == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(input) // #nosec G304
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			r = f
		}
		dr := apitypes.NewDiscoveryResult()
		if err := json.NewDecoder(r).Decode(dr); err != nil {
			return fmt.Errorf("cannot decode discovery results, %w", err)
		}
		result = dr.Result()
	} else {
		// Set up a Docker engine-connected containerizer
		moby, err := moby.New("", nil)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cizer := whalefriend.New(ctx, []watcher.Watcher{moby})
		// Run a full namespace discovery without mount point discovery, but
		// with containers.
		result = discover.Namespaces(
			discover.WithStandardDiscovery(),
			discover.WithContainerizer(cizer),
			discover.WithPIDMapper(), // recommended when using WithContainerizer.
		)
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
	// This is cobra boilerplate documentation, except for the missing call to
	// fmt.Println(err) which in the original boilerplate is just plain wrong:
	// it renders the error message twice, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
