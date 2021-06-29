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
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/species"
)

// DiscoveryOpts provides information about the extent of a Linux-kernel
// namespace discovery.
//
// This information is JSON-marshallable, with the exception of the
// Containerizer interface.
type DiscoverOpts struct {
	// The types of namespaces discovered: this is an OR'ed combination of Linux
	// kernel namespace constants, such as CLONE_NEWNS, CLONE_NEWNET, et cetera.
	// If zero, defaults to discovering all namespaces.
	NamespaceTypes species.NamespaceType `json:"-"`

	ScanProcs            bool `json:"from-procs"`      // Scan processes for attached namespaces.
	ScanFds              bool `json:"from-fds"`        // Scan open file descriptors for namespaces.
	ScanBindmounts       bool `json:"from-bindmounts"` // Scan bind-mounts for namespaces.
	DiscoverHierarchy    bool `json:"with-hierarchy"`  // Discover the hierarchy of PID and user namespaces.
	DiscoverOwnership    bool `json:"with-ownership"`  // Discover the ownership of non-user namespaces.
	DiscoverFreezerState bool `json:"with-freezer"`    // Discover the cgroup freezer state of processes.
	DiscoverMounts       bool `json:"with-mounts"`     // Discover mount point hierarchy with mount paths and visibility.

	Containerizer containerizer.Containerizer `json:"-"` // Discover containers using containerizer.
}

// DiscoveryOption represents a function able to set a particular discovery
// option state in DiscoverOpts.
type DiscoveryOption func(*DiscoverOpts)

// WithStandardDiscovery opts for a "standard" discovery, scanning not only
// processes, but also open file descriptors and bind-mounts, as well as the
// namespace hierarchy and ownership, and freezer states. All types of
// namespaces will be discovered. Please note that time namespaces can only be
// discovered on newer kernels with support for them.
//
// Please note that mount point discovery (including visibility calculation) it
// not opted in; it has to be opted in individually.
func WithStandardDiscovery() DiscoveryOption {
	return func(o *DiscoverOpts) {
		o.NamespaceTypes = species.AllNS
		o.ScanProcs = true
		o.ScanFds = true
		o.ScanBindmounts = true
		o.DiscoverHierarchy = true
		o.DiscoverOwnership = true
		o.DiscoverFreezerState = true
	}
}

var stddisco = WithStandardDiscovery()

// WithFullDiscovery opts in to all discovery features that lxkns has to offer.
// Please note that API users still need to set an optional Containerizer
// explicitly.
func WithFullDiscovery() DiscoveryOption {
	return func(o *DiscoverOpts) {
		stddisco(o)
		o.DiscoverMounts = true
	}
}

// WithNamespaceTypes sets the types of namespaces to discover, where multiple
// types need to be OR'ed together. Setting 0 will discover all available types.
func WithNamespaceTypes(t species.NamespaceType) DiscoveryOption {
	return func(o *DiscoverOpts) { o.NamespaceTypes = t }
}

// FromProcs opts to find namespaces attached to processes.
func FromProcs() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanProcs = true }
}

// NotFromProcs opts out of looking at processes when searching for namespaces.
func NotFromProcs() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanProcs = false }
}

// FromFds opts to find namespaces from the open file descriptors of processes.
func FromFds() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanFds = true }
}

// NotFromFds opts out looking at the open file descriptors of processes when
// searching for namespaces.
func NotFromFds() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanFds = false }
}

// FromBindmounts opts to find bind-mounted namespaces.
func FromBindmounts() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanBindmounts = true }
}

// NotFromBindmounts opts out from searching for bind-mounted namespaces.
func NotFromBindmounts() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanBindmounts = false }
}

// WithHierarchy opts to query the namespace hierarchy of PID and user
// namespaces.
func WithHierarchy() DiscoveryOption {
	return func(o *DiscoverOpts) { o.DiscoverHierarchy = true }
}

// WithoutHierarchy opts out of querying the namespace hierarchy of PID and user
// namespaces.
func WithoutHierarchy() DiscoveryOption {
	return func(o *DiscoverOpts) { o.DiscoverHierarchy = false }
}

// WithOwnership opts to find the ownership relations between user namespaces
// and all other namespaces.
func WithOwnership() DiscoveryOption {
	return func(o *DiscoverOpts) { o.DiscoverOwnership = true }
}

// WithoutOwnership opts out of looking for the ownership relations between user
// namespaces and all other namespaces.
func WithoutOwnership() DiscoveryOption {
	return func(o *DiscoverOpts) { o.DiscoverOwnership = false }
}

// WithMounts opts to find mount points and determine their visibility.
func WithMounts() DiscoveryOption {
	return func(o *DiscoverOpts) { o.DiscoverMounts = true }
}

// WithoutMounts opts out of finding mount points and determining their
// visibility.
func WithoutMounts() DiscoveryOption {
	return func(o *DiscoverOpts) { o.DiscoverMounts = false }
}

// WithContainerizer opts for discovery of containers related to namespaces,
// using the specified Containerizer.
func WithContainerizer(c containerizer.Containerizer) DiscoveryOption {
	return func(o *DiscoverOpts) {
		o.Containerizer = c
	}
}

// SameAs reuses the discovery options used for a previous discovery.
func SameAs(r *DiscoveryResult) DiscoveryOption {
	return func(o *DiscoverOpts) {
		*o = r.Options
	}
}
