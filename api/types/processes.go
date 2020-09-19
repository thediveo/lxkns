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
	"fmt"
	"strconv"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// ProcessTable is the JSON serializable (digital!) twin to the process table
// returned from discoveries. The processes (process tree) is represented in
// JSON as a JSON object, where the members (keys) are the stringified PIDs
// and the values are process objects.
//
// In order to unmarshal a ProcessTable a namespace dictionary is required,
// which can either be prefilled or empty: it is used to share the namespace
// objects with the same ID between individual process objects in the table.
//
// Additionally, a ProcessTable can be primed with ("preliminary") Process
// objects. In this case, these process objects will be reused and updated
// with the new state. Please see also the Get() method, which will
// automatically do priming for yet unknown PIDs.
type ProcessTable struct {
	model.ProcessTable
	Namespaces *NamespacesDict // for resolving (and priming) namespace references
}

// Get always(!) returns a Process object with the given PID. When the process
// is already known, then it is returned, else a new preliminary process
// object gets created, registered, and returned instead. Preliminary process
// objects have only their PID set, but nothing else with the sole exception
// for the list of child processes being initialized.
func (p *ProcessTable) Get(pid model.PIDType) *model.Process {
	proc, ok := p.ProcessTable[pid]
	if !ok {
		proc = &model.Process{
			PID:      pid,
			Children: []*model.Process{},
		}
		p.ProcessTable[pid] = proc
	}
	return proc
}

// MarshalJSON emits the JSON textual representation of a complete process
// table.
func (p *ProcessTable) MarshalJSON() ([]byte, error) {
	// Similar to Golang's mapEncoder.encode, we iterate over the key-value
	// pairs ourselves, because we need to serialize alias types for the
	// individual process values, not the process values verbatim. By
	// iterating ourselves, we avoid building a new transient map with process
	// alias objects.
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for _, proc := range p.ProcessTable {
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(strconv.Itoa(int(proc.PID)))
		b.WriteString(`":`)
		procjson, err := json.Marshal((*Process)(proc))
		if err != nil {
			return nil, err
		}
		b.Write(procjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// UnmarshalJSON reads in the textual JSON representation of a complete
// process table. It makes use of the namespace object dictionary associated
// with this process table instance.
func (p *ProcessTable) UnmarshalJSON(data []byte) error {
	// Unmarshal all the processes using our "JSON-empowered" type of Process,
	// which we will later need to post-process in order to resolve the
	// process fields which we don't serialize and which allow easy
	// navigation, et cetera. Since a process table consists of objects of
	// objects of objects, we play it safe by using json.RawMessage, so the
	// json package can still do the generic JSON parsing part for us here so
	// that we don't need to count brackets, quotes, handle escapes, and the
	// other JSON hell.
	aux := map[model.PIDType]json.RawMessage{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if p.ProcessTable == nil {
		p.ProcessTable = model.ProcessTable{}
	}
	for _, rawproc := range aux {
		proc := model.Process{}
		if err := (*Process)(&proc).unmarshalJSON(rawproc, p.Namespaces); err != nil {
			return err
		}
		*p.Get(proc.PID) = proc
	}
	// Scan through the processes and resolve the parent-child process
	// relationships, based on the PPIDs and PIDs.
	for _, proc := range p.ProcessTable {
		if pproc, ok := p.ProcessTable[proc.PPID]; ok {
			proc.Parent = pproc
			pproc.Children = append(pproc.Children, proc)
		}
	}
	return nil
}

// Process is the JSON representation of the information about a single process.
// The Process type is designed to be used under the hood of ProcessTable and
// not directly by 3rd (external) party users.
type Process model.Process

// MarshalJSON emits the textual JSON representation of a single process.
//
// Please note that marshalling uses a pointer receiver, so make sure to have
// a *Process when unmarshalling, as otherwise this method won't get called.
// The rationale for a pointer receiver as opposed to a value receiver is that
// as a Process contains some more data "beyond two ints", we don't want them
// to be passed around as values all the time, copying and copying again.
//
// And now some sour tangarine somewhere surely is claiming this total Golang
// design fubar to be absolutely great, innit?
func (p *Process) MarshalJSON() ([]byte, error) {
	// Using an anonymous alias structure allows us to override serialization
	// of the namespaces the process is attached to: we just want them as
	// typed references (namespace type and ID), not as deeply serialized
	// first-class data elements.
	return json.Marshal(&struct {
		Namespaces *NamespacesSetReferences `json:"namespaces"`
		*model.Process
	}{
		Namespaces: (*NamespacesSetReferences)(&p.Namespaces),
		Process:    (*model.Process)(p),
	})
}

// UnmarshalJSON simply panics in order to clearly indicate that Process is
// not to be unmarshalled without a namespace dictionary to find existing
// namespaces in or add new ones just learnt to. Unfortunately, Golang's
// generic json (un)marshalling mechanism doesn't allow "contexts".
func (p *Process) UnmarshalJSON(data []byte) error {
	panic("cannot directly unmarshal lxkns.api.types.Process")
}

// unmarshalJSON reads in the textual JSON representation of a single process.
// It uses the associated namespace dictionary to resolve existing references
// into namespace objects and also adds missing namespaces.
func (p *Process) unmarshalJSON(data []byte, allns *NamespacesDict) error {
	// While we unmarshal "most" of the process data using json's automated
	// mechanics, we need to deal with the namespaces a process is attached to
	// separately. Because we need context for the namespaces, we do it
	// manually and then extract only the parts we need here.
	aux := struct {
		Namespaces json.RawMessage `json:"namespaces"`
		*model.Process
	}{
		Process: (*model.Process)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if p.PID <= 0 {
		return errors.New("Process invalid PID")
	}
	if p.PPID < 0 {
		return errors.New("Process invalid PPID")
	}
	if err := (*NamespacesSetReferences)(&p.Namespaces).unmarshalJSON(aux.Namespaces, allns); err != nil {
		return err
	}
	return nil
}

// NamespacesSetReferences is the JSON representation of a set of typed
// namespace ID references and thus the JSON face to model.NamespaceSet. The
// set of namespaces is represented in form of a JSON object with the object
// keys being the namespace types and the IDs then being the number values.
// Other namespace details are completely ignored, these are on purpose not
// repeated for each and every process in a potentially large process table.
type NamespacesSetReferences model.NamespacesSet

// MarshalJSON emits the textual JSON representation of a set of typed
// namespace references. Please note that it emits only references in form of
// namespace IDs only (without device numbers, see also the discussion about
// bolted horses from open barn dors in the architecture documentation).
func (n *NamespacesSetReferences) MarshalJSON() ([]byte, error) {
	// Since we don't to marshal all the namespace object details, but only
	// references, we need to do everything manually. While we could slightly
	// simplify things by first building a suitable object which we then could
	// marshal using json.Marshal() we instead directly emit the final JSON
	// textual representation without an intermediate object.
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for nsidx, ns := range n {
		// Skip unreferenced namespace types to avoid unnecessary noise.
		if ns == nil {
			continue
		}
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(model.TypesByIndex[nsidx].Name())
		b.WriteString(`":`)
		nsjson, err := json.Marshal(ns.ID().Ino)
		if err != nil {
			return nil, err
		}
		b.Write(nsjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// UnmarshalJSON simply panics in order to clearly indicate that
// TypedNamespacesSet are not to be unmarshalled without a namespace
// dictionary to find existing namespaces in or add new ones just learnt to.
// Unfortunately, Golang's generic json (un)marshalling mechanism doesn't
// allow "contexts".
func (n NamespacesSetReferences) UnmarshalJSON(data []byte) error {
	panic("cannot directly unmarshal TypedNamespacesSet")
}

// unmarshalJSON reads in the textual JSON representation of a set of typed
// namespace references. It uses a namespace object dictionary in order to
// reuse already existing namespace objects and also updates missing entries.
func (n *NamespacesSetReferences) unmarshalJSON(data []byte, allns *NamespacesDict) error {
	// Just get the typed namespace references as a properly key-value typed
	// map, so we can easily work on it next.
	rawns := map[string]uint64{}
	if err := json.Unmarshal(data, &rawns); err != nil {
		return err
	}
	// Now check that the types are okay and then reference (or create)
	// appropriate namespace objects. For this, we need the "context" which as
	// a dictionary of all namespaces already known.
	for nstypename, id := range rawns {
		nstype := species.NameToType(nstypename)
		if nstype == 0 {
			return fmt.Errorf("invalid namespace type %q", nstypename)
		}
		nstypeidx := model.TypeIndex(nstype)
		nsid := species.NamespaceIDfromInode(id)
		n[nstypeidx] = allns.Get(nsid, nstype)
	}
	return nil
}
