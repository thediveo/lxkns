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

package discover

import (
	"io"

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
	visitedmntns := map[species.NamespaceID]bool{}
	// Now initialize a backlog with the mount namespaces we know so far,
	// because we need to visit them in order to potentially discover more
	// bind-mounted namespaces. These will then be added to the backlog if not
	// already known by then. And yes, this is ugly.
	mountnsBacklog := make([]model.Namespace, 0, len(result.Namespaces[model.MountNS]))
	for _, mntns := range result.Namespaces[model.MountNS] {
		mountnsBacklog = append(mountnsBacklog, mntns)
	}

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
			mntns.ID().Ino, mntns.Ref().String())
		visitedmntns[mntns.ID()] = true

		mnteer, err := mountineer.NewWithMountNamespace(mntns, result.Namespaces[model.MountNS])
		if err != nil {
			log.Errorf("cannot open mnt:[%d] (%s) for VFS operations: %s",
				mntns.ID().Ino, mntns.Ref().String(), err)
			continue
		}
		ownedbindmounts := ownedBindMounts(mnteer)
		mnteer.Close()
		updateNamespaces(ownedbindmounts)
	}
	log.Infof("found %s", plural.Elements(total, "bind-mounted namespaces"))
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
			log.Errorf("cannot resolve reference %s: %s", bindmount.MountPoint, err.Error())
		}
		log.Debugf("mount point for %s at %s", bindmount.Root, path)
		// Get the type of namespace, but ignore the inode number, because it
		// lacks the dev ID for a complete namespace ID.
		_, nstype := species.IDwithType(bindmount.Root)
		// Make sure to get the full namespace ID, not just the inode number.
		ns := ops.NamespacePath(path)
		nsid, err := ns.ID()
		if err != nil {
			// Ouch, we could not correctly get the namespace ID, so we log a
			// warning.
			log.Warnf("cannot read ID of namespace with reference: %s", err.Error())
			// And then we do some animal magic to come up at least with what
			// might be the namespace ID and type...
			if nsidr, nstyper := species.IDwithType(bindmount.Root); nsidr != species.NoneID && nstyper != 0 {
				nsid = nsidr
				nstype = nstyper
			}
		}
		// Also collect the information about the relation to the owning user
		// space.
		if usernsref, err := ns.User(); err == nil {
			ownernsid, _ = usernsref.ID()
			_ = usernsref.(io.Closer).Close() // do not leak.
		} else {
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
