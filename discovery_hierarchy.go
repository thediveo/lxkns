// Discovers the hierarchy of user and PID namespaces using Linux kernel
// syscalls. This might turn up namespaces we haven't found so far and which
// are somewhere hidden within the hierarchy without any /proc, fd, or bind
// mounted references. Most importantly, only we get the correct hierarchical
// information in this discovery phase.
//
// This discovery must be run late, so that any bind-mounted or fd'ed user and
// pid leaf namespaces were found by now, because we cannot find them here,
// due to the way the Linux kernel exposes the user/PID namespaces hierarchy:
// only given a child namespace, we can then query its parent namespace. But
// we cannot query a parent namespace for all its children.
//
// See also: http://man7.org/linux/man-pages/man2/ioctl_ns.2.html.

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
	"os"

	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
)

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
func discoverHierarchy(nstype species.NamespaceType, _ string, result *DiscoveryResult) {
	if result.Options.SkipHierarchy {
		return
	}
	nstypeidx := TypeIndex(nstype)
	nsmap := result.Namespaces[nstypeidx]
	for _, somens := range nsmap {
		ns := somens // ...so we can later climb rung by rung.
		if ns.(Hierarchy).Parent() != nil {
			// Early exit: skip this user/PID namespace, if it has already
			// been brought into the hierarchy as part of the
			// line-of-hierarchy for another user/PID namespace.
			continue
		}
		// For climbing up the hierarchy, Linux wants us to give it file
		// descriptors referencing the namespaces to be quieried for their
		// parents.
		nsf, err := ops.NewNamespaceFile(os.OpenFile(ns.Ref(), os.O_RDONLY, 0))
		if err != nil {
			continue
		}
		// Now, go climbing up the hierarchy...
		for {
			// We already worked on this user/pid namespace, so we don't need
			// to climb up further. This won't catch the initial user/pid
			// namespaces, but then these will break out of the loop anyway,
			// as they don't have any parents.
			if ns.(Hierarchy).Parent() != nil {
				break
			}
			// By the way ... if it's a user namespace, then get its owner's
			// UID, as we just happen to have a useful fd referencing the
			// namespace open anyway.
			if nstype == species.CLONE_NEWUSER {
				ns.(*userNamespace).detectUID(nsf)
			}
			// See if there is a parent of this namespace at all, or whether
			// we've reached the end of the road. Normally, this should be the
			// initial user or PID namespace. But if we have insufficient
			// capabilities, then we'll hit a brickwall earlier.
			parentnsf, err := nsf.Parent()
			if err != nil {
				// There is no parent user/PID namespace, so we're done in
				// this line. Let's move on to the next namespace. The reasons
				// for not having a parent are: (1) initial namespace, so no
				// parent; (2) no capabilities in parent namespace, so no
				// parent either.
				break
			}
			parentnsid, err := parentnsf.ID()
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
			parentns.(HierarchyConfigurer).AddChild(ns.(Hierarchy))
			ns = parentns
			nsf.Close()
			nsf = parentnsf
		}
		// Don't leak...
		nsf.Close()
	}
}
