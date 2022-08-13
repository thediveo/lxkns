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

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// DiscoveryOptions is the (digital) twin of an lxkns [discover.DiscoverOpts],
// which can be marshalled and unmarshalled to and from JSON. This type usually
// isn't used on its own but instead as part of un/marshalling the
// discover.DiscoverOpts type.
type DiscoveryOptions discover.DiscoverOpts

// MarshalJSON emits discovery options as JSON, handling the slightly involved
// part of marshalling the list of namespace types included in the discovery
// scan.
func (doh DiscoveryOptions) MarshalJSON() ([]byte, error) {
	aux := struct {
		*discover.DiscoverOpts
		NamespaceTypes []string `json:"scanned-namespace-types"`
	}{
		DiscoverOpts: (*discover.DiscoverOpts)(&doh),
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

// UnmarshalJSON unmarshals JSON into a [DiscoveryOptions] type. It especially
// handles the slightly involved task of unmarshalling a list of namespace types
// into a bitmap mask of CLONE_NEWxxx namespace type flags.
func (doh *DiscoveryOptions) UnmarshalJSON(data []byte) error {
	aux := struct {
		*discover.DiscoverOpts
		NamespaceTypes []string `json:"scanned-namespace-types"`
	}{
		DiscoverOpts: (*discover.DiscoverOpts)(doh),
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

// DiscoveryResult is basically the (digital) twin of an lxkns
// [discover.Result], adding marshalling and unmarshalling to and from JSON.
// Besides, DiscoveryResult acts also as an extensible discovery result wrapper
// that allows API users to freely add their own fields (with objects) to
// un/marshal additional result fields, as they see fit.
type DiscoveryResult struct {
	// Maps discovery result top-level JSON elements to their (un)marshalling
	// types (the ones that then must do the real work).
	Fields          map[string]interface{}
	DiscoveryResult *discover.Result `json:"-"`
	ContainerModel  *ContainerModel  `json:"-"`
}

// JSON object field names for the standardized discovery result parts.
const (
	FieldDiscoveryOptions = "discovery-options"
	FieldNamespaces       = "namespaces"
	FieldProcesses        = "processes"
	FieldMounts           = "mounts"
	FieldContainers       = "containers"
	FieldContainerEngines = "container-engines"
	FieldContainerGroups  = "container-groups"
)

// NewDiscoveryResult returns a discovery result object ready for unmarshalling
// JSON into it or marshalling an existing lxkns [discover.Result] result.
func NewDiscoveryResult(opts ...NewDiscoveryResultOption) *DiscoveryResult {
	// A very limited initialization only before immediately applying any
	// options specified to us.
	dr := &DiscoveryResult{Fields: map[string]interface{}{}}
	for _, opt := range opts {
		opt(dr)
	}
	// No existing discovery result specified? Then create a fresh one so that
	// we're ready for unmarshalling, hiding the slightly gory details of the
	// internal mechanics of the multiple parts of a discovery result
	// interacting with each other.
	if dr.DiscoveryResult == nil {
		dr.DiscoveryResult = &discover.Result{
			Namespaces: *model.NewAllNamespaces(),
			Processes:  model.ProcessTable{},
			Mounts:     discover.NamespacedMountPathMap{},
		}
	}
	// Wrap the discovery result options, so that they can be properly
	// un/marshalled. A plain discover.DiscoveryOption cannot be fully
	// un/marshalled.
	dr.Fields[FieldDiscoveryOptions] = (*DiscoveryOptions)(&dr.DiscoveryResult.Options)
	// Wrap the namespaces "map" (dictionary) so that the original result
	// namespace map can be properly un/marshalled.
	nsdict := NewNamespacesDict(dr.DiscoveryResult)
	dr.Fields[FieldNamespaces] = nsdict
	// And finally wrap the process table, again for proper un/marshalling.
	pt := NewProcessTable(
		WithProcessTable(dr.DiscoveryResult.Processes),
		WithNamespacesDict(nsdict))
	dr.Fields[FieldProcesses] = &pt
	// Don't forget about the (optional) mounts, if present or might be
	// expected.
	if dr.DiscoveryResult.Mounts != nil {
		m := NamespacedMountMap(dr.DiscoveryResult.Mounts)
		dr.Fields[FieldMounts] = &m
	}
	// And now for the user-space container-related things, comprising not only
	// containers, but also container groups and container engines.
	dr.ContainerModel = NewContainerModel(dr.DiscoveryResult.Containers)
	dr.Fields[FieldContainers] = &dr.ContainerModel.Containers
	dr.Fields[FieldContainerEngines] = &dr.ContainerModel.ContainerEngines
	dr.Fields[FieldContainerGroups] = &dr.ContainerModel.Groups
	// Done. Phew.
	return dr
}

// NewDiscoveryResultOption defines so-called functional options for use with
// [NewDiscoveryResult].
type NewDiscoveryResultOption func(newdiscoveryresult *DiscoveryResult)

// WithResult instructs [NewDiscoveryResult] to use an existing lxkns discovery
// result (of type [discover.Result]); this is typically used for marshalling
// only, but not needed for unmarshalling. In the latter case you probably want
// to prefer starting with a clean slate.
func WithResult(result *discover.Result) NewDiscoveryResultOption {
	return func(ndr *DiscoveryResult) {
		ndr.DiscoveryResult = result
	}
}

// WithElement allows API users to add their own top-level elements for
// un/marshalling to discovery results when using [NewDiscoveryResult]. For
// unmarshalling you need to use WithElement in order to add a non-nil zero
// value of the correct type in order to be able to unmarshal into the correct
// type instead of a generic map[string]interface{}:
//
//	// foobar is a JSON un/marshallable type of your own. For unmarshalling,
//	// allocate a non-nil zero value, which can be unmarshalled correctly.
//	discoresult := NewDiscoveryResult(
//	    WithDiscoveryResult(all),
//	    WithElement("foobar", foobar))
//	json.Marshal(discoresult)
//
// For unmarshalling:
//
//	discoresult := NewDiscoveryResult(WithElement("foobar", foobar))
//	json.Unmarshal(jsondata, discoresult)
func WithElement(name string, obj interface{}) NewDiscoveryResultOption {
	return func(ndr *DiscoveryResult) {
		ndr.Fields[name] = obj
	}
}

// Result returns the wrapped discover.DiscoveryResult.
func (dr DiscoveryResult) Result() *discover.Result {
	return dr.DiscoveryResult
}

// Processes returns the process table from the wrapped [discover.Result].
func (dr DiscoveryResult) Processes() model.ProcessTable {
	return dr.DiscoveryResult.Processes
}

// Mounts returns the namespace'd mount paths and points from the wrapped
// [discover.Result].
func (dr DiscoveryResult) Mounts() discover.NamespacedMountPathMap {
	return dr.DiscoveryResult.Mounts
}

// Get returns the user-specified result extension object for the specified
// extension field. The field must have been added first with the [WithElement]
// option when creating the un/marshalling wrapper object for discovery results.
func (dr DiscoveryResult) Get(name string) interface{} {
	return dr.Fields[name]
}

// MarshalJSON marshals discovery results into their JSON textual
// representation.
func (dr DiscoveryResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(dr.Fields)
}

// UnmarshalJSON unmarshals discovery results from JSON into a DiscoveryResult
// object, usually obtained with [NewDiscoveryResult] first.
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
	// Get the containers and put them into the underlying discovery result; the
	// containers will reference the engines and groups.
	dr.DiscoveryResult.Containers = dr.ContainerModel.Containers.ContainerSlice()

	return nil
}
