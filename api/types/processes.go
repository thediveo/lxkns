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

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/species"
)

// ProcessTable is the JSON serializable (digital!) twin to the process table
// returned from discoveries. The processes (process tree) is represented in
// JSON as a JSON object, where the members (keys) are the stringified PIDs
// and the values are process objects.
type ProcessTable struct {
	lxkns.ProcessTable
	Namespaces lxkns.AllNamespaces // aux. namespace information
}

// MarshalJSON returns the JSON textual representation of a process table.
func (p ProcessTable) MarshalJSON() ([]byte, error) {
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
		procjson, err := json.Marshal(proc)
		if err != nil {
			return nil, err
		}
		b.Write(procjson)
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// Process is the JSON representation of the information about a single
// process. It is the stock lxkns.Process struct, but with an additional
// namespace set context required for unmarshalling.
type Process struct {
	*lxkns.Process
	AllNamespaces *lxkns.AllNamespaces `json:"-"`
}

// MarshalJSON emits the textual JSON representation of a single process.
//
// Please note that marshalling uses a pointer receiver, so make sure to have
// a *Process when unmarshalling, as otherwise this method won't get called.
// The rationale for a pointer receiver as opposed to a value receiver is that
// as a Process contains some more data "beyond two ints", we don't want them
// to be passed around as values all the time, copying and copying again.
//
// And now some tangarine somewhere surely is claiming this total Golang
// design fubar to be so GREAT and AWESOME...
func (p *Process) MarshalJSON() ([]byte, error) {
	// Using an anonymous alias structure allows us to override serialization
	// of the namespaces the process is attached to: we just want them as
	// typed references (namespace type and ID), not as deeply serialized
	// first-class data elements.
	return json.Marshal(&struct {
		Namespaces *TypedNamespacesSet `json:"namespaces"`
		*lxkns.Process
	}{
		Namespaces: (*TypedNamespacesSet)(&p.Process.Namespaces),
		Process:    p.Process,
	})
}

// UnmarshalJSON reads in the textual JSON representation of a single process.
// It uses the associated namespace dictionary to resolve existing references
// into namespace objects and also adds missing namespaces.
func (p *Process) UnmarshalJSON(data []byte) error {
	// While we unmarshal "most" of the process data using json's automated
	// mechanics, we need to deal with the namespaces a process is attached to
	// separately. Because we need context for the namespaces, we do it
	// manually and then extract only the parts we need here.
	aux := struct {
		Namespaces json.RawMessage `json:"namespaces"`
		*lxkns.Process
	}{
		Process: p.Process,
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Process == nil {
		return errors.New("missing or invalid process information")
	}
	p.Process = aux.Process
	if err := (*TypedNamespacesSet)(&p.Process.Namespaces).unmarshalJSON(aux.Namespaces, p.AllNamespaces); err != nil {
		return err
	}
	return nil
}

// TypedNamespaceSet is the JSON representation a set of typed
// namespace IDs. The set is represented as a JSON object, with the keys being
// the namespace types and the IDs then being the number values.
type TypedNamespacesSet lxkns.NamespacesSet

func (n TypedNamespacesSet) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for nsidx, ns := range n {
		if ns == nil {
			continue
		}
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		b.WriteRune('"')
		b.WriteString(lxkns.TypesByIndex[nsidx].Name())
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

// FIXME: implement
func (n TypedNamespacesSet) UnmarshalJSON(data []byte) error {
	panic("do not directly unmarshal TypedNamespacesSet")
}

// unmarshalJSON reads in the textual representation of a set of typed
// namespace references. It uses a namespace object dictionary in order to
// reuse already existing namespace objects and also updates missing entries.
func (n *TypedNamespacesSet) unmarshalJSON(data []byte, allns *lxkns.AllNamespaces) error {
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
		nstypeidx := lxkns.TypeIndex(nstype)
		nsid := species.NamespaceIDfromInode(id)
		ns, ok := allns[nstypeidx][nsid]
		if !ok {
			// While we can already create the namespace object with type and
			// ID, the remaining information needs to be filled in elsewhere
			// when unmarshalling the complete namespace information. Here,
			// we're just creating the "hulls".
			ns = lxkns.NewNamespace(nstype, nsid, "")
			allns[nstypeidx][nsid] = ns
		}
		n[nstypeidx] = ns
	}
	return nil
}
