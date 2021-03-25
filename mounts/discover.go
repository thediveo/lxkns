// Copyright 2021 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package mounts

import (
	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// NamespacedMountPathMap maps mount namespaces identified by their namespace ID
// to their corresponding mount path maps.
type NamespacedMountPathMap map[species.NamespaceID]MountPathMap

// Discover returns the mount information for the set of mount namespaces
// specified in the given a map of mount namespaces.
func Discover(mountnamespaces model.NamespaceMap) (mntmpmaps NamespacedMountPathMap) {
	mntmpmaps = NamespacedMountPathMap{}
	for mntid, mountns := range mountnamespaces {
		if ealdorman := mountns.Ealdorman(); ealdorman != nil {
			mntpathmap := NewMountPathMap(mntinfo.MountsOfPid(int(ealdorman.PID)))
			mntmpmaps[mntid] = mntpathmap
		}
	}
	return
}
