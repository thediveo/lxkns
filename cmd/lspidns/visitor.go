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
type PIDNSVisitor struct {
	ShowUserNS bool
}

// Roots returns the given topmost hierarchical PID namespaces sorted. Well, to
// be honest, it returns the owning user namespaces instead in case user
// namespaces should be shown too.
func (v *PIDNSVisitor) Roots(roots reflect.Value) (children []reflect.Value) {
	pidroots := lxkns.SortNamespaces(roots.Interface().([]lxkns.Namespace))
	if !v.ShowUserNS {
		// When only showing PID namespaces, then sort all PID "root" namespaces
		// numerically and then visit them, descending down the hierarchy.
		count := len(pidroots)
		children = make([]reflect.Value, count)
		for idx := 0; idx < count; idx++ {
			children[idx] = reflect.ValueOf(pidroots[idx])
		}
		return
	}
	// When showing the owning user namespaces in the tree, find the (unique)
	// user namespaces for all PID "root" namespaces, so we start with the user
	// namespaces.
	userns := map[lxkns.Namespace]bool{}
	for _, pidns := range pidroots {
		userns[pidns.Owner().(lxkns.Namespace)] = true
	}
	userroots := []lxkns.Namespace{}
	for uns := range userns {
		userroots = append(userroots, uns.(lxkns.Namespace))
	}
	userroots = lxkns.SortNamespaces(userroots)
	count := len(userroots)
	children = make([]reflect.Value, count)
	for idx := 0; idx < count; idx++ {
		children[idx] = reflect.ValueOf(userroots[idx])
	}
	return
}

// Label returns the text label for a namespace node. Everything else will have
// no label.
func (v *PIDNSVisitor) Label(node reflect.Value) (label string) {
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
func (v *PIDNSVisitor) Get(node reflect.Value) (
	label string, properties []string, children reflect.Value) {
	// Label for this pid (or user) namespace...
	label = v.Label(node)
	// For a user namespace, the children are the owned PID namespaces.
	if uns, ok := node.Interface().(lxkns.Ownership); ok {
		pidns := lxkns.SortedNamespaces(uns.Ownings()[lxkns.PIDNS])
		children = reflect.ValueOf(pidns)
		return
	}
	// For a PID namespace, the children are either PID namespaces, or user
	// namespaces in case a child PID namespace lives in a different user
	// namespace.
	clist := []interface{}{}
	if hns, ok := node.Interface().(lxkns.Hierarchy); ok {
		if !v.ShowUserNS {
			// Show only the PID namespace hierarchy.
			for _, cpidns := range lxkns.SortChildNamespaces(hns.Children()) {
				clist = append(clist, cpidns)
			}
		} else {
			// Insert user namespaces into the PID namespace hierarchy, when
			// there is a change of user namespaces in the PID hierarchy.
			userns := node.Interface().(lxkns.Namespace).Owner()
			for _, cpidns := range lxkns.SortChildNamespaces(hns.Children()) {
				if ownerns := cpidns.(lxkns.Namespace).Owner(); ownerns == userns {
					// The child PID namespace is still in the same user namespace,
					// so we take it as a direct child.
					clist = append(clist, cpidns)
				} else {
					// The child PID namespace is in a different user namespace, so
					// we take the child's user namespace instead and will visit the
					// child PID namespace only later via the user namespace.
					clist = append(clist, ownerns)
				}
			}
		}
	}
	children = reflect.ValueOf(clist)
	return
}

// leadersString lists the (leader) processes joined to a namespace in text
// form.
func leadersString(ns lxkns.Namespace) string {
	procs := "process (none)"
	if ancient := ns.Ealdorman(); ancient != nil {
		procs = "process " +
			fmt.Sprintf("%q (%d)",
				style.ProcessStyle.V(style.ProcessName(ancient)),
				ancient.PID)
	}
	return procs
}
