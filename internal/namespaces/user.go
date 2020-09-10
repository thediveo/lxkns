package namespaces

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/relations"
)

// UserNamespace stores ownership information in addition to the information
// for hierarchical namespaces. On top of the interfaces supported by a
// hierarchicalNamespace, UserNamespace implements the Ownership interface.
type UserNamespace struct {
	HierarchicalNamespace
	owneruid int
	ownedns  model.AllNamespaces
}

var (
	_ model.Namespace         = (*UserNamespace)(nil)
	_ model.NamespaceStringer = (*UserNamespace)(nil)
	_ model.Hierarchy         = (*UserNamespace)(nil)
	_ model.Ownership         = (*UserNamespace)(nil)
)

func (uns *UserNamespace) UID() int                     { return uns.owneruid }
func (uns *UserNamespace) Ownings() model.AllNamespaces { return uns.ownedns }

// String describes this instance of a user namespace, with its parent,
// children, and owned namespaces. This description is non-recursive.
func (uns *UserNamespace) String() string {
	u, err := user.LookupId(fmt.Sprintf("%d", uns.owneruid))
	var userstr string
	if err == nil {
		userstr = fmt.Sprintf(" (%q)", u.Username)
	}
	owneds := ""
	var o []string
	for _, ownedbytype := range uns.ownedns {
		for _, owned := range ownedbytype {
			o = append(o, owned.(model.NamespaceStringer).TypeIDString())
		}
	}
	if len(o) != 0 {
		owneds = ", owning [" + strings.Join(o, ", ") + "]"
	}
	parentandchildren := uns.ParentChildrenString()
	leaders := uns.LeaderString()
	if leaders != "" {
		leaders = ", " + leaders
	}
	return fmt.Sprintf("%s, created by UID %d%s%s, %s%s",
		uns.TypeIDString(),
		uns.owneruid, userstr,
		leaders,
		parentandchildren,
		owneds)
}

// DetectUIDs takes an open file referencing a user namespace to query its
// owner's UID and then stores it for this user namespace proxy.
func (uns *UserNamespace) DetectUID(nsref relations.Relation) {
	uns.owneruid, _ = nsref.OwnerUID()
}

// ResolveOwner sets the owning user namespace reference based on the owning
// user namespace id discovered earlier. Yes, we're repeating us ourselves with
// this method, because Golang is self-inflicted pain when trying to emulate
// inheritance using embedding ... note: it doesn't work correctly. The reason
// is that we need the use the correct instance pointer and not a pointer to an
// embedded instance when setting the "owned" relationship.
func (uns *UserNamespace) ResolveOwner(usernsmap model.NamespaceMap) {
	uns.resolveOwner(uns, usernsmap)
}

// AddChild adds a child namespace to this (parent) namespace. It panics in case
// a child is tried to be added twice to either the same parent or different
// parents.
//
// Note: we must reimplement this method here, as otherwise Golang will totally
// fubar because it calls the embedded hierarchicalNamespace.AddChild and will
// then set us not as a userNamespace parent, but instead as a
// hierarchicalNamespace parent. If this Golang design isn't a fubar, then I
// really don't know what a fubar is.
func (uns *UserNamespace) AddChild(child model.Hierarchy) {
	child.(HierarchyConfigurer).SetParent(uns)
	uns.children = append(uns.children, child)
}
