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

	"github.com/thediveo/klo"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/nstypes"
)

// NamespaceRow stores information about a single namespace, to be printed
// as a single row.
type NamespaceRow struct {
	ID       nstypes.NamespaceID
	Type     string
	PID      int
	ProcName string
}

func main() {
	// Run a full namespace discovery.
	result := lxkns.Discover(lxkns.FullDiscovery)
	// Prepare output list from the discovery results. For this, we iterate
	// over all types of namespaces, because the discovery results contain the
	// namespaces organized by type of namespace.
	list := []NamespaceRow{}
	for nsidx := lxkns.MountNS; nsidx < lxkns.NamespaceTypesCount; nsidx++ {
		for _, ns := range result.SortedNamespaces(nsidx) {
			item := NamespaceRow{
				ID:   ns.ID(),
				Type: nstypes.TypeName(ns.Type()),
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
		&klo.Specs{DefaultColumnSpec: "NamespaceRow:{.ID},TYPE:{.Type},PID:{.PID},PROCESS:{.ProcName}"})
	if err != nil {
		panic(err)
	}
	table, err := klo.NewSortingPrinter("{.ID}", prn)
	if err != nil {
		panic(err)
	}
	table.Fprint(os.Stdout, list)
}
