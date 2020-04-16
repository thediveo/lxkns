// hierarchicalNamespace implements the Hierarchy interface for PID and user
// namespaces.

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

package lxkns

import (
	"fmt"
	"strings"
)

// hierarchicalNamespace stores hierarchy information in addition to the
// information for plain namespaces. Besides the interfaces for a
// plainNamespace, it additionally implements the public Hierarchy interface.
type hierarchicalNamespace struct {
	plainNamespace
	parent   Hierarchy
	children []Hierarchy
}

var _ Hierarchy = (*hierarchicalNamespace)(nil)
var _ Hierarchy = (*userNamespace)(nil)

// HierarchyConfigurer allows discovery mechanisms to configure the
// information hold by hierarchical namespaces.
type HierarchyConfigurer interface {
	AddChild(child Hierarchy)
	SetParent(parent Hierarchy)
}

func (hns *hierarchicalNamespace) Parent() Hierarchy     { return hns.parent }
func (hns *hierarchicalNamespace) Children() []Hierarchy { return hns.children }

// String describes this instance of a hierarchical Linux kernel namespace,
// with its parent and children (but not grand-children). This description is
// non-recursive.
func (hns *hierarchicalNamespace) String() string {
	return fmt.Sprintf("%s, %s",
		hns.plainNamespace.String(), hns.ParentChildrenString())
}

// ParentChildrenString just describes the parent and children of a
// hierarchical Linux kernel namespace, in a non-recursive form.
func (hns *hierarchicalNamespace) ParentChildrenString() string {
	var parent, children string
	// Who is our parent?
	if hns.parent == nil {
		parent = "none"
	} else {
		parent = hns.parent.(NamespaceStringer).TypeIDString()
	}
	// Who are our children?
	if len(hns.children) == 0 {
		children = "none"
	} else {
		c := make([]string, len(hns.children))
		for idx, child := range hns.children {
			c[idx] = child.(NamespaceStringer).TypeIDString()
		}
		children = "[" + strings.Join(c, ", ") + "]"
	}
	return fmt.Sprintf("parent %s, children %s", parent, children)
}

// AddChild adds a child namespace to this (parent) namespace. It panics in
// case a child is tried to be added twice to either the same parent or
// different parents.
func (hns *hierarchicalNamespace) AddChild(child Hierarchy) {
	child.(HierarchyConfigurer).SetParent(hns)
	hns.children = append(hns.children, child)
}

// SetParent sets the parent namespace of this child namespace. It panics in
// case the parent would change.
func (hns *hierarchicalNamespace) SetParent(parent Hierarchy) {
	if hns.parent != nil {
		panic("trying to change parents might sometimes not a good idea, especially just now.\n" +
			"parent: " + parent.(NamespaceStringer).String() + "\n" +
			"child: " + hns.String())
	}
	hns.parent = parent
}

// ResolveOwner sets the owning user namespace reference based on the owning
// user namespace id discovered earlier. Yes, we're repeating us ourselves with
// this method, because Golang is self-inflicted pain when trying to emulate
// inheritance using embedding ... note: it doesn't work correctly. The reason
// is that we need the use the correct instance pointer and not a pointer to an
// embedded instance when setting the "owned" relationship.
func (hns *hierarchicalNamespace) ResolveOwner(usernsmap NamespaceMap) {
	hns.resolveOwner(hns, usernsmap)
}
