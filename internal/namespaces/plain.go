package namespaces

import (
	"fmt"
	"io"
	"strings"

	"github.com/thediveo/lxkns/model"
	r "github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
)

// PlainNamespace stores useful information about a concrete Linux kernel
// namespace. It implements the interfaces Namespace, Hierarchy, Ownership,
// and NamespaceStringer. Additionally, it implements the package-private
// interface leaderAdder. (There, I did it. I documented implemented
// interfaces explicitly for clarity.)
type PlainNamespace struct {
	nsid      species.NamespaceID
	nstype    species.NamespaceType
	ownernsid species.NamespaceID
	owner     model.Ownership
	ref       string
	leaders   []*model.Process
}

// Ensure that our "class" *does* implement the required interfaces.
var (
	_ model.Namespace         = (*PlainNamespace)(nil)
	_ model.NamespaceStringer = (*PlainNamespace)(nil)
	_ NamespaceConfigurer     = (*PlainNamespace)(nil)
)

// ID returns the namespace identifier. This identifier is basically a tuple
// made of an inode number from the special "nsfs" namespace filesystem inside
// the Linux kernel, together with the device ID of the nsfs. IDs cannot be
// set as only the Linux allocates and manages them.
func (pns *PlainNamespace) ID() species.NamespaceID { return pns.nsid }

// Type returns the type of namespace in form of one of the NamespaceType,
// such as species.CLONE_NEWNS, species.CLONE_NEWCGROUP, et cetera.
func (pns *PlainNamespace) Type() species.NamespaceType { return pns.nstype }

// Owner returns the user namespace "owning" this namespace. According to
// Linux-kernel rules, the owner of a user namespace is the parent of that
// user namespace.
func (pns *PlainNamespace) Owner() model.Ownership { return pns.owner }

// Ref returns a filesystem path suitable for referencing this namespace. A
// zero ref indicates that there is no reference path available: this is the
// case for "hidden" PID and user namespaces sandwiched in between PID or user
// namespaces where reference paths are available, because these other
// namespaces have processes joined to them, or are either bind-mounted or
// fd-referenced. Hidden PID namespaces can appear only when there is no
// process in any of their child namespaces and the child PID namespace(s) is
// bind-mounted or fd-references (the parent PID namespace is then kept alive
// because the child PID namespaces are kept alive).
func (pns *PlainNamespace) Ref() string { return pns.ref }

// Leaders returns an unsorted list of Process-es which are joined to this
// namespace and which are the topmost processes in the process tree still
// joined to this namespace.
func (pns *PlainNamespace) Leaders() []*model.Process { return pns.leaders }

// Ealdorman returns the most senior leader process. The "most senior" process
// is the one which was created at the earliest, based on the start times from
// /proc/[PID]/stat.
//
// Me thinks, me has read too many Bernard Cornwell books. Wyrd bið ful aræd.
func (pns *PlainNamespace) Ealdorman() (p *model.Process) {
	// Sorting most probably will be more expensive than a single run through
	// the list, so take it easy without the sort package.
	for _, proc := range pns.leaders {
		if p == nil {
			p = proc
		} else if proc.Starttime < p.Starttime {
			p = proc
		} else if proc.Starttime == p.Starttime && proc.PID < p.PID {
			// Ensure stable results in case two processes have the exactly
			// same start time, as will be the case for the initial process 1
			// and kthredd 2. Otherwise, as the list of leader PIDs isn't
			// sorted, we would end up with non-deterministic ealdormen; this
			// is because the process table is a Golang map and we collect the
			// leader processes by iterating over this process table map (and
			// seeing which parent is the topmost parent still in the same
			// namespace). And Golang maps randomize iteration order.
			p = proc
		}
	}
	return
}

// LeaderPIDs returns the list of leader PIDs. This is a convenience method
// for those use cases where just a list of leader process PIDs is needed, but
// not the leader Process objects themselves.
func (pns *PlainNamespace) LeaderPIDs() []model.PIDType {
	pids := make([]model.PIDType, len(pns.leaders))
	for idx, leader := range pns.leaders {
		pids[idx] = leader.PID
	}
	return pids
}

// String describes this instance of a non-hierarchical ("plain") Linux kernel
// namespace.
func (pns *PlainNamespace) String() string {
	var s string
	if pns.owner != nil {
		s = fmt.Sprintf("%s, owned by %s",
			pns.TypeIDString(),
			pns.owner.(model.NamespaceStringer).TypeIDString())
	} else {
		s = pns.TypeIDString()
	}
	if l := pns.LeaderString(); l != "" {
		s += ", " + l
	}
	return s
}

// TypeIDString describes this instance of a Linux kernel namespace just by
// its type and identifier, and nothing else.
func (pns *PlainNamespace) TypeIDString() string {
	return fmt.Sprintf("%s:[%d]", pns.nstype.Name(), pns.nsid.Ino)
}

// LeaderString returns a textual list of leader process PIDs.
func (pns *PlainNamespace) LeaderString() string {
	if len(pns.leaders) == 0 {
		return ""
	}
	leaders := []string{}
	for _, leader := range pns.leaders {
		leaders = append(leaders, fmt.Sprintf("%q (%d)", leader.Name, leader.PID))
	}
	return fmt.Sprintf("joined by %s", strings.Join(leaders, ", "))
}

// AddLeader joins another leader process to the lot of leaders in this
// namespace. It ensures that each leader appears only once in the list, even
// if AddLeader is called multiple times for the same leader process.
func (pns *PlainNamespace) AddLeader(proc *model.Process) {
	for _, leader := range pns.leaders {
		if leader == proc && leader.PID == proc.PID {
			return
		}
	}
	pns.leaders = append(pns.leaders, proc)
}

// SetRef sets a filesystem path to reference this namespace.
func (pns *PlainNamespace) SetRef(ref string) {
	pns.ref = ref
}

// DetectOwner gets the ownering user namespace id from Linux, and stores it
// for later resolution, after when we have a complete map of all user
// namespaces.
func (pns *PlainNamespace) DetectOwner(nsr r.Relation) {
	if nsr == nil {
		return
	}
	// The User() call gives us an fd wrapped in an os.File, which we can then
	// ask for its namespace ID.
	usernsf, err := nsr.User()
	if err != nil {
		return
	}
	// Do not leak, so release user namespace (file) reference now, as we're
	// done using it. And yes, we're blindly type asserting here, so the
	// caller must pass in a closeable namespace reference object.
	pns.ownernsid, _ = usernsf.ID()
	usernsf.(io.Closer).Close()
}

// SetOwner set the namespace ID of the user namespace owning this namespace.
func (pns *PlainNamespace) SetOwner(usernsid species.NamespaceID) {
	pns.ownernsid = usernsid
}

// ResolveOwner sets the owning user namespace reference based on the owning
// user namespace id discovered earlier.
func (pns *PlainNamespace) ResolveOwner(usernsmap model.NamespaceMap) {
	pns.resolveOwner(pns, usernsmap)
}

// The internal shared implementation for resolving the owner namespace ID
// into its corresponding user namespace and storing it for this namespace.
// Because Golang doesn't support true inheritance but only a bad surrogate
// termed "embedding", we need to pass in not only the base struct pointer
// (receiver pointer), but also the real interface pointer which can point to
// a "subclass", that is the struct embedding the base class directly or
// indirectly.
func (pns *PlainNamespace) resolveOwner(namespace model.Namespace, usernsmap model.NamespaceMap) {
	// Only try to resolve when we actually got the user namespace id of the
	// owner, otherwise we must skip resolution. But even with a user
	// namespace ID, this might have become stale during the discovery
	// process, as discovery cannot be atomic. And we're not starting to worry
	// about how the nsfs reuses namespace ids, no, no, no ... argh!!!
	if pns.ownernsid != species.NoneID {
		owns := usernsmap[pns.ownernsid]
		if owns == nil {
			return
		}
		ownerns := owns.(*UserNamespace)
		pns.owner = ownerns
		// Do NOT assign the receiver pointer, as this would clamp us to a
		// plainNamespace, which sucks when we're in fact a PID or user
		// namespace. Oh, Golang's resistance to proper inheritance simply
		// sux.
		ownerns.ownedns[model.TypeIndex(pns.nstype)][pns.nsid] = namespace
	}
}
