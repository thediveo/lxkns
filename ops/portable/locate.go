// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package portable

import (
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// LocateNamespace tries to recover the path reference to a “lost” namespace
// based on its ID and type by running a slightly reduced discovery and
// searching the namestack for the needle (or so). Returns a [model.Namespace]
// if successful, otherwise nil. Please see also L[ocateNamespaceInNamespaces]
// in case a discovery result is ready at hand, thus avoiding the need for an
// additional namespace discovery.
func LocateNamespace(nsid species.NamespaceID, nstype species.NamespaceType) model.Namespace {
	if nsid == species.NoneID {
		return nil // bail out early, if necessary.
	}
	// For the discovery we can skip the hierarchy and ownership parts;
	// furthermore, if the type of namespace we're looking for is known, then we
	// can narrow the search accordingly. However, we always need to discover
	// mount namespaces in order to make the bind-mounts discovery work -- more
	// so when our process is containerized.
	nst := species.CLONE_NEWNS
	if nstype == 0 {
		nst = species.AllNS
	} else {
		nst |= nstype
	}
	discovery := discover.Namespaces(
		discover.FromProcs(),
		discover.FromFds(),
		discover.FromBindmounts(),
		discover.WithNamespaceTypes(nst))
	return LocateNamespaceInNamespaces(nsid, nstype, discovery.Namespaces)
}

// LocateNamespaceInNamespaces tries to recover the path reference to a “lost”
// namespace based on its ID and type, using the specified namespace information
// (map).
func LocateNamespaceInNamespaces(nsid species.NamespaceID, nstype species.NamespaceType, allnamespaces model.AllNamespaces) model.Namespace {
	if nstype == 0 {
		nstype = species.AllNS
	}
	// Try to find the namespace by ID; if we have a specific type, then we need
	// to look only into the corresponding namespace map. Otherwise, we need to
	// check all namespace maps for all types if we can find the specified
	// namespace ID.
	if nstype != species.AllNS {
		return allnamespaces[model.TypeIndex(nstype)][nsid]
	}
	for _, nsmap := range allnamespaces {
		if ns, ok := nsmap[nsid]; ok {
			return ns
		}
	}
	return nil
}
