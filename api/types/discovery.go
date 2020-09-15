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

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/model"
)

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
		Namespaces *NamespacesDict `json:"namespaces"`
		Processes  *ProcessTable   `json:"processes"`
	}{
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
		Namespaces *NamespacesDict `json:"namespaces"`
		Processes  *ProcessTable   `json:"processes"`
	}{
		Namespaces: nsdict,
		Processes:  &nsdict.ProcessTable,
	}
	return json.Unmarshal(data, &aux)
}
