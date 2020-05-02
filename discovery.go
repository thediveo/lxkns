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
	"sort"

	"github.com/thediveo/lxkns/species"
)

// DiscoverOpts gives control over the extend and thus time and resources
// spent on discovering Linux kernel namespaces, their relationships between
// them, and with processes.
type DiscoverOpts struct {
	// The types of namespaces to discover: this is an OR'ed combination of
	// Linux kernel namespace constants, such as CLONE_NEWNS, CLONE_NEWNET, et
	// cetera. If zero, defaults to discovering all namespaces.
	NamespaceTypes species.NamespaceType

	// Where to scan (or not scan) for signs of namespaces?
	SkipProcs      bool // Don't scan processes.
	SkipTasks      bool // Don't scan threads, a.k.a. tasks.
	SkipFds        bool // Don't scan process file descriptors for references to namespaces.
	SkipBindmounts bool // Don't scan for bind-mounted namespaces.
	SkipHierarchy  bool // Don't discover the hierarchy of PID and user namespaces.
	SkipOwnership  bool // Don't discover the ownership of non-user namespaces.
}

// FullDiscovery sets the discovery options to a full and thus extensive
// discovery process.
var FullDiscovery = DiscoverOpts{}

// NoDiscovery set the discovery options to not discover anything. This option
// set can be used to start from when only a few chosen discovery methods are
// to be enabled.
var NoDiscovery = DiscoverOpts{
	SkipProcs:      true,
	SkipTasks:      true,
	SkipFds:        true,
	SkipBindmounts: true,
	SkipHierarchy:  true,
	SkipOwnership:  true,
}

// DiscoveryResult stores the results of a tour through Linux processes and
// kernel namespaces.
type DiscoveryResult struct {
	Options           DiscoverOpts  // options used during discovery.
	Namespaces        AllNamespaces // all discovered namespaces, subject to filtering according to Options.
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
		return newnslist[i].ID().Ino < newnslist[j].ID().Ino
	})
	return newnslist
}

// SortChildNamespaces returns a sorted copy of a list of hierarchical
// namespaces. The namespaces are sorted by their namespace ids in ascending
// order. Please note that the list itself is flat, but this function can only
// be used on hierarchical namespaces (PID, user).
func SortChildNamespaces(nslist []Hierarchy) []Hierarchy {
	newnslist := make([]Hierarchy, len(nslist))
	copy(newnslist, nslist)
	sort.Slice(newnslist, func(i, j int) bool {
		return newnslist[i].(Namespace).ID().Ino < newnslist[j].(Namespace).ID().Ino
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
		return nslist[i].ID().Ino < nslist[j].ID().Ino
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
		result.Options.NamespaceTypes = species.CLONE_NEWNS |
			species.CLONE_NEWCGROUP | species.CLONE_NEWUTS |
			species.CLONE_NEWIPC | species.CLONE_NEWUSER |
			species.CLONE_NEWPID | species.CLONE_NEWNET
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
			disco.Discover(result.Options.NamespaceTypes, "/proc", result)
		} else {
			for _, nstypeidx := range *disco.When {
				if nstype := TypesByIndex[nstypeidx]; result.Options.NamespaceTypes&nstype != 0 {
					disco.Discover(nstype, "/proc", result)
				}
			}
		}
	}
	// Fill in some additional convenience fields in the result.
	if result.Options.NamespaceTypes&species.CLONE_NEWUSER != 0 {
		result.UserNSRoots = rootNamespaces(result.Namespaces[UserNS])
	}
	if result.Options.NamespaceTypes&species.CLONE_NEWPID != 0 {
		result.PIDNSRoots = rootNamespaces(result.Namespaces[PIDNS])
	}
	// TODO: Find the initial namespaces...

	// As a C oldie it gives me the shivers to return a pointer to what might
	// look like an "auto" local struct ;)
	return result
}

// discoveryFunc implements some Linux kernel namespace discovery
// functionality.
type discoveryFunc func(species.NamespaceType, string, *DiscoveryResult)

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
// them. Because discoverySequence will only be completed after this list has
// been initialized, we need to "late bind" it by reference (pointers to
// slices, where has the world come to ... mumble ... mumble...)
var discoverers = []discoverer{
	{&discoverySequence, discoverFromProc},
	{&discoveronce, discoverFromFd},
	{&discoveronce, discoverBindmounts},
	{&[]NamespaceTypeIndex{UserNS, PIDNS}, discoverHierarchy},
	{&discoverySequence, resolveOwnership},
}
