// Discovers, or rather, resolves, the ownership relations between non-user
// namespaces and user namespaces.

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

// +build linux

package lxkns

import (
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// resolveOwnership unearths which non-user namespaces are owned by which user
// namespaces. We only run the resolution phase after we've discovered a
// complete map of all user namespaces: only now we can resolve the owner
// userspace ids to their corresponding user namespace objects.
func resolveOwnership(nstype species.NamespaceType, _ string, result *DiscoveryResult) {
	if result.Options.SkipOwnership || nstype == species.CLONE_NEWUSER {
		if result.Options.SkipOwnership {
			log.Infof("skipping discovery of namespaces ownership")
		}
		return
	}
	log.Debugf("running discovery of %s namespaces ownership", nstype.Name())
	// The namespace type discovery sequence guarantees us that by the
	// time we got here, the user namespaces already have been fully
	// discovered, so we have a complete map of them.
	usernsmap := result.Namespaces[model.UserNS]
	nstypeidx := model.TypeIndex(nstype)
	nsmap := result.Namespaces[nstypeidx]
	for _, ns := range nsmap {
		ns.(namespaces.NamespaceConfigurer).ResolveOwner(usernsmap)
		if owner := ns.Owner(); owner != nil {
			log.Debugf("%s:[%d] owned by user:[%d]",
				ns.Type().Name(), ns.ID().Ino, owner.(model.Namespace).ID().Ino)
		}
	}
}
