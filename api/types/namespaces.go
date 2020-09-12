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

	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// NamespaceUnmarshal is the JSON serializable (digital!) twin to namespace
// objects, yet just for unmarshalling: the rationale to differentiate between
// marshalling and unmarshalling is that on unmarshalling we ignore some
// information which might be present, as we need to regenerate it anyway
// after unmarshalling (such as the list of children and the owned
// namespaces).
type NamespaceUnmarshal struct {
	ID      uint64          `json:"nsid"`               // namespace ID.
	Type    string          `json:"type"`               // "net", "user", et cetera...
	Owner   uint64          `json:"owner,omitempty"`    // namespace ID of owning user namespace.
	Ref     string          `json:"reference"`          // file system path reference.
	Leaders []model.PIDType `json:"leaders,omitempty"`  // list of PIDs.
	Parent  uint64          `json:"parent,omitempty"`   // PID/user: namespace ID of parent namespace.
	UserUID int             `json:"user-uid,omitempty"` // user: owner's user ID (UID).
}

// NamespaceMarshal adds those fields to NamespaceUnmarshal we marshal as a
// convenience to some JSON consumers, but which we rather prefer to ignore on
// unmarshalling.
type NamespaceMarshal struct {
	NamespaceUnmarshal
	Children []uint64 `json:"children,omitempty"`    // PID/user: IDs of child namespaces
	Tenants  []uint64 `json:"possessions,omitempty"` // user: list of owned namespace IDs
}

// MarshalNamespace emits a Namespace in JSON textual format.
func MarshalNamespace(ns model.Namespace) ([]byte, error) {
	// First set up the marshalling data used with all types of namespaces,
	// albeit some fields might not be marshalled when not in use.
	aux := NamespaceMarshal{
		NamespaceUnmarshal: NamespaceUnmarshal{
			ID:      ns.ID().Ino,
			Type:    ns.Type().Name(),
			Ref:     ns.Ref(),
			Leaders: ns.LeaderPIDs(),
		},
	}
	// If we have ownership information, then marshal a reference (ID) to the
	// owning user namespace.
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
	uns, ok := ns.(model.Ownership)
	if !ok {
		return json.Marshal(aux)
	}
	// And now take care of what is special for user namespaces; such as
	// enforcing sending a user ID, even if it is 0/root (which often will be
	// the case).
	return json.Marshal(&struct {
		NamespaceMarshal
		UserUID int `json:"user-uid"` // enforce owner's user ID (UID)
	}{
		NamespaceMarshal: aux,
		UserUID:          uns.UID(),
	})
}

// Unmarshal retrieves a Namespace from the given textual representation,
// making use of the additionally specified namespace dictionary to resolve
// references to other namespaces (if needed, by creating preliminary
// namespace objects so they can be referenced in advance). Moreover, it also
// resolves references to (leader) processes, priming the process table if
// necessary with preliminary process objects.
func UnmarshalNamespace(data []byte, allns *NamespacesDict, procs model.ProcessTable) (model.Namespace, error) {
	// Unmarshal the required information so we can at least create a "bare"
	// namespace object of the correct type and with the correct ID.
	var aux NamespaceUnmarshal
	if err := json.Unmarshal(data, &aux); err != nil {
		return nil, err
	}
	nstype := species.NameToType(aux.Type)
	if nstype == 0 {
		return nil, fmt.Errorf("invalid namespace type %q", aux.Type)
	}
	nsid := species.NamespaceIDfromInode(aux.ID)
	if nsid == species.NoneID {
		return nil, fmt.Errorf("invalid namespace id %d", aux.ID)
	}
	// Either get a preliminary namespace object that had been referenced
	// before its "definition", or create a suitable namespace object right
	// now.
	ns := allns.Get(nsid, nstype)
	cns := ns.(namespaces.NamespaceConfigurer)
	cns.SetRef(aux.Ref)
	// Resolve leader process references, if a process table was given...
	if len(aux.Leaders) != 0 && procs != nil {
		// TODO: use "Get" lookup/priming API in the future
		for _, pid := range aux.Leaders {
			proc, ok := procs[pid]
			if !ok {
				proc = &model.Process{PID: pid}
				procs[pid] = proc
			}
			cns.AddLeader(proc)
		}
	}
	// Resolve the reference to the owning user namespace, if any...
	if aux.Owner != 0 {
		ownernsid := species.NamespaceIDfromInode(aux.Owner)
		// Working around my own internal API for configuring Namespaces, it
		// seems...
		_ = allns.Get(ownernsid, species.CLONE_NEWUSER)
		cns.SetOwner(ownernsid)
		// This will correctly set references from our namespace <--> user
		// namespace in both(!) directions; thus we're ignoring the list of
		// owned namespaces for user namespaces during unmarshalling.
		cns.ResolveOwner(allns[model.UserNS])
	}
	// Resolve the reference to the parent namespace, if any...
	if hns, ok := ns.(model.Hierarchy); ok && aux.Parent != 0 {
		parentns := allns.Get(species.NamespaceIDfromInode(aux.Parent), nstype)
		parentns.(namespaces.HierarchyConfigurer).AddChild(hns)
	}
	// Set the user namespace's user ID, if applicable. Please note that we
	// here do not resolve the references to the owned namespaces.
	if uns, ok := ns.(namespaces.UserConfigurer); ok {
		uns.SetOwnerUID(aux.UserUID)
	}
	// Phew, done!
	return ns, nil
}
