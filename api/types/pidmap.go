// Copyright 2020 Harald Albrecht.
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
	"bytes"
	"encoding/json"
	"errors"

	"github.com/thediveo/lxkns/internal/namespaces"
	pm "github.com/thediveo/lxkns/internal/pidmap"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// PIDMap is the (Digital) Twin of an model.PIDMap and can be marshalled and
// unmarshalled to and from JSON. Nota bene: PIDMap is a small object, so it
// should simply be passed around by value.
//
//
// To marshal from an existing model.PIDMap:
//
//   pm := NewPIDMap(WithPIDMap(mypidmap))
//   out, err := json.Marshal(pm)
//
// To unmarshal into a fresh model.PIDMap, not caring about PID namespace
// details beyond PID namespace IDs, simply call NewPIDMap() without any
// options:
//
//  pm := NewPIDMap()
//
// On purpose, the external JSON representation of a PIDMap is reduced compared
// to an model.PIDMap: this optimizes the transfer size by marshalling only the
// absolutely necessary information necessary to recreate an model.PIDMap on
// unmarshalling. In contrast, the process-internal model.PIDMap trades memory
// consumption for performance, in oder to speed up translating PIDs between
// different PID namespaces.
type PIDMap struct {
	// The real PID map we wrap for the purpose of un/marshalling.
	PIDMap model.PIDMapper
	// An optional PID namespace map to reuse for resolving PID namespace
	// references; this avoids PIDMaps having to create their own "minimalist"
	// PID namespace objects during unmarshalling.
	PIDns model.NamespaceMap
}

// NewPIDMap creates a new twin of either an existing model.PIDMap or
// allocates a new and empty model.PIDMap.
func NewPIDMap(opts ...NewPIDMapOption) PIDMap {
	pidmap := PIDMap{}
	for _, opt := range opts {
		opt(&pidmap)
	}
	// Initialize wrapped PID map and PID namespaces map if not having been
	// set by an option by now.
	if pidmap.PIDMap == nil {
		pidmap.PIDMap = pm.PIDMap{}
	}
	if pidmap.PIDns == nil {
		pidmap.PIDns = model.NamespaceMap{}
	}
	return pidmap
}

// NewPIDMapOption defines so-called functional options to be used with
// NewPIDMap().
type NewPIDMapOption func(newpidmap *PIDMap)

// WithPIDMap configures a new PIDMap to wrap an existing model.PIDMap; either
// for marshalling an existing PIDMap or to unmarshal into a pre-allocated
// PIDMap.
func WithPIDMap(pidmap model.PIDMapper) NewPIDMapOption {
	return func(npm *PIDMap) {
		npm.PIDMap = pidmap
	}
}

// WithPIDNamespaces configures a new PIDMap to use an already known map of
// PID namespaces.
func WithPIDNamespaces(pidnsmap model.NamespaceMap) NewPIDMapOption {
	return func(npm *PIDMap) {
		npm.PIDns = pidnsmap
	}
}

// namespacedPID is the JSON representation of a namespaced PID: instead of
// referencing an in-process Namespace object, the external representation can
// only reference the namespace ID (in particular, the inode number).
type namespacedPID struct {
	PID         model.PIDType `json:"pid"`  // process ID as seen in PID namespace.
	NamespaceID uint64        `json:"nsid"` // PID namespace ID.
}
type namespacedPIDs []namespacedPID

// MarshalJSON emits a PIDMap as JSON. To reduce the transfer volume, this
// method only emits enough table data for UnmarshalJSON() later being able to
// regenerate the full table.
func (pidmap PIDMap) MarshalJSON() ([]byte, error) {
	pidmapper := pidmap.PIDMap.(pm.PIDMap)
	b := bytes.Buffer{}
	b.WriteRune('[')
	// Remember the processes for which we've already emitted their namespaced
	// PIDs.
	pidsdone := map[model.PIDType]bool{}
	first := true
	for _, pids := range pidmapper {
		// Did we already handle this process? If yes, we can skip its
		// potentially multiple keys (namespaced PIDs of the process).
		if _, ok := pidsdone[pids[len(pids)-1].PID]; ok {
			continue
		}
		pidsdone[pids[len(pids)-1].PID] = true
		// Separate array items with commas.
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('[')
		for idx, nspid := range pids {
			if idx > 0 {
				b.WriteRune(',')
			}
			out, _ := json.Marshal(namespacedPID{PID: nspid.PID, NamespaceID: nspid.PIDNS.ID().Ino})
			b.Write(out)
		}
		b.WriteRune(']')
	}
	b.WriteRune(']')
	return b.Bytes(), nil
}

// UnmarshalJSON converts the textual JSON representation of a PID map back into
// the binary object state.
func (pidmap *PIDMap) UnmarshalJSON(data []byte) error {
	pidmapper := pidmap.PIDMap.(pm.PIDMap)
	// We begin with reading in the PID (translation) map as an array
	// containing arrays of namespaced PIDs; the namespaces are PID namespaces
	// simply referenced by their namespace IDs (inode numbers).
	aux := []namespacedPIDs{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for _, pids := range aux {
		if len(pids) == 0 {
			return errors.New("PIDMap: invalid empty list of namespaced PIDs")
		}
		//
		nspids := make(model.NamespacedPIDs, len(pids))
		for idx, nspid := range pids {
			pidnsid := species.NamespaceIDfromInode(nspid.NamespaceID)
			pidns, ok := pidmap.PIDns[pidnsid]
			if !ok {
				pidns = namespaces.New(species.CLONE_NEWPID, pidnsid, nil)
				pidmap.PIDns[pidnsid] = pidns
			}
			nspids[idx] = model.NamespacedPID{
				PID:   pids[idx].PID,
				PIDNS: pidns,
			}
		}
		// Now index the namespaced PID list by the individual namespaced
		// PIDs, so we can later quickly look up the list of namespaced PIDs
		// for this process using (PID, PID-namespace).
		for _, namespacedpid := range nspids {
			pidmapper[namespacedpid] = nspids
		}
	}
	return nil
}
