// A visitor implementing the view on user namespaces and their owned
// namespaces, using the lxkns information model.

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
	"fmt"
	"os/user"
	"reflect"

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
)

// UserNSVisitor is an asciitree.Visitor which starts from a list (slice) of
// root user namespaces and then recursively dives into the user namespace
// hierarchy.
type UserNSVisitor struct {
	Details bool
}

// Roots returns the given topmost hierarchical user namespaces sorted.
func (v *UserNSVisitor) Roots(roots reflect.Value) (children []reflect.Value) {
	userroots := lxkns.SortNamespaces(roots.Interface().([]lxkns.Namespace))
	count := len(userroots)
	children = make([]reflect.Value, count)
	for idx := 0; idx < count; idx++ {
		children[idx] = reflect.ValueOf(userroots[idx])
	}
	return
}

// Label returns the text label for a namespace node. Everything else will have
// no label.
func (v *UserNSVisitor) Label(node reflect.Value) (label string) {
	if ns, ok := node.Interface().(lxkns.Namespace); ok {
		style := style.Styles[ns.Type().Name()]
		label = fmt.Sprintf("%s %s",
			style.V(ns.(lxkns.NamespaceStringer).TypeIDString()),
			leadersString(ns))
	}
	if uns, ok := node.Interface().(lxkns.Ownership); ok {
		username := ""
		if user, err := user.LookupId(fmt.Sprintf("%d", uns.UID())); err == nil {
			username = fmt.Sprintf(" (%q)", style.OwnerStyle.V(user.Username))
		}
		label += fmt.Sprintf(" created by UID %d%s",
			style.OwnerStyle.V(uns.UID()),
			username)
	}
	return
}

// Get returns the user namespace text label for the current node (which is
// always an user namespace), as well as the list of properties (owned
// non-user namespaces) and the list of child user namespace nodes.
func (v *UserNSVisitor) Get(node reflect.Value) (
	label string, properties []string, children reflect.Value) {
	// Label for this user namespace...
	label = v.Label(node)
	// Children of this user namespace...
	if hns, ok := node.Interface().(lxkns.Hierarchy); ok {
		children = reflect.ValueOf(lxkns.SortChildNamespaces(hns.Children()))
	}
	// Owned (non-user) namespaces...
	if v.Details {
		if uns, ok := node.Interface().(lxkns.Ownership); ok {
			ownedns := uns.Ownings()
			for _, nstype := range lxkns.TypeIndexLexicalOrder {
				if nstype == lxkns.UserNS {
					// Just in case: skip owned user namespaces; note: they
					// shouldn't be present in the model anyway.
					continue
				}
				nslist := lxkns.SortedNamespaces(ownedns[nstype])
				for _, ns := range nslist {
					style := style.Styles[ns.Type().Name()]
					s := fmt.Sprintf("%s %s",
						style.V(ns.(lxkns.NamespaceStringer).TypeIDString()),
						leadersString(ns))
					properties = append(properties, s)
				}
			}
		}
	}
	return
}

// leadersString lists the (leader) processes joined to a namespace in text
// form.
//
// TODO: implement control over simplified/full representation forms
func leadersString(ns lxkns.Namespace) string {
	procs := "process (none)"
	if ancient := ns.Ealdorman(); ancient != nil {
		procs = "process " +
			fmt.Sprintf("%q (%d)",
				style.ProcessStyle.V(style.ProcessName(ancient)),
				ancient.PID)
	}
	/*
		if leaders := ns.Leaders(); len(leaders) > 0 {
				sorted := make([]*lxkns.Process, len(leaders))
				copy(sorted, leaders)
				sort.Slice(sorted, func(i, j int) bool {
					return sorted[i].PID < sorted[j].PID
				})
				s := []string{}
				for _, leader := range sorted {
					s = append(s, fmt.Sprintf("%q (%d)", leader.Name, leader.PID))
				}
				procs = strings.Join(s, ", ")
				if len(leaders) > 1 {
					procs = "processes " + procs
				} else {
					procs = "process " + procs
				}
		}
	*/
	return procs
}
