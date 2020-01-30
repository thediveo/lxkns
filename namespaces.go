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
	"os"
	"os/user"
	"strings"

	"github.com/thediveo/lxkns/nstypes"
	rel "github.com/thediveo/lxkns/relations"
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
var typeIndices = map[nstypes.NamespaceType]NamespaceTypeIndex{
	nstypes.CLONE_NEWNS:     MountNS,
	nstypes.CLONE_NEWCGROUP: CgroupNS,
	nstypes.CLONE_NEWUTS:    UTSNS,
	nstypes.CLONE_NEWIPC:    IPCNS,
	nstypes.CLONE_NEWUSER:   UserNS,
	nstypes.CLONE_NEWPID:    PIDNS,
	nstypes.CLONE_NEWNET:    NetNS,
}

// TypesByIndex maps Allnamespaces array indices to their corresponding Linux'
// kernel namespace clone() syscall constants.
var TypesByIndex = [NamespaceTypesCount]nstypes.NamespaceType{
	nstypes.CLONE_NEWNS,
	nstypes.CLONE_NEWCGROUP,
	nstypes.CLONE_NEWUTS,
	nstypes.CLONE_NEWIPC,
	nstypes.CLONE_NEWUSER,
	nstypes.CLONE_NEWPID,
	nstypes.CLONE_NEWNET,
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
func TypeIndex(nstype nstypes.NamespaceType) NamespaceTypeIndex {
	return typeIndices[nstype]
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
// namespace indices will contain only namespaces of the same type. However,
// since the Linux kernel uses inode numbers from the special "nsfs"
// filesystem, it is guaranteed that there are never two namespaces of
// different type with the same identifier (=inode number).
type NamespaceMap map[nstypes.NamespaceID]Namespace

// Namespace represents a Linux kernel namespace in terms of its unique
// identifier, type, owning user namespace, et cetera.
type Namespace interface {
	ID() nstypes.NamespaceID     // unique identifier of this Linux kernel namespace.
	Type() nstypes.NamespaceType // type of namespace.
	Owner() Hierarchy            // user namespace "owning" this namespace.
	Ref() string                 // reference in form of a file system path.
	Leaders() []*Process         // "leader" process(es) "inside" this namespace.
	LeaderPIDs() []PIDType       // "leader" process PIDs only.
	String() string              // convenience; maps to long representation form.
}

// NamespaceStringer describes a namespace either in its descriptive form when
// using the well-known String() method, or in a terse format when going for
// TypeIDString(), which only describes the type and identifier of a
// namespace.
type NamespaceStringer interface {
	fmt.Stringer
	TypeIDString() string
}

// Hierarchy informs about the parent-child relationships of PID and user
// namespaces.
type Hierarchy interface {
	Parent() Hierarchy     // parent namespace of this namespace.
	Children() []Hierarchy // child namespaces, if any.
}

// Ownership informs about the owning user ID, as well as the namespaces owned
// by a specific user namespace. Only user namespaces can execute Ownership.
type Ownership interface {
	UID() int             // the user ID of the process that created this user namespace.
	Owned() AllNamespaces // all owned namespaces, except for child user namespaces.
}

// NewNamespace returns a new zero'ed namespace objects suitable for the
// specified type of namespace. Oh, this is a case where the "nongonformist"
// rule of "accept interfaces, return structs" doesn't make sense, because
// struct types don't support polymorphism. On the other hand, thousands of
// blog posts and SO answers cannot be wrong, more so, the more upvotes they
// accumulated ;)
func NewNamespace(nstype nstypes.NamespaceType, nsid nstypes.NamespaceID, ref string) Namespace {
	switch nstype {
	case nstypes.CLONE_NEWUSER:
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
	case nstypes.CLONE_NEWPID:
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

// plainNamespace stores useful information about a concrete Linux kernel
// namespace. It implements the interfaces Namespace, Hierarchy, Ownership,
// and NamespaceStringer. Additionally, it implements the package-private
// interface leaderAdder. (There, I did it. I documented implemented
// interfaces explicitly for clarity.)
type plainNamespace struct {
	nsid      nstypes.NamespaceID
	nstype    nstypes.NamespaceType
	ownernsid nstypes.NamespaceID
	owner     Hierarchy
	ref       string
	leaders   []*Process
}

// namespaceConfigurer allows discovery mechanisms to set up the information
// for a namespace.
type namespaceConfigurer interface {
	AddLeader(proc *Process)             // adds yet another self-styled leader.
	SetRef(string)                       // sets a filesystem path for referencing this namespace.
	DetectOwner(nsf *os.File)            // detects owning user namespace id.
	ResolveOwner(usernsmap NamespaceMap) // resolves owner ns id into object reference.
}

func (pns *plainNamespace) ID() nstypes.NamespaceID     { return pns.nsid }
func (pns *plainNamespace) Type() nstypes.NamespaceType { return pns.nstype }
func (pns *plainNamespace) Owner() Hierarchy            { return pns.owner }
func (pns *plainNamespace) Ref() string                 { return pns.ref }
func (pns *plainNamespace) Leaders() []*Process         { return pns.leaders }

// LeaderPIDs returns the list of leader PIDs.
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
	return fmt.Sprintf("%s:[%d]", pns.nstype.String(), pns.nsid)
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

// DetectOwner gets the ownering user namespace id from Linux, and stores it
// for later resolution only when we have a complete map of all user
// namespaces.
func (pns *plainNamespace) DetectOwner(nsf *os.File) {
	// The User() call gives us an fd wrapped in an os.File, which we can then
	// ask for its namespace ID.
	usernsf, err := rel.User(nsf)
	if err != nil {
		return
	}
	defer usernsf.Close() // Do NOT leak.
	pns.ownernsid, _ = rel.ID(usernsf)
}

// ResolveOwner sets the owning user namespace reference based on the owning
// user namespace id discovered earlier.
func (pns *plainNamespace) ResolveOwner(usernsmap NamespaceMap) {
	// Only try to resolve when we actually got the user namespace id
	// of the owner, otherwise we must skip resolution.
	if pns.ownernsid != 0 {
		ownerns := usernsmap[pns.ownernsid].(*userNamespace)
		pns.owner = ownerns
		ownerns.ownedns[TypeIndex(pns.nstype)][pns.nsid] = pns
	}
}

// hierarchicalNamespace stores hierarchy information in addition to the
// information for plain namespaces. Besides the interfaces for a
// plainNamespace, it additionally implements the public Hierarchy interface.
type hierarchicalNamespace struct {
	plainNamespace
	parent   Hierarchy
	children []Hierarchy
}

// hierarchyConfigurer allows discovery mechanisms to configure the
// information hold by hierarchical namespaces.
type hierarchyConfigurer interface {
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
	child.(hierarchyConfigurer).SetParent(hns)
	hns.children = append(hns.children, child)
}

// SetParent sets the parent namespace of this child namespace. It panics in
// case the parent would change.
func (hns *hierarchicalNamespace) SetParent(parent Hierarchy) {
	if hns.parent != nil && hns.parent != parent {
		panic("trying to change parents might sometimes not a good idea, especially just now.\n" +
			"parent: " + parent.(NamespaceStringer).String() + "\n" +
			"child: " + hns.String())
	}
	hns.parent = parent
}

// userNamespace stores ownership information in addition to the information
// for hierarchical namespaces. On top of the interfaces supported by a
// hierarchicalNamespace, userNamespace implements the Ownership interface.
type userNamespace struct {
	hierarchicalNamespace
	owneruid int
	ownedns  AllNamespaces
}

func (uns *userNamespace) UID() int             { return uns.owneruid }
func (uns *userNamespace) Owned() AllNamespaces { return uns.ownedns }

// String describes this instance of a user namespace, with its parent,
// children, and owned namespaces. This description is non-recursive.
func (uns *userNamespace) String() string {
	u, err := user.LookupId(fmt.Sprintf("%d", uns.owneruid))
	var userstr string
	if err == nil {
		userstr = fmt.Sprintf(" (%q)", u.Username)
	}
	owneds := ""
	var o []string
	for _, ownedbytype := range uns.ownedns {
		for _, owned := range ownedbytype {
			o = append(o, owned.(NamespaceStringer).TypeIDString())
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

// AddChild adds a new child user namespace to this user namespace. Oh, the
// glory of golang!
func (uns *userNamespace) AddChild(child Hierarchy) {
	uns.hierarchicalNamespace.AddChild(child)
}

// SetParent sets the parent user namespace for this user namespace. Oh, the
// glory of golang!
func (uns *userNamespace) SetParent(parent Hierarchy) {
	uns.hierarchicalNamespace.SetParent(parent)
}

// detectUIDs takes an open file referencing a user namespace to query its
// owner's UID and then stores it for this user namespace proxy.
func (uns *userNamespace) detectUID(nsf *os.File) {
	uns.owneruid, _ = rel.OwnerUID(nsf)
}
