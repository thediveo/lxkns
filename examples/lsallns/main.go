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
	"os"

	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/klo"
	"github.com/thediveo/lxkns"
)

// NamespaceRow stores information about a single namespace, to be printed
// as a single row.
type NamespaceRow struct {
	ID       uint64
	Type     string
	PID      int
	ProcName string
}

func main() {
	// For some discovery methods this app must be forked and re-executed; the
	// call to reexec.CheckAction() will automatically handle this situation
	// and then never return when in re-execution.
	reexec.CheckAction()
	// Run a full namespace discovery.
	result := lxkns.Discover(lxkns.FullDiscovery)
	// Prepare output list from the discovery results. For this, we iterate
	// over all types of namespaces, because the discovery results contain the
	// namespaces organized by type of namespace.
	list := []NamespaceRow{}
	for nsidx := range result.Namespaces {
		for _, ns := range result.SortedNamespaces(lxkns.NamespaceTypeIndex(nsidx)) {
			item := NamespaceRow{
				ID:   uint64(ns.ID()),
				Type: ns.Type().Name(),
			}
			if procs := ns.Leaders(); len(procs) > 0 {
				item.PID = int(procs[0].PID)
				item.ProcName = procs[0].Name
			}
			list = append(list, item)
		}
	}
	// For outputting a neat table, which is even sorted, we rely on "klo",
	// the "kubectl-like outputter". The DefaultColumnSpec specifies the table
	// headers in the form of "<Headertext>:{<JSON-Path-Expression>}".
	prn, err := klo.PrinterFromFlag("",
		&klo.Specs{DefaultColumnSpec: "NAMESPACE:{.ID},TYPE:{.Type},PID:{.PID},PROCESS:{.ProcName}"})
	if err != nil {
		panic(err)
	}
	table, err := klo.NewSortingPrinter("{.ID}", prn)
	if err != nil {
		panic(err)
	}
	table.Fprint(os.Stdout, list)
}
