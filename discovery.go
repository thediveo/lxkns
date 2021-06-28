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
	"strings"

	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// DiscoveryOpts provides information about the extent of a Linux-kernel
// namespace discovery.
type DiscoverOpts interface {
	// The types of namespaces discovered: this is an OR'ed combination of Linux
	// kernel namespace constants, such as CLONE_NEWNS, CLONE_NEWNET, et cetera.
	// If zero, defaults to discovering all namespaces.
	NamespaceTypes() species.NamespaceType

	ScanProcs() bool            // Scan processes for attached namespaces.
	ScanFds() bool              // Scan open file descriptors for namespaces.
	ScanBindmounts() bool       // Scan bind-mounts for namespaces.
	DiscoverHierarchy() bool    // Discover the hierarchy of PID and user namespaces.
	DiscoverOwnership() bool    // Discover the ownership of non-user namespaces.
	DiscoverFreezerState() bool // Discover the cgroup freezer state of processes.
	DiscoverMounts() bool       // Discover mount point hierarchy with mount paths and visibility.

	Containerizer() containerizer.Containerizer // Discover containers using containerizer.
}

// discoverOpts implements the DiscoveryOpts interface.
type discoverOpts struct {
	namespaceTypes       species.NamespaceType       `json:"-"`                  // types of namespaces to discover (OR'ed).
	scanProcs            bool                        `json:"skipped-procs"`      // scan processes.
	scanFds              bool                        `json:"skipped-fds"`        // scan process file descriptors for references to namespaces.
	scanBindmounts       bool                        `json:"skipped-bindmounts"` // scan for bind-mounted namespaces.
	discoverHierarchy    bool                        `json:"skipped-hierarchy"`  // discover the hierarchy of PID and user namespaces.
	discoverOwnership    bool                        `json:"skipped-ownership"`  // discover the ownership of non-user namespaces.
	discoverFreezerState bool                        `json:"skipped-freezer"`    // discover the cgroup freezer state of processes.
	discoverMounts       bool                        `json:"with-mounts"`        // discover mount paths with mount points.
	containerizer        containerizer.Containerizer `json:"-"`                  // containerizer for discovering containers.
}

func (o *discoverOpts) NamespaceTypes() species.NamespaceType      { return o.namespaceTypes }
func (o *discoverOpts) ScanProcs() bool                            { return o.scanProcs }
func (o *discoverOpts) ScanFds() bool                              { return o.scanFds }
func (o *discoverOpts) ScanBindmounts() bool                       { return o.scanBindmounts }
func (o *discoverOpts) DiscoverHierarchy() bool                    { return o.discoverHierarchy }
func (o *discoverOpts) DiscoverOwnership() bool                    { return o.discoverOwnership }
func (o *discoverOpts) DiscoverFreezerState() bool                 { return o.discoverFreezerState }
func (o *discoverOpts) DiscoverMounts() bool                       { return o.discoverMounts }
func (o *discoverOpts) Containerizer() containerizer.Containerizer { return o.containerizer }

// DiscoveryResult stores the results of a tour through Linux processes and
// kernel namespaces.
type DiscoveryResult struct {
	Options           DiscoverOpts           // options used during discovery.
	Namespaces        model.AllNamespaces    // all discovered namespaces, subject to filtering according to Options.
	InitialNamespaces model.NamespacesSet    // the 7 initial namespaces.
	UserNSRoots       []model.Namespace      // the topmost user namespace(s) in the hierarchy.
	PIDNSRoots        []model.Namespace      // the topmost PID namespace(s) in the hierarchy.
	Processes         model.ProcessTable     // processes checked for namespaces.
	Mounts            NamespacedMountPathMap // per mount-namespace mount paths and mount points.
	Containers        []model.Container      // all alive containers found
}

// discoveryOption ... TODO:
type discoveryOption func(*discoverOpts)

// WithFullDiscovery opts for a full discovery, scanning not only processes, but
// also open file descriptors and bind-mounts, as well as the namespace
// hierarchy and ownership, freezer states, and mount points with their
// visibility.
func WithFullDiscovery() discoveryOption {
	return func(o *discoverOpts) {
		o.namespaceTypes = species.AllNS
		o.scanProcs = true
		o.scanFds = true
		o.scanBindmounts = true
		o.discoverHierarchy = true
		o.discoverOwnership = true
		o.discoverFreezerState = true
		o.discoverMounts = true
	}
}

func WithNamespaceTypes(t species.NamespaceType) discoveryOption {
	return func(o *discoverOpts) { o.namespaceTypes = t }
}

func FromProcs() discoveryOption {
	return func(o *discoverOpts) { o.scanProcs = true }
}

func FromFds() discoveryOption {
	return func(o *discoverOpts) { o.scanFds = true }
}

func FromBindmounts() discoveryOption {
	return func(o *discoverOpts) { o.scanBindmounts = true }
}

func WithHierarchy() discoveryOption {
	return func(o *discoverOpts) { o.discoverHierarchy = true }
}

func WithOwnership() discoveryOption {
	return func(o *discoverOpts) { o.discoverOwnership = true }
}

func WithMounts() discoveryOption {
	return func(o *discoverOpts) { o.discoverMounts = true }
}

// WithContainerizer opts for discovery of containers related to namespaces,
// using the specified Containerizer.
func WithContainerizer(c containerizer.Containerizer) discoveryOption {
	return func(o *discoverOpts) {
		o.containerizer = c
	}
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
func (dr *DiscoveryResult) SortedNamespaces(nsidx model.NamespaceTypeIndex) []model.Namespace {
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

// Discover returns the Linux kernel namespaces found, based on discovery
// options specified in the call. The discovery results also specify the
// initial namespaces, as well the process table/tree on which the discovery
// bases at least in part.
func Discover(options ...discoveryOption) *DiscoveryResult {
	opts := &discoverOpts{}
	for _, opt := range options {
		opt(opts)
	}
	// If no namespace types are specified for discovery, we take this as
	// discovering all types of namespaces.
	if opts.namespaceTypes == 0 {
		opts.namespaceTypes = species.AllNS
	}
	result := &DiscoveryResult{
		Options:   opts,
		Processes: model.NewProcessTable(opts.discoverFreezerState),
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
			disco.Discover(opts.namespaceTypes, "/proc", result)
		} else {
			for _, nstypeidx := range *disco.When {
				if nstype := model.TypesByIndex[nstypeidx]; opts.namespaceTypes&nstype != 0 {
					disco.Discover(nstype, "/proc", result)
				}
			}
		}
	}
	// Fill in some additional convenience fields in the result.
	if opts.namespaceTypes&species.CLONE_NEWUSER != 0 {
		result.UserNSRoots = rootNamespaces(result.Namespaces[model.UserNS])
	}
	if opts.namespaceTypes&species.CLONE_NEWPID != 0 {
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

	// Optionally discover alive containers and relate the.
	discoverContainers(result)

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
