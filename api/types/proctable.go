// Copyright 2022 Harald Albrecht.
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
	"strconv"

	"github.com/thediveo/lxkns/model"
)

// ProcessTable is the JSON serializable (digital!) twin to the process table
// returned from discoveries. The processes (process tree) is represented in
// JSON as a JSON object, where the members (keys) are the stringified PIDs and
// the values are process objects.
//
// In order to unmarshal a ProcessTable a namespace dictionary is required,
// which can either be prefilled or empty: it is used to share the namespace
// objects with the same ID between individual process objects in the table.
//
// Additionally, a ProcessTable can be primed with ("preliminary") Process
// objects. In this case, these process objects will be reused and updated with
// the new state. Please see also the [ProcessTable.Get] method, which will
// automatically do priming for yet unknown PIDs.
type ProcessTable struct {
	model.ProcessTable
	Namespaces *NamespacesDict // for resolving (and priming) namespace references
}

// NewProcessTable creates a new process table that can be un/marshalled from or
// to JSON. Without any options, the process table returned can be used for
// unmarshalling right from the start. For marshalling an existing (hopefully
// filled) process table, use the [WithProcessTable] option to specify the
// process table to use.
func NewProcessTable(opts ...NewProcessTableOption) ProcessTable {
	proctable := ProcessTable{}
	for _, opt := range opts {
		opt(&proctable)
	}
	if proctable.ProcessTable == nil {
		proctable.ProcessTable = model.ProcessTable{}
	}
	if proctable.Namespaces == nil {
		proctable.Namespaces = NewNamespacesDict(nil)
	}
	return proctable
}

// NewProcessTableOption defines so-called functional options to be used with
// [NewProcessTable].
type NewProcessTableOption func(newproctable *ProcessTable)

// WithProcessTable specifies an existing (model) process table to use for
// marshalling.
func WithProcessTable(proctable model.ProcessTable) NewProcessTableOption {
	return func(npt *ProcessTable) {
		npt.ProcessTable = proctable
	}
}

// WithNamespacesDict specifies an existing namespaces dictionary to make use of
// while unmarshalling the namespace references of processes in a process table.
func WithNamespacesDict(nsdict *NamespacesDict) NewProcessTableOption {
	return func(npt *ProcessTable) {
		npt.Namespaces = nsdict
	}
}

// Get always(!) returns a [model.Process] object with the given PID. When the
// process is already known, then it is returned, else a new preliminary process
// object gets created, registered, and returned instead. Preliminary process
// objects have only their PID set, but nothing else with the sole exception for
// the list of child processes being initialized.
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
