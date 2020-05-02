// Defines our model view in terms of Linux kernel namespaces and their
// relationships with other namespaces, as well as processes. This source file
// only defines the namespace model elements, but not any discovery mechanisms
// for them.

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

// +build linux

package lxkns

import (
	"fmt"

	"github.com/thediveo/lxkns/species"
)

// NamespaceTypeIndex is an array index type for Linux kernel namespace types.
// It is used with the AllNamespaces type, which is an array of namespace
// maps, one map "id->namespace object" for each type of Linux kernel
// namespace. NamespaceTypeIndex must not be confused with the Linux' kernel
// namespace clone() syscall constants as typed as NamespaceType instead.
type NamespaceTypeIndex int

// Set of indices into AllNamespaces arrays, one for each type of Linux kernel
// namespace.
const (
	MountNS  NamespaceTypeIndex = iota // array index for mount namespaces map
	CgroupNS                           // array index for cgroup namespaces map
	UTSNS                              // array index for UTS namespaces map
	IPCNS                              // array index for IPC namespaces map
	UserNS                             // array index for user namespaces map
	PIDNS                              // array index for PID namespaces map
	NetNS                              // array index for net namespaces map

	NamespaceTypesCount // number of namespace types
)

// typeIndices maps Linux' kernel namespace clone() syscall constants to
// their corresponding AllNamespaces array indices.
var typeIndices = map[species.NamespaceType]NamespaceTypeIndex{
	species.CLONE_NEWNS:     MountNS,
	species.CLONE_NEWCGROUP: CgroupNS,
	species.CLONE_NEWUTS:    UTSNS,
	species.CLONE_NEWIPC:    IPCNS,
	species.CLONE_NEWUSER:   UserNS,
	species.CLONE_NEWPID:    PIDNS,
	species.CLONE_NEWNET:    NetNS,
}

// TypesByIndex maps Allnamespaces array indices to their corresponding Linux'
// kernel namespace clone() syscall constants.
var TypesByIndex = [NamespaceTypesCount]species.NamespaceType{
	species.CLONE_NEWNS,
	species.CLONE_NEWCGROUP,
	species.CLONE_NEWUTS,
	species.CLONE_NEWIPC,
	species.CLONE_NEWUSER,
	species.CLONE_NEWPID,
	species.CLONE_NEWNET,
}

// TypeIndexLexicalOrder contains Namespace type indices in lexical order.
var TypeIndexLexicalOrder = [NamespaceTypesCount]NamespaceTypeIndex{
	CgroupNS,
	IPCNS,
	MountNS,
	NetNS,
	PIDNS,
	UserNS,
	UTSNS,
}

// TypeIndex returns the AllNamespaces array index corresponding with the
// specified Linux' kernel clone() syscall namespace constant. For instance,
// for CLONE_NEWNET the index NetNS is then returned.
func TypeIndex(nstype species.NamespaceType) NamespaceTypeIndex {
	if idx, ok := typeIndices[nstype]; ok {
		return idx
	}
	return -1 // return an invalid index
}

// AllNamespaces contains separate NamespaceMaps for all types of Linux kernel
// namespaces. This type allows package functions to work on multiple
// namespace types simultaneously in order to optimize traversal of the /proc
// filesystem, bind-mounts, et cetera. AllNamespaces thus stores "all"
// namespaces that could be discovered in the system, subject to discovery
// filtering.
type AllNamespaces [NamespaceTypesCount]NamespaceMap

// NamespacesSet contains a Namespace reference of each type exactly once. For
// instance, it represents the set of 7 namespaces a process will always be
// joined ("attached", ...) to. Processes cannot be not attached to each type
// of Linux kernel namespace.
type NamespacesSet [NamespaceTypesCount]Namespace

// NamespaceMap indexes a bunch of Namespaces by their identifiers. Usually,
// namespace indices will contain only namespaces of the same type.
type NamespaceMap map[species.NamespaceID]Namespace

// Namespace represents a Linux kernel namespace in terms of its unique
// identifier, type, owning user namespace, joined (leader) processes, and some
// more.
type Namespace interface {
	// ID returns the unique identifier of this Linux-kernel namespace. This
	// identifier is basically a tuple consisting of an inode number from the
	// special "nsfs" namespace filesystem inside the Linux kernel, together
	// with the device ID of that nsfs filesystem. IDs cannot be set as only the
	// Linux allocates and manages them.
	ID() species.NamespaceID
	// Type returns the type of namespace in form of one of the NamespaceType, such
	// as species.CLONE_NEWNS, species.CLONE_NEWCGROUP, et cetera.
	Type() species.NamespaceType
	// Owner returns the user namespace "owning" this namespace. For user
	// namespaces, Owner always returns nil; use Hierarchy.Parent() instead, as
	// the owner of a user namespace is its parent user namespace.
	Owner() Ownership
	// Ref returns a filesystem path suitable for referencing this namespace. A zero
	// ref indicates that there is no reference path available: this is the case for
	// "hidden" PID and user namespaces sandwiched in between PID or user namespaces
	// where reference paths are available, because these other namespaces have
	// processes joined to them, or are either bind-mounted or fd-referenced. Hidden
	// PID namespaces can appear only when there is no process in any of their child
	// namespaces and the child PID namespace(s) is bind-mounted or fd-references
	// (the parent PID namespace is then kept alive because the child PID namespaces
	// are kept alive).
	Ref() string
	// Leaders returns an unsorted list of Process-es which are joined to this
	// namespace and which are the topmost processes in the process tree still
	// joined to this namespace.
	Leaders() []*Process
	// LeaderPIDs returns the list of leader PIDs. This is a convenience method for
	// those use cases where just a list of leader process PIDs is needed, but not
	// the leader Process objects themselves.
	LeaderPIDs() []PIDType // "leader" process PIDs only.
	// Ealdorman returns the most senior leader process. The "most senior"
	// process is the one which was created at the earliest, based on the start
	// times from /proc/[PID]/stat. Me thinks, me has read too many Bernard
	// Cornwell books. Wyrd bið ful aræd.
	Ealdorman() *Process
	// String describes this namespace with type, id, joined leader processes,
	// and optionally information about owner, children, parent.
	String() string
}

// NamespaceStringer describes a namespace either in its descriptive form when
// using the well-known String() method, or in a terse format when going for
// TypeIDString(), which only describes the type and identifier of a
// namespace.
type NamespaceStringer interface {
	fmt.Stringer
	// TypeIDString describes this instance of a Linux kernel namespace just by
	// its type and identifier, and nothing else.
	TypeIDString() string
}

// Hierarchy informs about the parent-child relationships of PID and user
// namespaces.
type Hierarchy interface {
	// Parent returns the parent user or PID namespace of this user or PID
	// namespace. If there is no parent namespace or the parent namespace in
	// inaccessible, then Parent returns nil.
	Parent() Hierarchy
	// Children returns a list of child PID or user namespaces for this PID or
	// user namespace.
	Children() []Hierarchy
}

// Ownership informs about the owning user ID, as well as the namespaces owned
// by a specific user namespace. Only user namespaces can execute Ownership.
type Ownership interface {
	// UID returns the user ID of the process that created this user namespace.
	UID() int
	// Ownings returns all namespaces owned by this user namespace, with the
	// exception of user namespaces. "Owned" user namespaces are actually child
	// user namespaces, so they are returned through Hierarchy.Children()
	// instead.
	Ownings() AllNamespaces
}

// SloppyByIno looks up a namespace given only its inode number, lacking the
// device ID which otherwise is necessary to unambiguously identify any
// particular namespace. In order to avoid an expensive linear search through
// the namespace map, SloppyByIno glances the device ID from an arbitrary map
// entry and then tries to look up the namespace in question. This assumes that
// a NamespaceMap stores only a single type and the caller absolutely
// understands the limitations of this function.
func (nsm NamespaceMap) SloppyByIno(nsid species.NamespaceID) Namespace {
	if nsid.Dev != 0 {
		return nsm[nsid]
	}
	for mnsid := range nsm {
		return nsm[species.NamespaceID{Dev: mnsid.Dev, Ino: nsid.Ino}]
	}
	return nil
}

// NewNamespace returns a new zero'ed namespace object suitable for the
// specified type of namespace. Now this is a real-world case where the
// "nongonformist" rule of "accept interfaces, return structs" doesn't make
// sense, because struct types don't support polymorphism. On the other hand,
// thousands of blog posts and SO answers cannot be wrong, more so, the more
// upvotes they accumulated ;)
func NewNamespace(nstype species.NamespaceType, nsid species.NamespaceID, ref string) Namespace {
	switch nstype {
	case species.CLONE_NEWUSER:
		// Someone please tell me that golang actually makes sense... at least
		// some quantum of sense. Hmm, could be the title of next summer's
		// blockbuster: "A Quantum of Sense". Erm, no. Won't ever fly in some
		// states.
		user := &userNamespace{
			hierarchicalNamespace: hierarchicalNamespace{
				plainNamespace: plainNamespace{
					nsid:   nsid,
					nstype: nstype,
					ref:    ref,
				},
			},
		}
		for idx := range user.ownedns {
			user.ownedns[idx] = NamespaceMap{}
		}
		return user
	case species.CLONE_NEWPID:
		return &hierarchicalNamespace{
			plainNamespace: plainNamespace{
				nsid:   nsid,
				nstype: nstype,
				ref:    ref,
			},
		}
	default:
		return &plainNamespace{nsid: nsid, nstype: nstype, ref: ref}
	}
}
