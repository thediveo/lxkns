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

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

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
	// Ensure that a nil command line won't ever get marshalled as "null" but
	// instead as an empty command line; this avoids downstream things, such as
	// OpenAPI validators eating their heart out...
	cmdline := p.Cmdline
	if cmdline == nil {
		cmdline = []string{}
	}
	// Using an anonymous alias structure allows us to override serialization of
	// the namespaces the process is attached to: we just want them as typed
	// references (namespace type and ID), not as deeply serialized first-class
	// data elements. In contrast, the tasks of a process are flat and thus can
	// be marshalled without taking too many special measures.
	tasks := make([]*Task, 0, len(p.Tasks))
	for _, task := range p.Tasks {
		tasks = append(tasks, (*Task)(task))
	}
	return json.Marshal(&struct {
		Namespaces *NamespacesSetReferences `json:"namespaces"`
		Cmdline    []string                 `json:"cmdline"`
		Tasks      []*Task                  `json:"tasks"`
		*model.Process
	}{
		Namespaces: (*NamespacesSetReferences)(&p.Namespaces),
		Cmdline:    cmdline,
		Tasks:      tasks,
		Process:    (*model.Process)(p),
	})
}

// UnmarshalJSON simply panics in order to clearly indicate that Process is
// not to be unmarshalled without a namespace dictionary to find existing
// namespaces in or add new ones just learnt to. Unfortunately, Golang's
// generic json (un)marshalling mechanism doesn't allow "contexts".
func (p *Process) UnmarshalJSON(data []byte) error {
	panic("cannot directly unmarshal github.com/thediveo/lxkns/api/types.Process")
}

// unmarshalJSON reads in the textual JSON representation of a single process.
// It uses the associated namespace dictionary to resolve existing references
// into namespace objects and also adds missing namespaces. The task disctionary
// is additionally used to correctly allocate the same Task only once by its
// TID, even if already been seen before as a loose thread reference.
func (p *Process) unmarshalJSON(data []byte, allns *NamespacesDict) error {
	// While we unmarshal "most" of the process data using json's automated
	// mechanics, we need to deal with the namespaces a process is attached to
	// separately. Because we need context for the namespaces, we do it
	// manually and then extract only the parts we need here.
	aux := struct {
		Namespaces json.RawMessage   `json:"namespaces"`
		Tasks      []json.RawMessage `json:"tasks"`
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
	// Unmarshal the remaining Process fields that need further special
	// treatment.
	if err := (*NamespacesSetReferences)(&p.Namespaces).unmarshalJSON(aux.Namespaces, allns); err != nil {
		return err
	}
	tasks := make([]*model.Task, len(aux.Tasks))
	for idx, rawtask := range aux.Tasks {
		var auxtask struct {
			TID model.PIDType `json:"tid"`
			// ...ignore everything else that is present in the JSON
		}
		if err := json.Unmarshal(rawtask, &auxtask); err != nil {
			return err
		}
		if auxtask.TID <= 0 {
			return errors.New("Task of Process invalid TID")
		}
		task := allns.TaskTable.Get((*model.Process)(p), auxtask.TID)
		if err := (*Task)(task).unmarshalJSON(rawtask, allns); err != nil {
			return err
		}
		tasks[idx] = task
	}
	p.Tasks = tasks
	// Convert a potential null command line into an empty command line.
	if p.Cmdline == nil {
		p.Cmdline = []string{}
	}
	return nil
}

// NamespacesSetReferences is the JSON representation of a set of typed
// namespace ID references and thus the JSON face to [model.NamespaceSet]. The
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
