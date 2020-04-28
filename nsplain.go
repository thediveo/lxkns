// plainNamespace implements the Namespace "base" interface.

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

	"github.com/thediveo/lxkns/nstypes"
	"github.com/thediveo/lxkns/ops"
)

// plainNamespace stores useful information about a concrete Linux kernel
// namespace. It implements the interfaces Namespace, Hierarchy, Ownership,
// and NamespaceStringer. Additionally, it implements the package-private
// interface leaderAdder. (There, I did it. I documented implemented
// interfaces explicitly for clarity.)
type plainNamespace struct {
	nsid      nstypes.NamespaceID
	nstype    nstypes.NamespaceType
	ownernsid nstypes.NamespaceID
	owner     Ownership
	ref       string
	leaders   []*Process
}

var _ Namespace = (*plainNamespace)(nil)
var _ NamespaceStringer = (*plainNamespace)(nil)

// NamespaceConfigurer allows discovery mechanisms to set up the information for
// a namespace. This is a lxkns-internal interface needed by other lxkns
// subpackages.
type NamespaceConfigurer interface {
	AddLeader(proc *Process)               // adds yet another self-styled leader.
	SetRef(string)                         // sets a filesystem path for referencing this namespace.
	DetectOwner(nsf *ops.NamespaceFile)    // detects owning user namespace id.
	SetOwner(usernsid nstypes.NamespaceID) // sets the owning user namespace id directly.
	ResolveOwner(usernsmap NamespaceMap)   // resolves owner ns id into object reference.
}

// ID returns the namespace identifier. This identifier is basically an inode
// number from the special "nsfs" namespace filesystem inside the Linux kernel.
// IDs cannot be set as only the Linux allocates and manages them.
func (pns *plainNamespace) ID() nstypes.NamespaceID { return pns.nsid }

// Type returns the type of namespace in form of one of the NamespaceType, such
// as nstypes.CLONE_NEWNS, nstypes.CLONE_NEWCGROUP, et cetera.
func (pns *plainNamespace) Type() nstypes.NamespaceType { return pns.nstype }

// Owner returns the user namespace "owning" this namespace. According to
// Linux-kernel rules, the owner of a user namespace is the parent of that user
// namespace.
func (pns *plainNamespace) Owner() Ownership { return pns.owner }

// Ref returns a filesystem path suitable for referencing this namespace. A zero
// ref indicates that there is no reference path available: this is the case for
// "hidden" PID and user namespaces sandwiched in between PID or user namespaces
// where reference paths are available, because these other namespaces have
// processes joined to them, or are either bind-mounted or fd-referenced. Hidden
// PID namespaces can appear only when there is no process in any of their child
// namespaces and the child PID namespace(s) is bind-mounted or fd-references
// (the parent PID namespace is then kept alive because the child PID namespaces
// are kept alive).
func (pns *plainNamespace) Ref() string { return pns.ref }

// Leaders returns an unsorted list of Process-es which are joined to this
// namespace and which are the topmost processes in the process tree still
// joined to this namespace.
func (pns *plainNamespace) Leaders() []*Process { return pns.leaders }

// Ealdorman returns the most senior leader process. The "most senior" process
// is the one which was created at the earliest, based on the start times from
// /proc/[PID]/stat.
//
// Me thinks, me has read too many Bernard Cornwell books. Wyrd bið ful aræd.
func (pns *plainNamespace) Ealdorman() (p *Process) {
	// Sorting most probably will be more expensive than a single run through
	// the list, so take it easy without the sort package.
	for _, proc := range pns.leaders {
		if p == nil {
			p = proc
		} else if proc.Starttime < p.Starttime {
			p = proc
		}
	}
	return
}

// LeaderPIDs returns the list of leader PIDs. This is a convenience method for
// those use cases where just a list of leader process PIDs is needed, but not
// the leader Process objects themselves.
func (pns *plainNamespace) LeaderPIDs() []PIDType {
	pids := make([]PIDType, len(pns.leaders))
	for idx, leader := range pns.leaders {
		pids[idx] = leader.PID
	}
	return pids
}

// String describes this instance of a non-hierarchical ("plain") Linux kernel
// namespace.
func (pns *plainNamespace) String() string {
	var s string
	if pns.owner != nil {
		s = fmt.Sprintf("%s, owned by %s",
			pns.TypeIDString(),
			pns.owner.(NamespaceStringer).TypeIDString())
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
func (pns *plainNamespace) TypeIDString() string {
	return fmt.Sprintf("%s:[%d]", pns.nstype.Name(), pns.nsid)
}

// LeaderString returns a textual list of leader process PIDs.
func (pns *plainNamespace) LeaderString() string {
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
func (pns *plainNamespace) AddLeader(proc *Process) {
	for _, leader := range pns.leaders {
		if leader == proc {
			return
		}
	}
	pns.leaders = append(pns.leaders, proc)
}

// SetRef sets a filesystem path to reference this namespace.
func (pns *plainNamespace) SetRef(ref string) {
	pns.ref = ref
}

// DetectOwner gets the ownering user namespace id from Linux, and stores it for
// later resolution, after when we have a complete map of all user namespaces.
func (pns *plainNamespace) DetectOwner(nsf *ops.NamespaceFile) {
	if nsf == nil {
		return
	}
	// The User() call gives us an fd wrapped in an os.File, which we can then
	// ask for its namespace ID.
	usernsf, err := nsf.User()
	if err != nil {
		return
	}
	defer usernsf.Close() // Do NOT leak.
	pns.ownernsid, _ = usernsf.ID()
}

// SetOwner set the namespace ID of the user namespace owning this namespace.
func (pns *plainNamespace) SetOwner(usernsid nstypes.NamespaceID) {
	pns.ownernsid = usernsid
}

// ResolveOwner sets the owning user namespace reference based on the owning
// user namespace id discovered earlier.
func (pns *plainNamespace) ResolveOwner(usernsmap NamespaceMap) {
	pns.resolveOwner(pns, usernsmap)
}

// The internal shared implementation for resolving the owner namespace ID into
// its corresponding user namespace and storing it for this namespace. Because
// Golang doesn't support true inheritance but only a bad surrogate termed
// "embedding", we need to pass in not only the base struct pointer (receiver
// pointer), but also the real interface pointer which can point to a
// "subclass", that is the struct embedding the base class directly or
// indirectly.
func (pns *plainNamespace) resolveOwner(namespace Namespace, usernsmap NamespaceMap) {
	// Only try to resolve when we actually got the user namespace id
	// of the owner, otherwise we must skip resolution.
	if pns.ownernsid != 0 {
		ownerns := usernsmap[pns.ownernsid].(*userNamespace)
		pns.owner = ownerns
		// Do NOT assign the receiver pointer, as this would clamp us to a
		// plainNamespace, which sucks when we're in fact a PID or user
		// namespace. Oh, Golang's resistance to proper inheritance simply sux.
		ownerns.ownedns[TypeIndex(pns.nstype)][pns.nsid] = namespace
	}
}
