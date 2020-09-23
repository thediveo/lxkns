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
	"fmt"
	"io"
	"os"

	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
)

// BindmountedNamespaceInfo describes a bind-mounted namespace in some (other)
// mount namespace, including the owning user namespace ID, so we can later
// correctly set up the ownership relations in the discovery results.
type BindmountedNamespaceInfo struct {
	ID        species.NamespaceID   `json:"id"`
	Type      species.NamespaceType `json:"type"`
	Path      string                `json:"path"`
	OwnernsID species.NamespaceID   `json:"ownernsid"`
	Log       []string              `json:"log"` // not strictly necessary, yet very helpful.
}

// discoverBindmounts checks bind-mounts to discover namespaces we haven't found
// so far in the process' joined namespaces. This discovery function is designed
// to be run only once per discovery: but it will search not only in the current
// mount namespace, but also in other mount namespaces (subject to having
// capabilities in them).
func discoverBindmounts(_ species.NamespaceType, _ string, result *DiscoveryResult) {
	if result.Options.SkipBindmounts {
		log.Infof("skipping discovery of bind-mounted namespaces")
		return
	}
	log.Debugf("starting discovery of bind-mounted namespaces...")
	total := 0
	// Helper function which adds namespaces not yet known to the discovery
	// result. We keep this inline in order to allow the helper to access the
	// outer result.Namespaces map and easily update it.
	updateNamespaces := func(ownedbindmounts []BindmountedNamespaceInfo) {
		for _, bmntns := range ownedbindmounts {
			// If there were errors noticed while trying to gather the
			// information about this specific bind-mounted namespace, then log
			// them now as errors from the main process.
			for _, l := range bmntns.Log {
				log.Errorf("namespace discovery error for %s:[%d] (%q): %s",
					bmntns.Type.Name(), bmntns.ID.Ino, bmntns.Path, l)
			}
			if bmntns.ID == species.NoneID {
				log.Errorf("skipping bind-mounted namespace at %q: could not discover namespace ID", bmntns.Path)
				continue
			}
			// log.Debugf("checking bind-mounted namespace [%d]", bmntns.ID.Ino)

			// Now we can finally look up whether we have seen this bind-mounted
			// namespace elsewhere...
			typeidx := model.TypeIndex(bmntns.Type)
			ns, ok := result.Namespaces[typeidx][bmntns.ID]
			if !ok {
				// As we haven't seen this namespace yet, record it with our
				// results.
				ns = namespaces.New(bmntns.Type, bmntns.ID, "")
				result.Namespaces[typeidx][bmntns.ID] = ns
				ns.(namespaces.NamespaceConfigurer).SetRef(bmntns.Path)
				log.Debugf("found namespace %s:[%d] bind-mounted at %q",
					bmntns.Type.Name(), bmntns.ID.Ino, bmntns.Path)
				total++
			}
			// Set the owning user namespace, but only if this ain't ;) a
			// user namespace and we actually got a owner namespace ID.
			if bmntns.Type != species.CLONE_NEWUSER && bmntns.OwnernsID != species.NoneID {
				ns.(namespaces.NamespaceConfigurer).SetOwner(bmntns.OwnernsID)
			}
		}
	}
	// In order to avoid multiple visits to the same namespace, keep track of
	// which mount namespaces not to visit again. This also includes the mount
	// namespace we've started our discovery in, as this will otherwise be
	// visited twice.
	visitedmntns := map[species.NamespaceID]bool{}
	ownmntnsid, _ := ops.NamespacePath("/proc/self/ns/mnt").ID()
	ownusernsid, _ := ops.NamespacePath("/proc/self/ns/user").ID()
	visitedmntns[ownmntnsid] = true
	// Find any bind-mounted namespaces in the current namespace we're running
	// in, and add them to the results.
	log.Debugf("scanning (own) mnt:[%d] (%q) for bind-mounted namespaces...",
		ownmntnsid.Ino, "/proc/self/ns/mnt")
	updateNamespaces(ownedBindMounts())
	// Now initialize a backlog with the mount namespaces we know so far,
	// because we need to visit them in order to potentially discover more
	// bind-mounted namespaces. These will then be added to the backlog if not
	// already known by then. And yes, this is ugly.
	mountnsBacklog := make([]model.Namespace, 0, len(result.Namespaces[model.MountNS]))
	for _, mntns := range result.Namespaces[model.MountNS] {
		mountnsBacklog = append(mountnsBacklog, mntns)
	}
	// Now try to clear the back log of mount namespaces to visit and to
	// search for further bind-mounted namespaces. Because we marked the
	// current mount namespace as already visited, we know after checking that
	// every mount namespace we'll find will be a different mount namespace,
	// so we need to re-execute when we want to switch into it (thanks to the
	// Go runtime which makes switching mount namespaces impossible after it
	// has spun up).
	for len(mountnsBacklog) > 0 {
		var mntns model.Namespace // NEVER merge this into the following pop operation!
		mntns, mountnsBacklog = mountnsBacklog[0], mountnsBacklog[1:]
		if _, ok := visitedmntns[mntns.ID()]; ok {
			continue // We already visited you ... next one!
		}
		log.Debugf("scanning mnt:[%d] (%q) for bind-mounted namespaces...",
			mntns.ID().Ino, mntns.Ref())
		// If we're running without the necessary privileges to change into
		// mount namespaces, but we are running under the user which is the
		// owner of the mount namespace, then we first gain the necessary
		// privileges by switching into the user namespace for the mount
		// namespace we're the owner (creator) of, and then can successfully
		// enter the mount namespaces. And yes, this is how Linux namespaces,
		// and especially the user namespaces and setns() are supposed to
		// work. Simplicity if for the World's most stable genius, we're going
		// for the real stuff instead.
		enterns := []model.Namespace{mntns}
		if usermntnsref, err := ops.NamespacePath(mntns.Ref()).User(); err == nil {
			usernsid, _ := usermntnsref.ID()
			// Do not leak, release user namespace immediately, as we're done with it.
			usermntnsref.(io.Closer).Close()
			if userns, ok := result.Namespaces[model.UserNS][usernsid]; ok &&
				userns.ID() != ownusernsid {
				// Prepend the user namespace to the list of namespaces we
				// need to enter, due to the magic capabilities of entering
				// user namespaces. And, by the way, worst programming
				// language syntax ever, even more so than Perl. TECO isn't in
				// the competition, though.
				enterns = append([]model.Namespace{userns}, enterns...)
			}
		}
		// Finally, we can try to enter the mount namespace in order to find
		// out which namespace-related bind mounts might be found there...
		visitedmntns[mntns.ID()] = true
		var ownedbindmounts []BindmountedNamespaceInfo
		if err := ReexecIntoAction(
			"discover-nsfs-bindmounts", enterns, &ownedbindmounts); err == nil {
			// TODO: remember mount namespace for namespaces found, so we
			// still have a chance later to enter them by using the
			// bind-mounted reference in a different mount namespace.
			updateNamespaces(ownedbindmounts)
		} else {
			log.Errorf("could not discover in mnt:[%d]: %s", mntns.ID().Ino, err.Error())
		}
	}
	log.Infof("found %d bind-mounted namespaces", total)
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

// Returns a list of bind-mounted namespaces, including owning user namespace ID
// information.
func ownedBindMounts() []BindmountedNamespaceInfo {
	// Please note that while the mount details of /proc/mountinfo tell us about
	// bind-mounted namespaces with their types and inodes, they don't tell us
	// the device IDs of those namespaces. Argh, again we need to go through
	// hoops and loops just in order to satisfy Eric Biederman's dire warning of
	// a future where we will need to deal with multiple namespace filesystem
	// types. Some like it complicated.
	bindmounts := mntinfo.MountsOfType(-1, "nsfs")
	ownedbindmounts := make([]BindmountedNamespaceInfo, len(bindmounts))
	for idx := range bindmounts {
		path := bindmounts[idx].MountPoint
		// Get the type of namespace, but ignore the inode number, because it
		// lacks the dev ID for a complete namespace ID.
		_, ownedbindmounts[idx].Type = species.IDwithType(bindmounts[idx].Root)
		// Make sure to get the full namespace ID, not just the inode number.
		// Argh. We must do this while still inside the correct mount namespace,
		// as otherwise the path might not exist, or even worse, it might point
		// to another namespace.
		ownedbindmounts[idx].Path = path
		ns := ops.NamespacePath(path)
		nsid, err := ns.ID()
		if err != nil {
			// Ouch, we could not correctly get the namespace ID, so we log an
			// error.
			ownedbindmounts[idx].Log = append(ownedbindmounts[idx].Log,
				fmt.Sprintf("while reading namespace ID: %s", err.Error()))
			// And then we do some animal magic to come up at least with what
			// might be the namespace ID and type...
			nsid, nstype := species.IDwithType(bindmounts[idx].Root)
			if nsid != species.NoneID && nstype != 0 {
				ownedbindmounts[idx].ID = nsid
				ownedbindmounts[idx].Type = nstype
			}
			continue
		}
		ownedbindmounts[idx].ID = nsid
		// While we're in the correct mount namespace, we need to collect also
		// the information about the relation to the owning user space.
		var ownernsid species.NamespaceID
		if usernsref, err := ns.User(); err == nil {
			ownernsid, _ = usernsref.ID()
			usernsref.(io.Closer).Close() // do not leak.
		} else {
			ownedbindmounts[idx].Log = append(ownedbindmounts[idx].Log, err.Error())
		}
		ownedbindmounts[idx].OwnernsID = ownernsid
	}
	return ownedbindmounts
}
