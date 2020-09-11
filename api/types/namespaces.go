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

	"github.com/thediveo/lxkns/model"
)

// NamespaceUnmarshal is the JSON serializable (digital!) twin to namespace
// objects, but just for unmarshalling: we ignore some information which might
// be present, because we need to regenerate it anyway after unmarshalling.
type NamespaceUnmarshal struct {
	ID      uint64          `json:"nsid"`
	Type    string          `json:"type"`                // "net", "user", et cetera...
	Owner   uint64          `json:"owner,omitempty"`     // namespace ID of owning user namespace
	Ref     string          `json:"reference,omitempty"` // file system path reference
	Leaders []model.PIDType `json:"leaders,omitempty"`   // list of PIDs
	Parent  uint64          `json:"parent,omitempty"`    // PID/user: namespace ID of parent namespace
	UserUID int             `json:"user-uid,omitempty"`  // user: owner's user ID (UID)
}

type NamespaceMarshal struct {
	NamespaceUnmarshal
	Children []uint64 `json:"children,omitempty"`    // PID/user: IDs of child namespaces
	Tenants  []uint64 `json:"possessions,omitempty"` // user: list of owned namespace IDs
}

type UserNamespaceMarshal struct {
	NamespaceMarshal
	UserUID int `json:"user-uid"` // enforce owner's user ID (UID)
}

func MarshalNamespace(ns model.Namespace) ([]byte, error) {
	// First set up the marshalling data used with all types of namespaces...
	aux := NamespaceMarshal{
		NamespaceUnmarshal: NamespaceUnmarshal{
			ID:      ns.ID().Ino,
			Type:    ns.Type().Name(),
			Ref:     ns.Ref(),
			Leaders: ns.LeaderPIDs(),
		},
	}
	// If we have ownership information about the owning user namespace, then
	// marshal a reference to it.
	if owner := ns.Owner(); owner != nil {
		aux.Owner = owner.(model.Namespace).ID().Ino
	}
	// Now take care of hierarchical PID and user namespaces...
	if hns, ok := ns.(model.Hierarchy); ok {
		if parent := hns.Parent(); parent != nil {
			aux.Parent = parent.(model.Namespace).ID().Ino
		}
		children := hns.Children()
		for _, child := range children {
			aux.Children = append(aux.Children, child.(model.Namespace).ID().Ino)
		}
	}
	// And now take care of what is special for user namespaces...
	if uns, ok := ns.(model.Ownership); ok {
		uaux := UserNamespaceMarshal{
			NamespaceMarshal: aux,
			UserUID:          uns.UID(),
		}
		// ...and ship it with enforced fields!
		return json.Marshal(uaux)
	}
	// Finally ship it for non-user namespaces.
	return json.Marshal(aux)
}
