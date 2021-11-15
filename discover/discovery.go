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

//go:build linux
// +build linux

package discover

import (
	"fmt"
	"sort"
	"strings"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// Result stores the results of a tour through Linux processes and
// kernel namespaces.
type Result struct {
	Options           DiscoverOpts           // options used during discovery.
	Namespaces        model.AllNamespaces    // all discovered namespaces, subject to filtering according to Options.
	InitialNamespaces model.NamespacesSet    // the 7 initial namespaces.
	UserNSRoots       []model.Namespace      // the topmost user namespace(s) in the hierarchy.
	PIDNSRoots        []model.Namespace      // the topmost PID namespace(s) in the hierarchy.
	Processes         model.ProcessTable     // processes checked for namespaces.
	PIDMap            model.PIDMapper        `json:"-"` // optional PID translator.
	Mounts            NamespacedMountPathMap // per mount-namespace mount paths and mount points.
	Containers        model.Containers       // all alive containers found
}

// SortNamespaces returns a sorted copy of a list of namespaces. The
// namespaces are sorted by their namespace ids in ascending order.
func SortNamespaces(nslist []model.Namespace) []model.Namespace {
	newnslist := make([]model.Namespace, len(nslist))
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
func SortChildNamespaces(nslist []model.Hierarchy) []model.Hierarchy {
	newnslist := make([]model.Hierarchy, len(nslist))
	copy(newnslist, nslist)
	sort.Slice(newnslist, func(i, j int) bool {
		return newnslist[i].(model.Namespace).ID().Ino < newnslist[j].(model.Namespace).ID().Ino
	})
	return newnslist
}

// SortedNamespaces returns the namespaces from a map sorted.
func SortedNamespaces(nsmap model.NamespaceMap) []model.Namespace {
	// Copy the namespaces from the map into a slice, so we can then sort it
	// next.
	nslist := make([]model.Namespace, len(nsmap))
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
// inode number (on the special "nsfs" filesystem), ignoring a namespace's
// device ID.
func (dr *Result) SortedNamespaces(nsidx model.NamespaceTypeIndex) []model.Namespace {
	return SortedNamespaces(dr.Namespaces[nsidx])
}

// rootNamespaces returns the topmost namespace(s) in a hierarchy of
// namespaces. This function can be used only on hierarchical namespaces and
// will panic if misused.
func rootNamespaces(nsmap model.NamespaceMap) []model.Namespace {
	result := []model.Namespace{}
	for _, ns := range nsmap {
		hns, ok := ns.(model.Hierarchy)
		if !ok {
			panic(fmt.Sprintf(
				"rootNamespaces: found invalid non-hierarchical namespace %s",
				ns.(model.NamespaceStringer).TypeIDString()))
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
var discoverySequence = []model.NamespaceTypeIndex{
	model.UserNS,
	model.PIDNS,
}

// Completes discoveryOrder to finally contain all namespace types indices;
// the exact ordering of the remaining namespace types doesn't matter. Only
// that user namespaces come first, then PID namespaces, then all other
// namespace types.
func init() {
	for typeidx := model.NamespaceTypeIndex(0); typeidx < model.NamespaceTypesCount; typeidx++ {
		if typeidx != model.UserNS && typeidx != model.PIDNS {
			discoverySequence = append(discoverySequence, typeidx)
		}
	}
}

// Namespaces returns the Linux kernel namespaces found, based on discovery
// options specified in the call. The discovery results also specify the initial
// namespaces, as well the process table/tree on which the discovery bases at
// least in part.
func Namespaces(options ...DiscoveryOption) *Result {
	opts := DiscoverOpts{}
	for _, opt := range options {
		opt(&opts)
	}
	// If no namespace types are specified for discovery, we take this as
	// discovering all types of namespaces.
	if opts.NamespaceTypes == 0 {
		opts.NamespaceTypes = species.AllNS
	}
	result := &Result{
		Options:   opts,
		Processes: model.NewProcessTable(opts.DiscoverFreezerState),
	}
	// Finish initialization.
	for idx := range result.Namespaces {
		result.Namespaces[idx] = model.NamespaceMap{}
	}
	// Now go for discovery: we run the available discovery functions in
	// sequence, subject to the following rules for the When field:
	//   - []: call discovery function once; it'll know what to do.
	//   - [...]: call discovery function multiple times, once for each
	//     namespace type listed in the When field, and in the same order of
	//     sequence.
	for _, disco := range discoverers {
		if len(*disco.When) == 0 {
			disco.Discover(opts.NamespaceTypes, "/proc", result)
		} else {
			for _, nstypeidx := range *disco.When {
				if nstype := model.TypesByIndex[nstypeidx]; opts.NamespaceTypes&nstype != 0 {
					disco.Discover(nstype, "/proc", result)
				}
			}
		}
	}
	// Fill in some additional convenience fields in the result.
	if opts.NamespaceTypes&species.CLONE_NEWUSER != 0 {
		result.UserNSRoots = rootNamespaces(result.Namespaces[model.UserNS])
	}
	if opts.NamespaceTypes&species.CLONE_NEWPID != 0 {
		result.PIDNSRoots = rootNamespaces(result.Namespaces[model.PIDNS])
	}
	// TODO: Find the initial namespaces...

	log.Infofn(func() string {
		perns := []string{}
		for nstypeidx, nsmap := range result.Namespaces {
			perns = append(perns, fmt.Sprintf("%d %s", len(nsmap), model.TypesByIndex[nstypeidx].Name()))
		}
		return fmt.Sprintf("discovered %s namespaces", strings.Join(perns, ", "))
	})

	// Do we need a PID mapping between PID namespaces?
	if opts.withPIDmap {
		result.PIDMap = NewPIDMap(result)
	}

	// Optionally discover alive containers and relate the.
	discoverContainers(result)

	// As a C oldie it gives me the shivers to return a pointer to what might
	// look like an "auto" local struct ;)
	return result
}

// discoveryFunc implements some Linux kernel namespace discovery
// functionality.
type discoveryFunc func(species.NamespaceType, string, *Result)

// discoverer describes a single discoveryFunc and when to call it: once, per
// each namespace type and for which namespace types in what sequence. Please
// note that we use a reference to a slice here, as discoverySequence will
// only be only completed during the init() phase, but after(!)
// discoverySequence has been set to its initial value. Sigh.
type discoverer struct {
	When     *[]model.NamespaceTypeIndex // indices of namespace types this discovery function discovers.
	Discover discoveryFunc               // the concrete namespace discovery functionality.
}

// Run a discoveryFunc only once per discovery, because it needs to work on
// multiple namespace types in a single discovery call, and doesn't like
// multiple per-type calls.
var discoveronce = []model.NamespaceTypeIndex{}

// The sequence of discovery functions implemented in lxkns, and how to call
// them. Because discoverySequence will only be completed after this list has
// been initialized, we need to "late bind" it by reference (pointers to
// slices, where has the world come to ... mumble ... mumble...)
var discoverers = []discoverer{
	{&discoverySequence, discoverFromProc},
	{&discoveronce, discoverFromFd},
	{&discoveronce, discoverBindmounts},
	{&[]model.NamespaceTypeIndex{model.UserNS, model.PIDNS}, discoverHierarchy},
	{&discoverySequence, resolveOwnership},
	{&discoveronce, discoverFromMountinfo},
}
