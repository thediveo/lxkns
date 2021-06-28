// Discovers mount paths with mount points in mount namespaces. This discovery
// requires an explicit opt-in. It needs to be run after the mount namespaces
// have been discovered. In consequence, it cannot be run without mount
// namespace discovery.

// Copyright 2021 Harald Albrecht.
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
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/mounts"
	"github.com/thediveo/lxkns/species"
)

// NamespacedMountPathMap maps mount namespaces identified by their namespace ID
// to their corresponding mount path maps.
type NamespacedMountPathMap map[species.NamespaceID]mounts.MountPathMap

// discoverFromMountinfo discovers the mount paths with their mount points in
// mount namespaces. API users must have opted in not only to this discovery
// step but must have also enabled discovery of mount namespaces. Otherwise,
// this step will be skipped.
func discoverFromMountinfo(_ species.NamespaceType, _ string, result *DiscoveryResult) {
	if !result.Options.DiscoverMounts {
		log.Infof("skipping discovery of mount paths and mount points")
		return
	}
	if result.Options.NamespaceTypes&species.CLONE_NEWNS == 0 {
		log.Warnf("mount namespace discovery skipped, so skipping mount path and points discovery")
	}
	log.Debugf("discovering namespaced mount paths and mount points...")
	// For every discovered mount namespace read its mount points, then
	// determine the mount paths and the mount point visibility from this
	// information.
	result.Mounts = NamespacedMountPathMap{}
	mountpointtotal := 0
	for mntid, mountns := range result.Namespaces[model.MountNS] {
		var mountpoints []mntinfo.Mountinfo
		if ealdorman := mountns.Ealdorman(); ealdorman != nil {
			// As we have a process attached to the mount namespace in question,
			// we can directly gather the mount point information from that
			// process -- albeit there's a slight catch here: if the ealdorman
			// has chroot'ed then we won't get the full original picture.
			mountpoints = mntinfo.MountsOfPid(int(ealdorman.PID))
		} else {
			// This is a bind-mounted mount namespace without any process at the
			// moment. In order to be able to get mount point information for
			// such a process-less mount namespace, we fork and re-execute a
			// discovery probe process in the mount namespace and that then
			// reads its own mount point information from that mount namespace.
			log.Debugf("reading mount point information from bind-mounted mnt:[%d] (%q)...",
				mountns.ID().Ino, mountns.Ref())
			// Warp speed Mr Sulu, through the proc root wormhole! (TODO: this
			// might need some more generalization in the future -- no, not the
			// Star Treck future -- as to handle bind-mounted mount namespaces
			// that are bind-mounted in other mount namespaces, et cetera.)
			wormholedmountns := namespaces.New(species.CLONE_NEWNS, mountns.ID(),
				"/proc/1/root"+mountns.Ref())
			enterns := MountEnterNamespaces(wormholedmountns, result.Namespaces)
			if err := ReexecIntoAction(
				"discover-mountinfo", enterns, &mountpoints); err != nil {
				log.Errorf("could not discover mount points in mnt:[%d]: %s",
					mountns.ID().Ino, err.Error())
				continue
			}
		}
		log.Debugf("mnt:[%d] contains %d mount points",
			mountns.ID().Ino, len(mountpoints))
		mountpointtotal += len(mountpoints)
		result.Mounts[mntid] = mounts.NewMountPathMap(mountpoints)
	}
	log.Infof("found %d mount points in %d mount namespaces",
		mountpointtotal, len(result.Mounts))
}

// Register discoverMounts() as an action for re-execution.
func init() {
	reexec.Register("discover-mountinfo", discoverMountinfo)
}

// discoverMountinfo is the reexec action run in a separate (bind-mounted) mount
// namespace in order to retrieve the mount information a process attached to it
// would see. Since it is a bind-mounted mount namespace for which we didn't
// find any process attached to it, we're now temporarily playing this role
// ourselves in order to discover what the mount points in this mount namespace
// will be. The information gathered is then serialized as JSON as sent back to
// the parent discovery process via stdout.
func discoverMountinfo() {
	if err := json.NewEncoder(os.Stdout).Encode(mntinfo.Mounts()); err != nil {
		panic(err.Error())
	}
}
