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

	"github.com/thediveo/go-asciitree/v2"
	"github.com/thediveo/lxkns/cmd/cli/style"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/internal/xslices"
	"github.com/thediveo/lxkns/internal/xstrings"
	"github.com/thediveo/lxkns/model"
)

// UserNSVisitor is an asciitree.Visitor which starts from a list (slice) of
// root user namespaces and then recursively dives into the user namespace
// hierarchy.
type UserNSVisitor struct {
	// when true, also show all owned non-user namespaces.
	Details bool
	// configured namespace filter function.
	Filter func(model.Namespace) bool
	// render function for namespace icons, where its exact behavior depends on
	// CLI flags.
	NamespaceIcon func(model.Namespace) string
	// render function for namespace references in form of either process names
	// (as well as additional process properties) or file system references.
	NamespaceReferenceLabel func(model.Namespace) string
}

var _ asciitree.Visitor = (*UserNSVisitor)(nil)

// Roots returns the given topmost hierarchical user namespaces sorted.
func (v *UserNSVisitor) Roots(roots any) []any {
	rootNamespaces, _ := roots.([]model.Namespace)
	discover.SortNamespaces(rootNamespaces)
	return xslices.Any(rootNamespaces)
}

// Label returns the text label for a namespace node. Everything else will have
// no label.
func (v *UserNSVisitor) Label(node any) (label string) {
	if ns, ok := node.(model.Namespace); ok {
		style := style.Styles[ns.Type().Name()]
		label = xstrings.Join(
			v.NamespaceIcon(ns)+
				style.V(ns.(model.NamespaceStringer).TypeIDString()).String(),
			v.NamespaceReferenceLabel(ns))
	}
	// If it is a user namespace we render information about the user that
	// created this particular user namespace: the user ID and, if available,
	// the user name.
	if uns, ok := node.(model.Ownership); ok {
		label = xstrings.Join(label, fmt.Sprintf("created by UID %d",
			style.OwnerStyle.V(uns.UID())))
		if user, err := user.LookupId(fmt.Sprintf("%d", uns.UID())); err == nil {
			label += fmt.Sprintf(" (%q)", style.OwnerStyle.V(user.Username))
		}
	}
	return
}

// Get returns the user namespace text label for the current node (which is
// always a user namespace), as well as the list of properties (owned
// non-user namespaces) and the list of child user namespace nodes.
func (v *UserNSVisitor) Get(node any) (label string, properties []string, children []any) {
	// Determine the label text for this user namespace.
	label = v.Label(node)
	// Determine the children of this user namespace, which are in turn user
	// namespaces.
	if hierns, ok := node.(model.Hierarchy); ok {
		children = xslices.Any(discover.SortChildNamespaces(hierns.Children()))
	}
	// In case a detailed tree has been requested, determine the owned non-user
	// namespaces.
	if v.Details {
		if userns, ok := node.(model.Ownership); ok {
			ownedns := userns.Ownings()
			for _, nstype := range model.TypeIndexLexicalOrder {
				if nstype == model.UserNS {
					// The lxkns information model does not add child user
					// namespaces to the model, but instead models the
					// parent-child relationship. So, there should be no owned
					// user namespaces present anyway. But we skip just as a
					// safeguard in case the model would change anytime later.
					continue
				}
				nslist := discover.SortedNamespaces(ownedns[nstype])
				for _, ns := range nslist {
					if !v.Filter(ns) {
						continue
					}
					style := style.Styles[ns.Type().Name()]
					s := fmt.Sprintf("%s%s %s",
						v.NamespaceIcon(ns),
						style.V(ns.(model.NamespaceStringer).TypeIDString()),
						v.NamespaceReferenceLabel(ns))
					properties = append(properties, s)
				}
			}
		}
	}
	return
}
