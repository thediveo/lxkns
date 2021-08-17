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
	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/mounts"
	"github.com/thediveo/lxkns/ops/mounteneer"
	"github.com/thediveo/lxkns/plural"
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
		mnteer, err := mounteneer.NewWithMountNamespace(
			mountns,
			result.Namespaces[model.UserNS])
		if err != nil {
			log.Errorf("could not discover mount points in mnt:[%d]: %s",
				mountns.ID().Ino, err.Error())
			continue
		}
		log.Debugf("reading mount point information from bind-mounted mnt:[%d] at %s...",
			mountns.ID().Ino, mountns.Ref().String())
		// Warp speed Mr Sulu, through the proc root wormhole!
		mountpoints := mntinfo.MountsOfPid(int(mnteer.PID()))
		mnteer.Close()
		log.Debugf("mnt:[%d] contains %s",
			mountns.ID().Ino, plural.Elements(len(mountpoints), "mount points"))
		mountpointtotal += len(mountpoints)
		result.Mounts[mntid] = mounts.NewMountPathMap(mountpoints)
	}
	log.Infof("found %s in %s",
		plural.Elements(mountpointtotal, "mount points"),
		plural.Elements(len(result.Mounts), "%s namespaces", "mount"))
}
