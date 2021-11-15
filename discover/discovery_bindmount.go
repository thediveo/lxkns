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
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/plural"
	"github.com/thediveo/lxkns/species"
)

// BindmountedNamespaceInfo describes a bind-mounted namespace in some (other)
// mount namespace, including the owning user namespace ID, so we can later
// correctly set up the ownership relations in the discovery results.
type BindmountedNamespaceInfo struct {
	ID        species.NamespaceID
	Type      species.NamespaceType
	Ref       model.NamespaceRef
	OwnernsID species.NamespaceID
}

// discoverBindmounts checks bind-mounts to discover namespaces we haven't found
// so far in the process' joined namespaces. This discovery function is designed
// to be run only once per discovery: but it will search not only in the current
// mount namespace, but also in other mount namespaces (subject to having
// capabilities in them).
func discoverBindmounts(_ species.NamespaceType, _ string, result *Result) {
	if !result.Options.ScanBindmounts {
		log.Infof("skipping discovery of bind-mounted namespaces")
		return
	}
	log.Debugf("starting discovery of bind-mounted namespaces...")
	total := 0
	// In order to avoid multiple visits to the same namespace, keep track of
	// which mount namespaces not to visit again.
	visitedmntns := map[species.NamespaceID]struct{}{}
	// Now initialize a backlog with the mount namespaces we know so far,
	// because we need to visit them in order to potentially discover more
	// bind-mounted namespaces. These will then be added to the backlog if not
	// already known by then. And yes, this is ugly.
	mountnsBacklog := make([]model.Namespace, 0, len(result.Namespaces[model.MountNS]))
	for _, mntns := range result.Namespaces[model.MountNS] {
		mountnsBacklog = append(mountnsBacklog, mntns)
	}
	// Iterating over the map of mount namespaces results in non-deterministic
	// order. Normally, this wouldn't be of any concern to us ... but
	// unfortunately, in our case there's a catch and that's due to mount point
	// propagation between mount namespaces. In particular, as Docker
	// bind-mounts the network namespaces of the containers it manages into a
	// place where they propagate to certain other mount namespaces, when the
	// ealdorman container process switches into a different network namespace
	// we thus end up with a randomly chosen bind-mount path. By sorting the
	// mount namespaces by their IDs, we end up with the initial mount namespace
	// always being first and thus the first one to turn up bind-mounted Docker
	// network namespaces without container processes. This way, we're ensuring
	// stability of network namespace references.
	sort.Slice(mountnsBacklog, func(i, j int) bool {
		return mountnsBacklog[i].ID().Dev < mountnsBacklog[j].ID().Dev
	})

	// Helper function which adds namespaces not yet known to the discovery
	// result. We keep this inline in order to allow the helper to access the
	// outer result.Namespaces map and easily update it.
	updateNamespaces := func(ownedbindmounts []BindmountedNamespaceInfo) {
		for _, bmntns := range ownedbindmounts {
			// Did we see this bind-mounted namespace already elsewhere...?
			typeidx := model.TypeIndex(bmntns.Type)
			ns, ok := result.Namespaces[typeidx][bmntns.ID]
			if !ok {
				// As we haven't seen this namespace yet, record it with our
				// results.
				ns = namespaces.New(bmntns.Type, bmntns.ID, nil)
				result.Namespaces[typeidx][bmntns.ID] = ns
				ns.(namespaces.NamespaceConfigurer).SetRef(bmntns.Ref)
				log.Debugf("found bind-mounted namespace %s:[%d] at %s",
					bmntns.Type.Name(), bmntns.ID.Ino, bmntns.Ref.String())
				total++
				// And if its a mount namespace we haven't yet visited, add it
				// to our backlock.
				if bmntns.Type == species.CLONE_NEWNS {
					if _, ok := visitedmntns[bmntns.ID]; !ok {
						mountnsBacklog = append(mountnsBacklog, ns)
					}
				}
			}
			// Set the owning user namespace, but only if this ain't ;) a
			// user namespace and we actually got a owner namespace ID.
			if bmntns.Type != species.CLONE_NEWUSER && bmntns.OwnernsID != species.NoneID {
				ns.(namespaces.NamespaceConfigurer).SetOwner(bmntns.OwnernsID)
			}
		}
	}

	// Now try to clear the back log of mount namespaces to visit and to
	// search for further bind-mounted namespaces.
	for len(mountnsBacklog) > 0 {
		var mntns model.Namespace // NEVER merge this into the following pop operation!
		mntns, mountnsBacklog = mountnsBacklog[0], mountnsBacklog[1:]
		if _, ok := visitedmntns[mntns.ID()]; ok {
			continue // We already visited you ... next one!
		}
		log.Debugf("scanning mnt:[%d] (%s) for bind-mounted namespaces...",
			mntns.ID().Ino, refString(mntns, result))
		visitedmntns[mntns.ID()] = struct{}{}

		mnteer, err := mountineer.NewWithMountNamespace(mntns, result.Namespaces[model.MountNS])
		if err != nil {
			log.Errorf("cannot open mnt:[%d] (reference: %s) for VFS operations: %s",
				mntns.ID().Ino, mntns.Ref().String(), err)
			continue
		}
		ownedbindmounts := ownedBindMounts(mnteer)
		mnteer.Close()
		updateNamespaces(ownedbindmounts)
	}
	log.Infof("found %s", plural.Elements(total, "bind-mounted namespaces"))
}

// refString returns a printable namespace reference, additionally resolving
// /proc-based reference elements to the names of their corresponding processes,
// if found in the additionally specified process table (from the discovery
// result).
func refString(mntns model.Namespace, r *Result) string {
	refs := mntns.Ref()
	s := []string{}
	for _, ref := range refs {
		if strings.HasPrefix(ref, "/proc/") {
			if f := strings.SplitN(ref, "/", 4); len(f) >= 3 {
				if pid, err := strconv.ParseUint(f[2], 10, 32); err == nil {
					if proc := r.Processes[model.PIDType(pid)]; proc != nil {
						s = append(s, fmt.Sprintf("%s[=%s]", ref, proc.Name))
					}
				}
			}
		}
		s = append(s, ref)
	}
	return strings.Join(s, "»")
}

// Returns a list of bind-mounted namespaces for process with PID, including
// owning user namespace ID information.
func ownedBindMounts(mnteer *mountineer.Mountineer) []BindmountedNamespaceInfo {
	// Please note that while the mount details of /proc/[PID]/mountinfo tell us
	// about bind-mounted namespaces with their types and inodes, they don't
	// tell us the device IDs of those namespaces. Argh, again we need to go
	// through hoops and loops just in order to satisfy Eric Biederman's dire
	// warning of a future where we will need to deal with multiple namespace
	// filesystem types. Some like it complicated.
	bindmounts := mntinfo.MountsOfType(int(mnteer.PID()), "nsfs")
	ownedbindmounts := make([]BindmountedNamespaceInfo, 0, len(bindmounts))
	for _, bindmount := range bindmounts {
		var ownernsid species.NamespaceID
		path, err := mnteer.Resolve(bindmount.MountPoint)
		if err != nil {
			log.Errorf("cannot resolve reference %s: %s",
				bindmount.MountPoint, err.Error())
			continue
		}
		log.Debugf("mount point for %s at %s", bindmount.Root, path)
		// Get the type and ID of the bind-mounted namespace. IDwithType now
		// always returns a complete ID consisting of dev number and inode
		// number, even though the specified string lacks the dev number (which
		// IDwithType) will get from elsewhere.
		nsid, nstype := species.IDwithType(bindmount.Root)
		// Also collect information about the relation to the owning user space.
		ns := ops.NamespacePath(path)
		if usernsref, err := ns.User(); err == nil {
			ownernsid, _ = usernsref.ID()
			_ = usernsref.(io.Closer).Close() // do not leak.
		} else {
			// log an error, but otherwise ignore the missing ownership in order
			// to create an albeit incomplete namespace entry.
			log.Errorf(err.Error())
		}
		ownedbindmounts = append(ownedbindmounts, BindmountedNamespaceInfo{
			Type:      nstype,
			ID:        nsid,
			Ref:       append(mnteer.Ref(), bindmount.MountPoint),
			OwnernsID: ownernsid,
		})
	}
	return ownedbindmounts
}
