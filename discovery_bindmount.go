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
	"encoding/json"
	"os"

	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns/nstypes"
	rel "github.com/thediveo/lxkns/relations"
)

// OwnedMountInfo augments bind-mount information with the owning user
// namespace ID, so we can later correctly set up the ownership relations in
// the discovery results.
type OwnedMountInfo struct {
	*mntinfo.Mountinfo
	OwnernsID nstypes.NamespaceID `json:"ownernsid"`
}

// discoverBindmounts checks bind-mounts to discover namespaces we haven't
// found so far in the process' joined namespaces. This discovery function is
// designed to be run only once per discovery.
func discoverBindmounts(_ nstypes.NamespaceType, _ string, result *DiscoveryResult) {
	if result.Options.SkipBindmounts {
		return
	}
	// Helper function which adds namespaces not yet known to the discovery
	// result.
	updateNamespaces := func(ownedbindmounts []OwnedMountInfo) {
		for _, bmnt := range ownedbindmounts {
			nsid, nstype := nstypes.IDwithType(bmnt.Root)
			if nstype == nstypes.NaNS {
				continue // Play safe.
			}
			typeidx := TypeIndex(nstype)
			ns, ok := result.Namespaces[typeidx][nsid]
			if !ok {
				// As we haven't seen this namespace yet, record it with our
				// results.
				ns = NewNamespace(nstype, nsid, "")
				result.Namespaces[typeidx][nsid] = ns
				ns.(NamespaceConfigurer).SetRef(bmnt.MountPoint)
			}
			// Set the owning user namespace, but only if this ain't ;) a
			// user namespace and we actually got a owner namespace ID.
			if nstype != nstypes.CLONE_NEWUSER && bmnt.OwnernsID != nstypes.NoneID {
				ns.(NamespaceConfigurer).SetOwner(bmnt.OwnernsID)
			}
		}
	}
	// Find any bind-mounted namespaces in the current namespace we're running
	// in, and add them to the results.
	updateNamespaces(ownedBindMounts())
	// Now initialize a backlog with the mount namespaces we know so far,
	// because we need to visit them in order to potentially discover more
	// bind-mounted namespaces. These will then be added to the backlog if not
	// already known by then. And yes, this is ugly.
	mountnsBacklog := make([]Namespace, 0, len(result.Namespaces[MountNS]))
	for _, mntns := range result.Namespaces[MountNS] {
		mountnsBacklog = append(mountnsBacklog, mntns)
	}
	// In order to avoid multiple visits to the same namespace, keep track of
	// which mount namespaces not to visit again. This also includes the mount
	// namespace we've started our discovery in, as this will otherwise be
	// visited twice.
	visitedmntns := map[nstypes.NamespaceID]bool{}
	ownmntnsid, _ := rel.NamespacePath("/proc/self/ns/mnt").ID()
	ownusernsid, _ := rel.NamespacePath("/proc/self/ns/user").ID()
	visitedmntns[ownmntnsid] = true
	// Now try to clear the back log of mount namespaces to visit and to
	// search for further bind-mounted namespaces. Because we marked the
	// current mount namespace as already visited, we know after checking that
	// every mount namespace we'll find will be a different mount namespace,
	// so we need to re-execute when we want to switch into it (thanks to the
	// Go runtime which makes switching mount namespaces impossible after it
	// has spun up).
	for len(mountnsBacklog) > 0 {
		var mntns Namespace // NEVER merge this into the following pop operation!
		mntns, mountnsBacklog = mountnsBacklog[0], mountnsBacklog[1:]
		if _, ok := visitedmntns[mntns.ID()]; ok {
			continue // We already visited you ... next one!
		}
		// If we're running without the necessary privileges to change into
		// mount namespaces, but we are running under the user which is the
		// owner of the mount namespace, then we first gain the necessary
		// privileges by switching into the user namespace for the mount
		// namespace we're the owner (creator) of, and then can successfully
		// enter the mount namespaces. And yes, this is how Linux namespaces,
		// and especially the user namespaces and setns() are supposed to
		// work. Simplicity if for the World's most stable genius, we're going
		// for the real stuff instead.
		enterns := []Namespace{mntns}
		if usermntnsref, err := rel.NamespacePath(mntns.Ref()).User(); err == nil {
			usernsid, _ := usermntnsref.ID()
			usermntnsref.Close() // do not leak (again)
			if userns, ok := result.Namespaces[UserNS][usernsid]; ok && userns.ID() != ownusernsid {
				// Prepend the user namespace to the list of namespaces we
				// need to enter, due to the magic capabilities of entering
				// user namespaces. And, by the way, worst programming
				// language syntax ever, even more so than Perl. TECO isn't in
				// the competition, though.
				enterns = append([]Namespace{userns}, enterns...)
			}
		}
		// Finally, we can try to enter the mount namespace in order to find
		// out which namespace-related bind mounts might be found there...
		visitedmntns[mntns.ID()] = true
		var ownedbindmounts []OwnedMountInfo
		if err := ReexecIntoAction(
			"discover-nsfs-bindmounts", enterns, &ownedbindmounts); err == nil {
			// TODO: remember mount namespace for namespaces found, so we
			// still have a chance later to enter them by using the
			// bind-mounted reference in a different mount namespace.
			updateNamespaces(ownedbindmounts)
		} else {
			// TODO: for diagnosis:
			// fmt.Fprintf(os.Stderr, "failed: %s\n", err.Error())
		}
	}
}

// Register discoverNsfsBindmounts() as an action for re-execution.
func init() {
	reexec.Register("discover-nsfs-bindmounts", discoverNsfsBindmounts)
}

// discoverNsfsBindmounts is the reexec action run in a separate mount to
// gather information about bind-mounted namespaces in that other mount
// namespace. The information gathered is then serialized as JSON as sent back
// to the parent discovery process via stdout.
func discoverNsfsBindmounts() {
	if err := json.NewEncoder(os.Stdout).Encode(ownedBindMounts()); err != nil {
		panic(err.Error())
	}
}

// Returns a list of bind-mounts, including owning user namespace ID
// information, which are namespace bind-mounts.
func ownedBindMounts() []OwnedMountInfo {
	bindmounts := mntinfo.MountsOfType(-1, "nsfs")
	ownedbindmounts := make([]OwnedMountInfo, len(bindmounts))
	for idx := range bindmounts {
		bmnt := &bindmounts[idx] // avoid copying the mount information.
		// While we're in the correct mount namespace, we need to collect also
		// the information about the relation to the owning user space.
		var ownernsid nstypes.NamespaceID
		if usernsref, err := rel.NamespacePath(bmnt.MountPoint).User(); err == nil {
			ownernsid, _ = usernsref.ID()
			usernsref.Close()
		}
		ownedbindmounts[idx].Mountinfo = bmnt
		ownedbindmounts[idx].OwnernsID = ownernsid
	}
	return ownedbindmounts
}
