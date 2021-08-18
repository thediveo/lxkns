package namespaces

import (
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
)

// NamespaceConfigurer allows discovery and unmarshalling mechanisms to set up
// the information for a namespace. This is a lxkns-internal interface needed
// by other lxkns subpackages.
type NamespaceConfigurer interface {
	AddLeader(proc *model.Process)             // adds yet another self-styled leader.
	SetRef(ref model.NamespaceRef)             // sets a filesystem path for referencing this namespace.
	DetectOwner(nsr relations.Relation)        // detects owning user namespace id.
	SetOwner(usernsid species.NamespaceID)     // sets the owning user namespace id directly.
	ResolveOwner(usernsmap model.NamespaceMap) // resolves owner ns id into object reference.
}

// HierarchyConfigurer allows discovery and unmarshalling mechanisms to
// configure the information hold by hierarchical namespaces.
type HierarchyConfigurer interface {
	AddChild(child model.Hierarchy)
	SetParent(parent model.Hierarchy)
}

// UserConfigurer allows discovery and unmarshalling mechanisms to configure
// the information hold by user namespaces.
type UserConfigurer interface {
	SetOwnerUID(uid int)
}
