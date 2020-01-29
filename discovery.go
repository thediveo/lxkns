// Defines the core procedures for discovering Linux kernel namespaces in
// different places of a running Linux system.

// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build linux

package lxkns

import (
	"fmt"
	"github.com/thediveo/lxkns/nstypes"
	rel "github.com/thediveo/lxkns/relations"
	"os"
	"sort"
)

// DiscoverOpts gives control over the extend and thus time and resources
// spent on discovering Linux kernel namespaces, their relationships between
// them, and with processes.
type DiscoverOpts struct {
	// The types of namespaces to discover: this is an OR'ed combination of
	// Linux kernel namespace constants, such as CLONE_NEWNS, CLONE_NEWNET, et
	// cetera. If zero, defaults to discovering all namespaces.
	NamespaceTypes nstypes.NamespaceType

	// Where to scan (or not scan) for signs of namespaces?
	SkipBindmounts bool // Don't scan for bind-mounted namespaces.
	SkipFds        bool // Don't scan process file descriptors for references to namespaces.
	SkipHierarchy  bool // Don't discover the hierarchy of PID and user namespaces.
	SkipOwnership  bool // Don't discover the ownership of non-user namespaces.
}

// FullDiscovery sets the discovery options to a full and thus extensive
// discovery process.
var FullDiscovery = DiscoverOpts{}

// DiscoveryResult stores the results of a tour through Linux processes and
// kernel namespaces.
type DiscoveryResult struct {
	Options           DiscoverOpts  // options used during discovery.
	Namespaces        AllNamespaces // all discovered namespaces, subjectg to filtering according to Options.
	InitialNamespaces NamespacesSet // the 7 initial namespaces.
	UserNSRoots       []Namespace   // the topmost user namespace(s) in the hierarchy
	PIDNSRoots        []Namespace   // the topmost PID namespace(s) in the hierarchy
	Processes         ProcessTable  // processes checked for namespaces.
}

// SortNamespaces returns a sorted copy of a list of namespaces. The
// namespaces are sorted by their namespace ids in ascending order.
func SortNamespaces(nslist []Namespace) []Namespace {
	newnslist := make([]Namespace, len(nslist))
	copy(newnslist, nslist)
	sort.Slice(newnslist, func(i, j int) bool {
		return newnslist[i].ID() < newnslist[j].ID()
	})
	return newnslist
}

// SortChildNamespaces returns a sorted copy of a list of hierarchical
// namespaces. The namespaces are sorted by their namespace ids in ascending
// order.
func SortChildNamespaces(nslist []Hierarchy) []Hierarchy {
	newnslist := make([]Hierarchy, len(nslist))
	copy(newnslist, nslist)
	sort.Slice(newnslist, func(i, j int) bool {
		return newnslist[i].(Namespace).ID() < newnslist[j].(Namespace).ID()
	})
	return newnslist
}

// SortedNamespaces returns the namespaces from a map sorted.
func SortedNamespaces(nsmap NamespaceMap) []Namespace {
	// Copy the namespaces from the map into a slice, so we can then sort it
	// next.
	nslist := make([]Namespace, len(nsmap))
	idx := 0
	for _, ns := range nsmap {
		nslist[idx] = ns
		idx++
	}
	sort.Slice(nslist, func(i, j int) bool {
		return nslist[i].ID() < nslist[j].ID()
	})
	return nslist
}

// SortedNamespaces returns a sorted list of discovered namespaces of the
// specified type. The namespaces are sorted by their identifier, which is an
// inode number (on the special "nsfs" filesystem).
func (dr *DiscoveryResult) SortedNamespaces(nsidx NamespaceTypeIndex) []Namespace {
	return SortedNamespaces(dr.Namespaces[nsidx])
}

// rootNamespaces returns the topmost namespace(s) in a hierarchy of
// namespaces. This function can be used only on hierarchical namespaces and
// will panic if misused.
func rootNamespaces(nsmap NamespaceMap) []Namespace {
	result := []Namespace{}
	for _, ns := range nsmap {
		hns, ok := ns.(Hierarchy)
		if !ok {
			panic(fmt.Sprintf(
				"rootNamespaces: found invalid non-hierarchical namespace %s",
				ns.(NamespaceStringer).TypeIDString()))
		}
		if hns.Parent() == nil {
			result = append(result, ns)
		}
	}
	return result
}

// discoverySequence contains the namespace type indices in the order of
// preferred discovery. While often the order of sequence doesn't matter,
// there are few cases where it makes coding discovery functionality easier
// when there is a guaranteed type order in place.
var discoverySequence = []NamespaceTypeIndex{
	UserNS,
	PIDNS,
}

// Completes discoveryOrder to finally contain all namespace types indices.
func init() {
	for _, typeidx := range typeIndices {
		if typeidx != UserNS && typeidx != PIDNS {
			discoverySequence = append(discoverySequence, typeidx)
		}
	}
}

// Discover returns the Linux kernel namespaces found, based on discovery
// options specified in the call. The discovery results also specify the
// initial namespaces, as well the process table/tree on which the discovery
// bases at least in part.
func Discover(opts DiscoverOpts) *DiscoveryResult {
	result := &DiscoveryResult{
		Options:   opts,
		Processes: NewProcessTable(),
	}
	// If no namespace types are specified for discovery, we take this as
	// discovering all types of namespaces.
	if result.Options.NamespaceTypes == 0 {
		result.Options.NamespaceTypes = nstypes.CLONE_NEWNS |
			nstypes.CLONE_NEWCGROUP | nstypes.CLONE_NEWUTS |
			nstypes.CLONE_NEWIPC | nstypes.CLONE_NEWUSER |
			nstypes.CLONE_NEWPID | nstypes.CLONE_NEWNET
	}
	// Finish initialization.
	for idx := range result.Namespaces {
		result.Namespaces[idx] = NamespaceMap{}
	}
	// Now go for discovery: we run the available discovery functions in
	// sequence, subject to the following rules for the When field:
	//   - []: call discovery function once; it'll know what to do.
	//   - [...]: call discovery function multiple times, once for each
	//     namespace type listed in the When field, and in the same order of
	//     sequence.
	for _, disco := range discoverers {
		if len(*disco.When) == 0 {
			disco.Discover(^nstypes.NamespaceType(0), result)
		} else {
			for _, nstypeidx := range *disco.When {
				if nstype := TypesByIndex[nstypeidx]; result.Options.NamespaceTypes&nstype != 0 {
					disco.Discover(nstype, result)
				}
			}
		}
	}
	// Fill in some additional convenience fields in the result.
	if result.Options.NamespaceTypes&nstypes.CLONE_NEWUSER != 0 {
		result.UserNSRoots = rootNamespaces(result.Namespaces[UserNS])
	}
	if result.Options.NamespaceTypes&nstypes.CLONE_NEWPID != 0 {
		result.PIDNSRoots = rootNamespaces(result.Namespaces[PIDNS])
	}
	// As a C oldie it gives me the shivers to return a pointer to what might
	// look like an "auto" local struct ;)
	return result
}

// discoveryFunc implements some Linux kernel namespace discovery
// functionality.
type discoveryFunc func(nstypes.NamespaceType, *DiscoveryResult)

// discoverer describes a single discoveryFunc and when to call it: once, per
// each namespace type and for which namespace types in what sequence. Please
// note that we use a reference to a slice here, as discoverySequence will
// only be only completed during the init() phase, but after(!)
// discoverySequence has been set to its initial value. Sigh.
type discoverer struct {
	When     *[]NamespaceTypeIndex // indices of namespace types this discovery function discovers.
	Discover discoveryFunc         // the concrete namespace discovery functionality.
}

// Run a discoveryFunc only once per discovery, because it needs to work on
// multiple namespace types in a single discovery call, and doesn't like
// multiple per-type calls.
var discoveronce = []NamespaceTypeIndex{}

// The sequence of discovery functions implemented in lxkns, and how to call
// them.
var discoverers = []discoverer{
	{&discoverySequence, discoverFromProc},
	{&[]NamespaceTypeIndex{UserNS, PIDNS}, discoverHierarchy},
	{&discoveronce, resolveOwnership},
}

// discoverFromProc discovers Linux kernel namespaces from the process table,
// using the namespace links inside the proc filesystem: "/proc/[PID]/ns/...".
// It does not check any other places, as these are covered by separate
// discovery functions.
func discoverFromProc(nstype nstypes.NamespaceType, result *DiscoveryResult) {
	nstypename := nstypes.TypeName(nstype)
	nstypeidx := TypeIndex(nstype)
	nsmap := result.Namespaces[nstypeidx]
	// For all processes (but not tasks/threads) listed in /proc try to gather
	// the namespaces of a given type they use.
	for pid, proc := range result.Processes {
		// Discover the namespace instance of the specified type which this
		// particular process has joined. Please note that namespace
		// references for processes appear as symbolic(!) links in the /proc
		// filesystem, but in fact are behaving like hard links. Nevertheless,
		// we have to follow them like symbolic links in order to find the
		// identifier in form of the inode # of the referenced namespace.
		nsref := fmt.Sprintf("/proc/%d/ns/%s", pid, nstypename)
		// Avoid using high-level golang i/o calls, as these like to hand over
		// to yet another goroutine, something which really doesn't help us
		// here.
		nsf, err := os.OpenFile(nsref, os.O_RDONLY, 0)
		if err != nil {
			continue
		}
		nsid, err := rel.ID(nsf)
		if err != nil {
			nsf.Close() // ...don't leak!
			continue
		}
		ns, ok := nsmap[nsid]
		if !ok {
			// Only add a namespace we haven't yet seen. And yes, we don't
			// give a reference here, as we want to use a reference from a
			// leader process, instead of some child process deep down the
			// hierarchy, which might not even live for long (as sad as this
			// might be).
			ns = NewNamespace(nstype, nsid, "")
			nsmap[nsid] = ns
		}
		// To speed up finding the process leaders in a specific namespace, we
		// remember this namespace as joined by the process we're just looking
		// at.
		proc.Namespaces[nstypeidx] = ns
		// Let's also get the owning user namespace id, while we still have a
		// suitable fd open. For user namespaces, we skip this step, as this
		// is the same as the parent relationship. Additionally, it makes
		// things too awkward in the model, because then we would need to
		// treat ownership differently for non-user namespaces versus user
		// namespaces all the time. Thus, sorry, no user namespaces here.
		if !result.Options.SkipOwnership && nstype != nstypes.CLONE_NEWUSER {
			ns.(namespaceConfigurer).DetectOwner(nsf)
		}
		// Don't leak... And no, defer won't help us here.
		nsf.Close()
	}
	// Now that we know which namespaces are existing with processes joined to
	// them, let's find out the leader processes in these namespaces...
	for pid, proc := range result.Processes {
		// In case we got no access to this process, we must skip it. And we
		// must remove it from our process table, so others won't try to use
		// them. This will not remove the process from the process tree, rest
		// assured.
		if proc.Namespaces[nstypeidx] == nil {
			delete(result.Processes, pid) // FIXME: really?
			continue
		}
		// Find leader from this position in the process tree: a "leader" is
		// the topmost process in the process tree which is still joined to
		// the same namespace as the namespace of the process from which we
		// started our quest.
		p := proc
		parentp := p.Parent
		for parentp != nil && parentp.Namespaces[nstypeidx] == p.Namespaces[nstypeidx] {
			p = parentp
			parentp = p.Parent
		}
		p.Namespaces[nstypeidx].(namespaceConfigurer).AddLeader(p)
	}
	// Try to set namespace references which we hope to be as longlived as
	// possible; so we use one of the leader processes.
	for _, ns := range nsmap {
		if leaders := ns.Leaders(); len(leaders) > 0 {
			ns.(namespaceConfigurer).SetRef(
				fmt.Sprintf("/proc/%d/ns/%s", leaders[0].PID, nstypename))
		}
	}
}

// discoverHierarchy unmasks the hierarchy of user and PID namespaces. All
// other types of Linux kernel namespaces don't form hierarchies within their
// type. (This simplifies ownership relations to not be hierarchical, as your
// cats surely will testify to with greatest pleasure.)
//
// For user namespaces, this also discovers the owner's UID; the rationale is
// that this is the most efficient way to do it, otherwise we would need to
// retraverse the hierarchy for all user namespaces again during discovering
// the overall ownership relations. The problem with a later discovery is that
// hidden namespaces don't have file paths as references but instead can only
// be referenced by fd's returned by the kernel namespace ioctl()s. This would
// then force us to keep potentially a larger number of fd's open.
func discoverHierarchy(nstype nstypes.NamespaceType, result *DiscoveryResult) {
	if result.Options.SkipHierarchy {
		return
	}
	nstypeidx := TypeIndex(nstype)
	nsmap := result.Namespaces[nstypeidx]
	for _, somens := range nsmap {
		ns := somens // ...so we can later climb rung by rung.
		if ns.(Hierarchy).Parent() != nil {
			// Skip this user/PID namespace, if it has already been brought
			// into the hierarchy as part of the line-of-hierarchy for another
			// user/PID namespace.
			continue
		}
		// For climbing up the hierarchy, Linux wants us to give it file
		// descriptors referencing the namespaces to be quieried for their
		// parents.
		nsf, err := os.OpenFile(ns.Ref(), os.O_RDONLY, 0)
		if err != nil {
			continue
		}
		// Now, go climbing up the hierarchy...
		for {
			// By the way ... if it's a user namespace, then get its owner's
			// UID, as we just happen to have a useful fd referencing the
			// namespace open anyway.
			if nstype == nstypes.CLONE_NEWUSER {
				ns.(*userNamespace).detectUID(nsf)
			}
			// See if there is a parent of this namespace at all, or whether
			// we've reached the end of the road. Normally, this should be the
			// initial user or PID namespace. But if we have insufficient
			// capabilities, then we'll hit a brickwall earlier.
			parentnsf, err := rel.Parent(nsf)
			if err != nil {
				// There is no parent user/PID namespace, so we're done in
				// this line. Let's move on to the next namespace. The reasons
				// for not having a parent are: (1) initial namespace, so no
				// parent; (2) no capabilities in parent namespace, so no
				// parent either.
				break
			}
			parentnsid, err := rel.ID(parentnsf)
			if err != nil {
				// There is something severely rotten here, because the kernel
				// just gave us a parent namespace reference which we cannot
				// stat. Either we get a parent namespace reference which then
				// has to work, or we won't get a reference from the parent
				// namespace ioctl() syscall.
				panic("cannot stat parent namespace fd reference")
			}
			parentns, ok := nsmap[parentnsid]
			if !ok {
				// So we've found a "hidden" namespace. For user namespaces
				// this happens when there are no processes joined to a
				// particular user namespace, but this user namespace has
				// still child user namespaces. For PID namespaces this can
				// only happen when bind-mounting a PID namespace or keeping
				// it opened by an file descriptor ("fd-tied"), and there are
				// no processes either in it or any of its child processes
				// (which are also bind-mounted or fd-tied).
				//
				// Anyway, we need to create a new namespace node for what we
				// found.
				parentns = NewNamespace(nstype, parentnsid, "")
				nsmap[parentnsid] = parentns
			}
			// Now insert the current namespace as a child of its parent in
			// the hierarchy, and then prepare for the next rung...
			parentns.(hierarchyConfigurer).AddChild(ns.(Hierarchy))
			ns = parentns
			nsf.Close()
			nsf = parentnsf
		}
		// Don't leak...
		nsf.Close()
	}
}

// resolveOwnership unearths which non-user namespaces are owned by which user
// namespaces. We only run the resolution phase after we've discovered a
// complete map of all user namespaces: only now we can resolve the owner
// userspace ids to their corresponding user namespace objects.
func resolveOwnership(nstype nstypes.NamespaceType, result *DiscoveryResult) {
	if !result.Options.SkipOwnership && nstype != nstypes.CLONE_NEWUSER {
		// The namespace type discovery sequence guarantees us that by the
		// time we got here, the user namespaces already have been fully
		// discovered, so we have a complete map of them.
		usernsmap := result.Namespaces[UserNS]
		nstypeidx := TypeIndex(nstype)
		nsmap := result.Namespaces[nstypeidx]
		for _, ns := range nsmap {
			ns.(namespaceConfigurer).ResolveOwner(usernsmap)
		}
	}
}
