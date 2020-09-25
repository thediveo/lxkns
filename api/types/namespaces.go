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
	"fmt"
	"os/user"
	"strconv"

	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// NamespacesDict is a dictionary of all namespaces, basically a
// model.AllNamespaces but with the added twist of creating new preliminary
// namespace objects when looking up namespaces which we don't have seen yet.
type NamespacesDict struct {
	*model.AllNamespaces
	ProcessTable // our enhanced process table ;)
}

// NewNamespacesDict returns a new and properly initialized NamespacesDict ready
// for use. It will be empty if nil discovery results are specified; otherwise,
// the information from the discovery results will be used by this namespace
// dictionary.
func NewNamespacesDict(discoveryresults *lxkns.DiscoveryResult) *NamespacesDict {
	var d *NamespacesDict
	if discoveryresults == nil {
		d = &NamespacesDict{
			AllNamespaces: model.NewAllNamespaces(),
			ProcessTable:  ProcessTable{model.ProcessTable{}, nil},
		}
	} else {
		d = &NamespacesDict{
			AllNamespaces: &discoveryresults.Namespaces,
			ProcessTable:  ProcessTable{discoveryresults.Processes, nil},
		}
	}
	d.ProcessTable.Namespaces = d
	return d
}

// Get always(!) returns a Namespace interface (to a namespace object) with
// the given ID and type. When the namespace is already known, then it is
// returned, otherwise a new preliminary namespace object gets created,
// registered, and returned instead. Preliminary namespace objects have their
// ID and type set, but everything is still zero, including the reference
// (path).
func (d NamespacesDict) Get(nsid species.NamespaceID, nstype species.NamespaceType) model.Namespace {
	nsidx := model.TypeIndex(nstype)
	ns, ok := d.AllNamespaces[nsidx][nsid]
	if !ok {
		ns = namespaces.New(nstype, nsid, "")
		d.AllNamespaces[nsidx][nsid] = ns
	}
	return ns
}

// MarshalJSON emits a Linux-kernel namespace dictionary as JSON, with details
// about the individual namespaces.
func (d *NamespacesDict) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('{')
	first := true
	for _, bunchofns := range d.AllNamespaces {
		for _, ns := range bunchofns {
			if first {
				first = false
			} else {
				b.WriteRune(',')
			}
			b.WriteRune('"')
			b.WriteString(strconv.FormatUint(ns.ID().Ino, 10))
			b.WriteString(`":`)
			nsjson, err := d.MarshalNamespace(ns)
			if err != nil {
				return nil, err
			}
			b.Write(nsjson)
		}
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

// UnmarshalJSON unmarshals a Linux-kernel namespace dictionary from its JSON
// textual representation.
func (d *NamespacesDict) UnmarshalJSON(data []byte) error {
	aux := map[uint64]json.RawMessage{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	for _, rawns := range aux {
		if _, err := d.UnmarshalNamespace(rawns); err != nil {
			return err
		}
	}
	return nil
}

// NamespaceUnmarshal is the JSON serializable (digital!) twin to namespace
// objects, yet just for unmarshalling: the rationale to differentiate between
// marshalling and unmarshalling is that on unmarshalling we ignore some
// information which might be present, as we need to regenerate it anyway
// after unmarshalling (such as the list of children and the owned
// namespaces).
type NamespaceUnmarshal struct {
	ID       uint64          `json:"nsid"`                // namespace ID.
	Type     string          `json:"type"`                // "net", "user", et cetera...
	Owner    uint64          `json:"owner,omitempty"`     // namespace ID of owning user namespace.
	Ref      string          `json:"reference,omitempty"` // file system path reference.
	Leaders  []model.PIDType `json:"leaders,omitempty"`   // list of leader PIDs.
	Parent   uint64          `json:"parent,omitempty"`    // PID/user: namespace ID of parent namespace.
	UserUID  int             `json:"user-id,omitempty"`   // user: owner's user ID (UID).
	UserName string          `json:"user-name,omitempty"` // user: name.
}

// NamespaceMarshal adds those fields to NamespaceUnmarshal we marshal as a
// convenience to some JSON consumers, but which we rather prefer to ignore on
// unmarshalling.
type NamespaceMarshal struct {
	NamespaceUnmarshal
	Ealdorman model.PIDType `json:"ealdorman,omitempty"`   // PID of most senior leader process
	Children  []uint64      `json:"children,omitempty"`    // PID/user: IDs of child namespaces
	Tenants   []uint64      `json:"possessions,omitempty"` // user: list of owned namespace IDs
}

// MarshalNamespace emits a Namespace in JSON textual format.
func (d NamespacesDict) MarshalNamespace(ns model.Namespace) ([]byte, error) {
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
	// For convenience, throw in the ealdoman's PID, if available...
	if ealdorman := ns.Ealdorman(); ealdorman != nil {
		aux.Ealdorman = ealdorman.PID
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
	username := ""
	if user, err := user.LookupId(strconv.Itoa(uns.UID())); err == nil {
		username = user.Name
	}
	return json.Marshal(&struct {
		NamespaceMarshal
		UserUID  int    `json:"user-id"`   // enforce owner's user ID (UID)
		UserName string `json:"user-name"` // enforce owner's user name
	}{
		NamespaceMarshal: aux,
		UserUID:          uns.UID(),
		UserName:         username,
	})
}

// UnmarshalNamespace retrieves a Namespace from the given textual
// representation, making use of the additionally specified namespace dictionary
// to resolve references to other namespaces (if needed, by creating preliminary
// namespace objects so they can be referenced in advance). Moreover, it also
// resolves references to (leader) processes, priming the process table if
// necessary with preliminary process objects.
func (d NamespacesDict) UnmarshalNamespace(data []byte) (model.Namespace, error) {
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
	ns := d.Get(nsid, nstype)
	cns := ns.(namespaces.NamespaceConfigurer)
	cns.SetRef(aux.Ref)
	// Resolve leader process references, if a process table was given...
	if len(aux.Leaders) != 0 {
		for _, pid := range aux.Leaders {
			cns.AddLeader(d.ProcessTable.Get(pid))
		}
	}
	// Resolve the reference to the owning user namespace, if any...
	if aux.Owner != 0 {
		ownernsid := species.NamespaceIDfromInode(aux.Owner)
		// Working around my own internal API for configuring Namespaces, it
		// seems...
		_ = d.Get(ownernsid, species.CLONE_NEWUSER)
		cns.SetOwner(ownernsid)
		// This will correctly set references from our namespace <--> user
		// namespace in both(!) directions; thus we're ignoring the list of
		// owned namespaces for user namespaces during unmarshalling.
		cns.ResolveOwner(d.AllNamespaces[model.UserNS])
	}
	// Resolve the reference to the parent namespace, if any...
	if hns, ok := ns.(model.Hierarchy); ok && aux.Parent != 0 {
		parentns := d.Get(species.NamespaceIDfromInode(aux.Parent), nstype)
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
