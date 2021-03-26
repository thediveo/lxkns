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

	"github.com/thediveo/lxkns/mounts"
)

type NamespacedMountMap struct {
	*mounts.NamespacedMountPathMap
}

func (m *NamespacedMountMap) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (m *NamespacedMountMap) UnmarshalJSON(data []byte) error {
	return nil
}

// MountPathMap is a JSON marshallable mount path map.
type MountPathMap mounts.MountPathMap

// MountPath wraps a mounts.MountPathMap so that it can be marshalled with
// identifiers in place of mount path object references.
type MountPath struct {
	*mounts.MountPath
	ID       int   `json:"pathid"`   // unique mount path identifier, per mount namespace.
	ParentID int   `json:"parentid"` // ID of parent mount path, if any, otherwise 0.
	ChildIDs []int `json:"childids"` // IDs of child mount paths.
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
			ChildIDs:  []int{}, // ensure empty array instead of null
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
		for _, child := range mountpath.Children {
			mountpath.ChildIDs = append(mountpath.ChildIDs, mapwrapper[child.Path()].ID)
		}
	}
	return json.Marshal(mapwrapper)
}

// UnmarshalJSON unmarshals an object mapping mount paths to their mount
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
		mmp.Parent = idmap[mountpath.ParentID]
		for _, childid := range mountpath.ChildIDs {
			if childmp := idmap[childid]; childmp != nil {
				mmp.Children = append(mmp.Children, childmp)
			}
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
