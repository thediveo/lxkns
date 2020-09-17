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

// DiscoveryResult is the (digital) twin of an lxkns DiscoveryResult, which can
// be marshalled and unmarshalled to and from JSON.
type DiscoveryResult lxkns.DiscoveryResult

// NewDiscoveryResult returns a discovery result object ready for marshalling
// JSON into it or marshalling an existing lxkns discovery result.
func NewDiscoveryResult(discoveryresult *lxkns.DiscoveryResult) *DiscoveryResult {
	if discoveryresult == nil {
		return (*DiscoveryResult)(&lxkns.DiscoveryResult{
			Namespaces: *model.NewAllNamespaces(),
			Processes:  model.ProcessTable{},
		})
	}
	return (*DiscoveryResult)(discoveryresult)
}

// MarshalJSON emits the results of a discovery as JSON.
func (d *DiscoveryResult) MarshalJSON() ([]byte, error) {
	nsdict := &NamespacesDict{
		AllNamespaces: &d.Namespaces,
		ProcessTable: ProcessTable{
			ProcessTable: d.Processes,
			Namespaces:   nil,
		},
	}
	nsdict.ProcessTable.Namespaces = nsdict
	aux := struct {
		Options    *DiscoveryOptions `json:"discovery-options"`
		Namespaces *NamespacesDict   `json:"namespaces"`
		Processes  *ProcessTable     `json:"processes"`
	}{
		Options:    (*DiscoveryOptions)(&d.Options),
		Namespaces: nsdict,
		Processes:  &nsdict.ProcessTable,
	}
	return json.Marshal(aux)
}

// UnmarshalJSON unmarshals discovery results from JSON into a DiscoveryResult
// object, usually obtained with NewDiscoveryResult() first.
func (d *DiscoveryResult) UnmarshalJSON(data []byte) error {
	nsdict := &NamespacesDict{
		AllNamespaces: &d.Namespaces,
		ProcessTable: ProcessTable{
			ProcessTable: d.Processes,
			Namespaces:   nil,
		},
	}
	nsdict.ProcessTable.Namespaces = nsdict
	aux := struct {
		Options    *DiscoveryOptions `json:"discovery-options"`
		Namespaces *NamespacesDict   `json:"namespaces"`
		Processes  *ProcessTable     `json:"processes"`
	}{
		Options:    (*DiscoveryOptions)(&d.Options),
		Namespaces: nsdict,
		Processes:  &nsdict.ProcessTable,
	}
	return json.Unmarshal(data, &aux)
}
