package namespaces

import (
	"github.com/thediveo/lxkns/model"
	r "github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
)

// NamespaceConfigurer allows discovery mechanisms to set up the information for
// a namespace. This is a lxkns-internal interface needed by other lxkns
// subpackages.
type NamespaceConfigurer interface {
	AddLeader(proc *model.Process)             // adds yet another self-styled leader.
	SetRef(string)                             // sets a filesystem path for referencing this namespace.
	DetectOwner(nsr r.Relation)                // detects owning user namespace id.
	SetOwner(usernsid species.NamespaceID)     // sets the owning user namespace id directly.
	ResolveOwner(usernsmap model.NamespaceMap) // resolves owner ns id into object reference.
}
