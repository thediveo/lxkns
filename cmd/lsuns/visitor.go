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
	"github.com/thediveo/lxkns/cmd/internal/pkg/filter"
	"github.com/thediveo/lxkns/cmd/internal/pkg/output"
	"github.com/thediveo/lxkns/cmd/internal/pkg/style"
	"github.com/thediveo/lxkns/model"
)

// UserNSVisitor is an asciitree.Visitor which starts from a list (slice) of
// root user namespaces and then recursively dives into the user namespace
// hierarchy.
type UserNSVisitor struct {
	Details bool
}

// Roots returns the given topmost hierarchical user namespaces sorted.
func (v *UserNSVisitor) Roots(roots reflect.Value) (children []reflect.Value) {
	userroots := lxkns.SortNamespaces(roots.Interface().([]model.Namespace))
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
	if ns, ok := node.Interface().(model.Namespace); ok {
		style := style.Styles[ns.Type().Name()]
		label = fmt.Sprintf("%s%s %s",
			output.NamespaceIcon(ns),
			style.V(ns.(model.NamespaceStringer).TypeIDString()),
			output.NamespaceReferenceLabel(ns))
	}
	if uns, ok := node.Interface().(model.Ownership); ok {
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
// always a user namespace), as well as the list of properties (owned
// non-user namespaces) and the list of child user namespace nodes.
func (v *UserNSVisitor) Get(node reflect.Value) (
	label string, properties []string, children reflect.Value) {
	// Label for this user namespace...
	label = v.Label(node)
	// Children of this user namespace...
	if hns, ok := node.Interface().(model.Hierarchy); ok {
		children = reflect.ValueOf(lxkns.SortChildNamespaces(hns.Children()))
	}
	// Owned (non-user) namespaces...
	if v.Details {
		if uns, ok := node.Interface().(model.Ownership); ok {
			ownedns := uns.Ownings()
			for _, nstype := range model.TypeIndexLexicalOrder {
				if nstype == model.UserNS {
					// The lxkns information model does not add child user
					// namespaces to the model, but instead models the
					// parent-child relationship. So, there should be no owned
					// user namespaces present anyway. But we skip just as a
					// safeguard in case the model would change anytime later.
					continue
				}
				nslist := lxkns.SortedNamespaces(ownedns[nstype])
				for _, ns := range nslist {
					if !filter.Filter(ns) {
						continue
					}
					style := style.Styles[ns.Type().Name()]
					s := fmt.Sprintf("%s%s %s",
						output.NamespaceIcon(ns),
						style.V(ns.(model.NamespaceStringer).TypeIDString()),
						output.NamespaceReferenceLabel(ns))
					properties = append(properties, s)
				}
			}
		}
	}
	return
}
