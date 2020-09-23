package namespaces

import (
	"fmt"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// HierarchicalNamespace stores hierarchy information in addition to the
// information for plain namespaces. Besides the interfaces for a
// plainNamespace, it additionally implements the public Hierarchy interface.
type HierarchicalNamespace struct {
	PlainNamespace
	parent   model.Hierarchy
	children []model.Hierarchy
}

// Ensure that our "class" *does* implement the required interfaces.
var (
	_ model.Namespace         = (*HierarchicalNamespace)(nil)
	_ model.NamespaceStringer = (*HierarchicalNamespace)(nil)
	_ model.Hierarchy         = (*HierarchicalNamespace)(nil)
	_ NamespaceConfigurer     = (*HierarchicalNamespace)(nil)
	_ HierarchyConfigurer     = (*HierarchicalNamespace)(nil)
)

// Parent returns the parent user or PID namespace of this user or PID
// namespace. If there is no parent namespace or the parent namespace in
// inaccessible, then Parent returns nil.
func (hns *HierarchicalNamespace) Parent() model.Hierarchy { return hns.parent }

// Children returns a list of child PID or user namespaces for this PID or
// user namespace.
func (hns *HierarchicalNamespace) Children() []model.Hierarchy { return hns.children }

// String describes this instance of a hierarchical Linux kernel namespace,
// with its parent and children (but not grand-children). This description is
// non-recursive.
func (hns *HierarchicalNamespace) String() string {
	return fmt.Sprintf("%s, %s",
		hns.PlainNamespace.String(), hns.ParentChildrenString())
}

// ParentChildrenString just describes the parent and children of a
// hierarchical Linux kernel namespace, in a non-recursive form.
func (hns *HierarchicalNamespace) ParentChildrenString() string {
	var parent, children string
	// Who is our parent?
	if hns.parent == nil {
		parent = "none"
	} else {
		parent = hns.parent.(model.NamespaceStringer).TypeIDString()
	}
	// Who are our children?
	if len(hns.children) == 0 {
		children = "none"
	} else {
		c := make([]string, len(hns.children))
		for idx, child := range hns.children {
			c[idx] = child.(model.NamespaceStringer).TypeIDString()
		}
		children = "[" + strings.Join(c, ", ") + "]"
	}
	return fmt.Sprintf("parent %s, children %s", parent, children)
}

// AddChild adds a child namespace to this (parent) namespace. It panics in
// case a child is tried to be added twice to either the same parent or
// different parents.
func (hns *HierarchicalNamespace) AddChild(child model.Hierarchy) {
	child.(HierarchyConfigurer).SetParent(hns)
	hns.children = append(hns.children, child)
}

// SetParent sets the parent namespace of this child namespace. It panics in
// case the parent would change.
func (hns *HierarchicalNamespace) SetParent(parent model.Hierarchy) {
	if hns.parent != nil {
		panic("trying to change parents might sometimes not a good idea, especially just now.\n" +
			"parent: " + parent.(model.NamespaceStringer).String() + "\n" +
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
func (hns *HierarchicalNamespace) ResolveOwner(usernsmap model.NamespaceMap) {
	hns.resolveOwner(hns, usernsmap)
}
