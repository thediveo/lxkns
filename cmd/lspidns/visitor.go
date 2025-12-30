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
	"maps"
	"os/user"
	"slices"
	"strings"

	"github.com/thediveo/go-asciitree/v2"
	"github.com/thediveo/lxkns/cmd/cli/style"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/internal/xslices"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// PIDNSVisitor is an asciitree.Visitor which starts from a list (slice) of root
// PID namespaces and then recursively dives into the PID namespace hierarchy,
// optionally showing the intermediate user namespaces owing PID namespaces.
type PIDNSVisitor struct {
	// if true, show the (owning) user namespaces intertwined with the PID
	// namespaces; otherwise, don't show any owning user namespaces for the PID
	// namespaces.
	ShowUserNamespaces bool
	// render function for namespace icons, where its exact behavior depends on
	// CLI flags.
	NamespaceIcon func(model.Namespace) string
	// render function for namespace references in form of either process names
	// (as well as additional process properties) or file system references.
	NamespaceReferenceLabel func(model.Namespace) string
}

var _ asciitree.Visitor = (*PIDNSVisitor)(nil)

// Roots returns the given topmost hierarchical PID namespaces sorted. Well, to
// be honest, it returns the owning user namespaces instead in case user
// namespaces should be shown too.
func (v *PIDNSVisitor) Roots(roots any) []any {
	if v.ShowUserNamespaces {
		// When showing the owning user namespaces in the tree, find the
		// (unique) user namespaces for all PID "root" namespaces and start with
		// these user namespaces instead of the root PID namespaces.
		uniqueUserNamespaces := map[species.NamespaceID]model.Namespace{}
		for _, pidns := range roots.([]model.Namespace) {
			userns := pidns.Owner().(model.Namespace)
			// we don't buffer with checking if we've already seen this owning
			// user namespace, just put it into the map, overwriting itself in
			// case.
			uniqueUserNamespaces[userns.ID()] = userns
		}
		return xslices.Any(discover.SortNamespaces(
			slices.Collect(maps.Values(uniqueUserNamespaces))))
	}
	// We are asked for the conventional way of only showing PID namespaces: for
	// this, we sort all PID "root" namespaces numerically by their IDs and then
	// visit them, descending down the hierarchy. Usually, there should be only
	// a single PID root namespace found though.
	rootNamespaces, _ := roots.([]model.Namespace)
	discover.SortNamespaces(rootNamespaces)
	return xslices.Any(rootNamespaces)
}

// Label returns the text label for a namespace node. Everything else will have
// no label.
func (v *PIDNSVisitor) Label(node any) (label string) {
	var b strings.Builder
	if ns, ok := node.(model.Namespace); ok {
		style := style.Styles[ns.Type().Name()]
		b.WriteString(v.NamespaceIcon(ns))
		b.WriteString(style.V(ns.(model.NamespaceStringer).TypeIDString()).String())
		b.WriteRune(' ')
		b.WriteString(v.NamespaceReferenceLabel(ns))
	}
	if uns, ok := node.(model.Ownership); ok {
		b.WriteString(" created by UID ")
		b.WriteString(style.OwnerStyle.V(uns.UID()).String())
		if user, err := user.LookupId(fmt.Sprintf("%d", uns.UID())); err == nil {
			b.WriteRune(' ')
			b.WriteRune('(')
			b.WriteString(style.OwnerStyle.V(user.Username).String())
			b.WriteRune(')')
		}
	}
	return b.String()
}

// Get returns the user namespace text label for the current node (which is
// always an user namespace), as well as the list of properties (owned
// non-user namespaces) and the list of child user namespace nodes.
func (v *PIDNSVisitor) Get(node any) (label string, properties []string, children []any) {
	// For a user namespace, its children are the owned PID namespaces ... but
	// ... we must only take the topmost owned PID namespaces, otherwise the
	// result isn't exactly correct and we would all subtrees mirrored to the
	// topmost level. Now, a "topmost" owned PID namespace is one that either
	// has no parent PID namespace, or the parent PID namespace has a different
	// owner. That's all that's to it.
	if userNamespace, ok := node.(model.Ownership); ok {
		children := []model.Namespace{}
		for _, namespace := range userNamespace.Ownings()[model.PIDNS] {
			pidNamespace := namespace.(model.Hierarchy)
			parentPIDNamespace := pidNamespace.Parent()
			if parentPIDNamespace == nil || parentPIDNamespace.(model.Namespace).Owner() != userNamespace {
				children = append(children, namespace)
			}
		}
		return v.Label(node), nil, xslices.Any(discover.SortNamespaces(children))
	}
	// For a PID namespace, the children are either PID namespaces, or user
	// namespaces in case a child PID namespace lives in a different user
	// namespace.
	hierarchicalNamespace, ok := node.(model.Hierarchy)
	if !ok {
		// something strange going on here, we have a namespace node that is a
		// flat namespace ... and this should not be the case! Gracefully exit
		// stage left/right.
		return v.Label(node), nil, nil
	}
	if v.ShowUserNamespaces {
		// Insert user namespaces into the PID namespace hierarchy, whenever
		// there is a change of user namespaces in the PID hierarchy.
		userNamespace := node.(model.Namespace).Owner()
		for _, childPIDNamespace := range discover.SortChildNamespaces(hierarchicalNamespace.Children()) {
			if ownerns := childPIDNamespace.(model.Namespace).Owner(); ownerns != userNamespace {
				// The child PID namespace is in a different user namespace, so
				// we take the child's user namespace instead and will visit the
				// child PID namespace only later via the user namespace.
				children = append(children, ownerns)
				continue
			}
			// The child PID namespace is still in the same user namespace, so
			// we take it as a direct child.
			children = append(children, childPIDNamespace)
		}
		return v.Label(node), nil, children
	}
	// Show only the PID namespace hierarchy: this is easy, as we all we need to
	// do is to take all child PID namespaces and return them. That's it. Done.
	return v.Label(node), nil, xslices.Any(
		discover.SortChildNamespaces(hierarchicalNamespace.Children()))
}
