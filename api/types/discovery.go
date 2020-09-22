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
	"encoding/json"
	"fmt"

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// DiscoveryOptions is the (digital) twin of an lxkns DiscoveryOptions, which
// can be marshalled and unmarshalled to and from JSON.
type DiscoveryOptions lxkns.DiscoverOpts

// MarshalJSON emits discovery options as JSON.
func (doh DiscoveryOptions) MarshalJSON() ([]byte, error) {
	aux := struct {
		*lxkns.DiscoverOpts
		NamespaceTypes []string `json:"scanned-namespace-types"`
	}{
		DiscoverOpts: (*lxkns.DiscoverOpts)(&doh),
	}
	// Convert the bitmask of CLONE_NEWxxx into the usual textual type names for
	// namespaces, such as "net", et cetera.
	for bit := 1; bit != 0; bit <<= 1 {
		if doh.NamespaceTypes&species.NamespaceType(bit) != 0 {
			aux.NamespaceTypes = append(aux.NamespaceTypes, species.NamespaceType(bit).Name())
		}
	}
	return json.Marshal(aux)
}

func (doh *DiscoveryOptions) UnmarshalJSON(data []byte) error {
	aux := struct {
		*lxkns.DiscoverOpts
		NamespaceTypes []string `json:"scanned-namespace-types"`
	}{
		DiscoverOpts: (*lxkns.DiscoverOpts)(doh),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for _, nstypename := range aux.NamespaceTypes {
		nstype := species.NameToType(nstypename)
		if nstype == 0 {
			return fmt.Errorf("invalid type of namespace %q", nstypename)
		}
		doh.NamespaceTypes |= nstype
	}
	return nil
}

// DiscoveryResult is basically the (digital) twin of an lxkns DiscoveryResult,
// which can be marshalled and unmarshalled to and from JSON. Additionally, it
// acts as an extensible discovery result wrapper, which allows API users to
// freely add their own fields (with objects) to un/marshal additional result
// fields, as they see fit.
type DiscoveryResult struct {
	Fields          map[string]interface{}
	DiscoveryResult *lxkns.DiscoveryResult `json:"-"`
}

const (
	FieldDiscoveryOptions = "discovery-options"
	FieldNamespaces       = "namespaces"
	FieldProcesses        = "processes"
)

// NewDiscoveryResult returns a discovery result object ready for marshalling
// JSON into it or marshalling an existing lxkns discovery result.
func NewDiscoveryResult(opts ...NewDiscoveryResultOption) *DiscoveryResult {
	dr := &DiscoveryResult{Fields: map[string]interface{}{}}
	for _, opt := range opts {
		opt(dr)
	}
	if dr.DiscoveryResult == nil {
		dr.DiscoveryResult = &lxkns.DiscoveryResult{
			Namespaces: *model.NewAllNamespaces(),
			Processes:  model.ProcessTable{},
		}
	}
	dr.Fields[FieldDiscoveryOptions] = (*DiscoveryOptions)(&dr.DiscoveryResult.Options)
	nsdict := NewNamespacesDict(dr.DiscoveryResult)
	dr.Fields[FieldNamespaces] = nsdict
	pt := NewProcessTable(
		WithProcessTable(dr.DiscoveryResult.Processes),
		WithNamespacesDict(nsdict))
	dr.Fields[FieldProcesses] = &pt
	return dr
}

// FIXME:
type NewDiscoveryResultOption func(newdiscoveryresult *DiscoveryResult)

// FIXME:
func WithResult(result *lxkns.DiscoveryResult) NewDiscoveryResultOption {
	return func(ndr *DiscoveryResult) {
		ndr.DiscoveryResult = result
	}
}

// FIXME:
func WithElement(name string, obj interface{}) NewDiscoveryResultOption {
	return func(ndr *DiscoveryResult) {
		ndr.Fields[name] = obj
	}
}

// FIXME:
func (dr DiscoveryResult) Result() *lxkns.DiscoveryResult {
	return dr.DiscoveryResult
}

// FIXME:
func (dr DiscoveryResult) Processes() model.ProcessTable {
	return dr.DiscoveryResult.Processes
}

// FIXME:
func (dr DiscoveryResult) Get(name string) interface{} {
	return dr.Fields[name]
}

// MarshalJSON marshals discovery results into their JSON textual
// representation.
func (dr DiscoveryResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(dr.Fields)
}

// UnmarshalJSON unmarshals discovery results from JSON into a DiscoveryResult
// object, usually obtained with NewDiscoveryResult() first.
func (dr *DiscoveryResult) UnmarshalJSON(data []byte) error {
	var aux map[string]json.RawMessage
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for name, field := range aux {
		if err := json.Unmarshal(field, dr.Fields[name]); err != nil {
			return err
		}
	}
	// Prune the process table of "dangling" processes which were references
	// but never specified.
	for _, proc := range dr.DiscoveryResult.Processes {
		if proc.PPID == 0 && proc.Name == "" && len(proc.Cmdline) == 0 && proc.Starttime == 0 {
			delete(dr.DiscoveryResult.Processes, proc.PID)
		}
	}
	return nil
}
