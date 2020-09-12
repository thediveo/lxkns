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
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// NamespacesDict is a dictionary of all namespaces, basically a
// model.AllNamespaces but with the added twist of creating new preliminary
// namespace objects when looking up namespaces which we don't have seen yet.
type NamespacesDict model.AllNamespaces

// NewNamespacesDict returns a new and properly initialized NamespacesDict
// ready for use.
func NewNamespacesDict() *NamespacesDict {
	return (*NamespacesDict)(model.NewAllNamespaces())
}

// Get always(!) returns a namespace object with the given ID and type. When
// the namespace is already known, then it is returned, otherwise a new
// preliminary namespace object gets created, registered and returned instead.
// Preliminary namespace objects have their ID and type set, but everything is
// still zero, including the reference (path).
func (d NamespacesDict) Get(nsid species.NamespaceID, nstype species.NamespaceType) model.Namespace {
	nsidx := model.TypeIndex(nstype)
	ns, ok := d[nsidx][nsid]
	if !ok {
		ns = namespaces.New(nstype, nsid, "")
		d[nsidx][nsid] = ns
	}
	return ns
}
