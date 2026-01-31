// A visitor implementing the view on the process tree and PID namespaces.

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
	"slices"

	"github.com/thediveo/go-asciitree/v2"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/internal/xslices"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// TreeVisitor is an asciitree.Visitor which works on discovery results and
// visits them in order to produce a process tree. Differing from `ps fax`, we
// also show the PID namespaces in between the process hierarchy where the PID
// namespace changes from one to another.
type TreeVisitor struct {
	Details   bool
	PIDMap    model.PIDMapper
	RootPIDNS model.Namespace
	// render function for namespace icons, where its exact behavior depends on
	// CLI flags.
	NamespaceIcon func(model.Namespace) string
	// render function for cgroup (path) names
	CgroupDisplayName func(string) string
}

var _ asciitree.Visitor = (*TreeVisitor)(nil)

// Roots returns the given "topmost" hierarchical process namespaces sorted;
// it will be called on the list of "topmost" PID namespace(s). It won't be
// called ever afterwards. In our case, we'll only ever pass in a slice of
// exactly one PID namespace, the "root" PID namespace. However, we leave this
// code in for instructional purposes.
func (v *TreeVisitor) Roots(roots any) (children []any) {
	rootNamespaces, _ := roots.([]model.Namespace)
	discover.SortNamespaces(rootNamespaces)
	return xslices.Any(rootNamespaces)
}

// Label returns a node label text, which varies depending on whether the node
// is a Process or a (PID) Namespace. In case of a PID Namespace, the label
// will show the namespace type and its ID, as well as the owner name and UID
// (via the owning user Namespace). If it's a Process instead, then the text
// label contains the name and "global" PID, but also the translated "local"
// PID (which is the PID as seen from inside the PID namespace of the
// Process).
func (v *TreeVisitor) Label(node any) (label string) {
	if proc, ok := node.(*model.Process); ok {
		return ProcessLabel(proc, v.PIDMap, v.RootPIDNS, v.CgroupDisplayName)
	}
	return PIDNamespaceLabel(node.(model.Namespace), v.NamespaceIcon)
}

// Get is called on nodes which can be either (1) PID namespaces or (2)
// processes. For (1), the visitor returns information about the PID
// namespace, but then specifies processes as children. For (2), the visitor
// returns process children, unless these children are in a different PID
// namespace: then, their PID namespaces are returned instead. Polymorphism
// galore!
func (v *TreeVisitor) Get(node any) (label string, properties []string, children []any) {
	// Label for this (1) PID namespace or (2) process.
	label = v.Label(node)
	// Children of this (1) PID namespace are always processes, but for (2)
	// processes the children can be any combination of (a) child processes
	// still in the same namespace and (b) child PID namespaces.
	if proc, ok := node.(*model.Process); ok {
		nodePIDNs := proc.Namespaces[model.PIDNS]
		childProcs := slices.Clone(proc.Children)
		slices.SortFunc(childProcs, model.SortProcessByPID)

		childPIDNsSet := map[species.NamespaceID]struct{}{}
		for _, childProc := range childProcs {
			if childProc.Namespaces[model.PIDNS] == nodePIDNs {
				children = append(children, childProc)
				continue
			}
			// We might also end up here in case we have insufficient
			// privileges (capabilities) to discover the PID namespace of
			// a process. In this case, we only can dump the processes,
			// but with a signature indicating that we don't known about
			// their namespaces. Otherwise, we insert a PID namespace
			// node, from which the tree will branch into that PID
			// namespace's leader processes.
			childPIDNs := childProc.Namespaces[model.PIDNS]
			if childPIDNs == nil {
				// PID namespace is not known.
				children = append(children, childProc)
				continue
			}
			// Ensure to add the same PID namespace only once.
			if _, ok := childPIDNsSet[childPIDNs.ID()]; !ok {
				// Avoid ending up in a circular loop if we are dealing
				// with an intermediate process node where we could not
				// determine its PID namespace and now our child yet has
				// again a properly discovered PID namespace. Oh, the
				// fallacies of non-root...
				if nodePIDNs == nil {
					proc /*sic!*/ := proc.Parent
					hasPIDNsAbove := false
					for proc != nil {
						if proc.Namespaces[model.PIDNS] == childPIDNs {
							hasPIDNsAbove = true
							break
						}
						proc = proc.Parent
					}
					if hasPIDNsAbove {
						children = append(children, childProc)
						continue
					}
				}

				children = append(children, childPIDNs)
				childPIDNsSet[childPIDNs.ID()] = struct{}{}
			}
		}
	} else {
		// The child nodes of a PID namespace tree node will be the "leader"
		// (or "topmost") processes inside the PID namespace.
		leaders := slices.Clone(node.(model.Namespace).Leaders())
		slices.SortFunc(leaders, model.SortProcessByPID)
		for _, proc := range leaders {
			children = append(children, proc)
		}
	}
	return label, nil, children
}
