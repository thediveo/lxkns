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

// MountPathMap ...
type MountPathMap mounts.MountPathMap

type MountPath struct {
	*mounts.MountPath
	ID       int   `json:"pathid"`      // unique mount path identifier, per mount namespace.
	ParentID int   `json:"parentid"`    // ID of parent mount path, if any, otherwise 0.
	ChildIDs []int `json:"childrenids"` // IDs of child mount paths.
}

// MarshalJSON emits an object (map/dictionary) of mount paths with their mount
// point(s) all belonging a single mount namespace.
func (m MountPathMap) MarshalJSON() ([]byte, error) {
	jm := map[string]*MountPath{}
	id := 1
	// We first need to assign unique IDs to all our mount paths (wrapper
	// objects) within the same mount namespace. Fun fact: IDs depend on the
	// random order of iterating over the mount path map ;)
	for mountpath, mp := range m {
		jm[mountpath] = &MountPath{
			MountPath: mp,
			ID:        id,
			ChildIDs:  []int{}, // ensure empty array instead of null
		}
		id++
	}
	// Then we can set the ID references based on the object references. True,
	// receivers of the emitted JSON could do this perfectly themselves, but
	// then we're offering it for convenience.
	for _, mp := range jm {
		if mp.Parent != nil {
			mp.ParentID = jm[mp.Parent.Path()].ID
		}
		for _, child := range mp.Children {
			mp.ChildIDs = append(mp.ChildIDs, jm[child.Path()].ID)
		}
	}
	return json.Marshal(jm)
}

// UnmarshalJSON unmarshals an object mapping mount paths to their mount
// point(s).
func (m *MountPathMap) UnmarshalJSON(data []byte) error {
	jm := map[string]MountPath{}
	if err := json.Unmarshal(data, &jm); err != nil {
		return err
	}
	idmap := map[int]*mounts.MountPath{}
	for mountpath, mp := range jm {
		idmap[mp.ID] = mp.MountPath
		(*m)[mountpath] = mp.MountPath
	}
	for mountpath, mp := range jm {
		mmp := (*m)[mountpath]
		mmp.Parent = idmap[mp.ParentID]
		for _, childid := range mp.ChildIDs {
			if childmp := idmap[childid]; childmp != nil {
				mmp.Children = append(mmp.Children, childmp)
			}
		}
	}
	return nil
}
