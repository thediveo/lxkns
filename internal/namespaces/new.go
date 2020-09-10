package namespaces

import (
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// New returns a new zero'ed namespace object suitable for the specified type
// of namespace. Now this is a real-world case where the "nongonformist" rule
// of "accept interfaces, return structs" doesn't make sense, because struct
// types don't support polymorphism. On the other hand, thousands of blog
// posts and SO answers cannot be wrong, more so, the more upvotes they
// accumulated ;)
func New(nstype species.NamespaceType, nsid species.NamespaceID, ref string) model.Namespace {
	switch nstype {
	case species.CLONE_NEWUSER:
		// Someone please tell me that golang actually makes sense... at least
		// some quantum of sense. Hmm, could be the title of next summer's
		// blockbuster: "A Quantum of Sense". Erm, no. Won't ever fly in some
		// states.
		user := &UserNamespace{
			HierarchicalNamespace: HierarchicalNamespace{
				PlainNamespace: PlainNamespace{
					nsid:   nsid,
					nstype: nstype,
					ref:    ref,
				},
			},
		}
		for idx := range user.ownedns {
			user.ownedns[idx] = model.NamespaceMap{}
		}
		return user
	case species.CLONE_NEWPID:
		return &HierarchicalNamespace{
			PlainNamespace: PlainNamespace{
				nsid:   nsid,
				nstype: nstype,
				ref:    ref,
			},
		}
	default:
		return &PlainNamespace{nsid: nsid, nstype: nstype, ref: ref}
	}
}
