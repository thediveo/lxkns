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

package types

import (
	"encoding/json"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/mounts"
	"github.com/thediveo/lxkns/species"
)

// NamespacedMountMap is a JSON marshallable map from mount namespace
// identifiers (inode numbers only) to their respective mount path maps. The
// mount path maps further reference mount points.
type NamespacedMountMap discover.NamespacedMountPathMap

// MarshalJSON emits an object (map/dictionary) of mount namespace identifiers
// (inode numbers only) with their corresponding mount path maps.
func (m NamespacedMountMap) MarshalJSON() ([]byte, error) {
	wrapper := map[uint64]MountPathMap{}
	for mntnsid, mountpathmap := range m {
		wrapper[mntnsid.Ino] = MountPathMap(mountpathmap)
	}
	return json.Marshal(wrapper)
}

// UnmarshalJSON decodes an object (map/dictionary) of mount namespaced mount
// path maps (with mount points).
func (m *NamespacedMountMap) UnmarshalJSON(data []byte) error {
	var wrapper map[uint64]MountPathMap
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	if *m == nil {
		*m = NamespacedMountMap{}
	}
	for mntnsid, mountpathmap := range wrapper {
		(*m)[species.NamespaceIDfromInode(mntnsid)] = mounts.MountPathMap(mountpathmap)
	}
	return nil
}

// MountPathMap is a JSON marshallable mount path map.
type MountPathMap mounts.MountPathMap

// MountPath wraps a [mounts.MountPathMap] so that it can be marshalled with
// identifiers in place of mount path object references.
type MountPath struct {
	*mounts.MountPath
	ID       int `json:"pathid"`   // unique mount path identifier, per mount namespace.
	ParentID int `json:"parentid"` // ID of parent mount path, if any, otherwise 0.
}

// MarshalJSON emits an object (map/dictionary) of mount paths with their mount
// point(s) all belonging a single mount namespace.
func (m MountPathMap) MarshalJSON() ([]byte, error) {
	mapwrapper := map[string]*MountPath{}
	id := 1
	// We first need to assign unique IDs to all our mount paths (wrapper
	// objects) within the same mount namespace. Fun fact: IDs depend on the
	// random order of iterating over the mount path map ;)
	for path, mountpath := range m {
		mapwrapper[path] = &MountPath{
			MountPath: mountpath,
			ID:        id,
		}
		id++
	}
	// Then we can set the ID references based on the object references. True,
	// receivers of the emitted JSON could do this perfectly themselves, but
	// then we're offering it for convenience.
	for _, mountpath := range mapwrapper {
		if mountpath.Parent != nil {
			mountpath.ParentID = mapwrapper[mountpath.Parent.Path()].ID
		}
	}
	return json.Marshal(mapwrapper)
}

// UnmarshalJSON decodes an object mapping mount paths to their mount
// point(s). This restores not only the object hierarchy of the mount paths, but
// also of the mount points corresponding with the mount paths.
func (m *MountPathMap) UnmarshalJSON(data []byte) error {
	mapwrapper := map[string]MountPath{}
	if err := json.Unmarshal(data, &mapwrapper); err != nil {
		return err
	}
	if *m == nil {
		(*m) = MountPathMap{}
	}
	// During our first run we gather all mount path IDs and also map the mount
	// paths to their mount path objects (detail information).
	idmap := map[int]*mounts.MountPath{}
	for mountpath, mp := range mapwrapper {
		idmap[mp.ID] = mp.MountPath
		(*m)[mountpath] = mp.MountPath
	}
	// On our second run, we now resolve hierarchy IDs into object references in
	// order to link all mount paths into a single hierarchy.
	mountidmap := map[int]*mounts.MountPoint{}
	for path, mountpath := range mapwrapper {
		mmp := (*m)[path]
		if parent := idmap[mountpath.ParentID]; parent != nil {
			mmp.Parent = parent
			mmp.Parent.Children = append(mmp.Parent.Children, mmp)
		}
		// Remember the IDs of the mount points at this mount path in the map.
		for _, mount := range mountpath.Mounts {
			mountidmap[mount.MountID] = mount
		}
	}
	// Finally resolve the hierarchy of mount point reference IDs into object
	// references.
	for _, mount := range mountidmap {
		if parent := mountidmap[mount.ParentID]; parent != nil {
			mount.Parent = parent
			parent.Children = append(parent.Children, mount)
		}
	}
	return nil
}
