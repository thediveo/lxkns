// userNamespace implements the Ownership interface of user namespaces.

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

package lxkns

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/thediveo/lxkns/ops"
)

// userNamespace stores ownership information in addition to the information
// for hierarchical namespaces. On top of the interfaces supported by a
// hierarchicalNamespace, userNamespace implements the Ownership interface.
type userNamespace struct {
	hierarchicalNamespace
	owneruid int
	ownedns  AllNamespaces
}

var _ Ownership = (*userNamespace)(nil)

func (uns *userNamespace) UID() int               { return uns.owneruid }
func (uns *userNamespace) Ownings() AllNamespaces { return uns.ownedns }

// String describes this instance of a user namespace, with its parent,
// children, and owned namespaces. This description is non-recursive.
func (uns *userNamespace) String() string {
	u, err := user.LookupId(fmt.Sprintf("%d", uns.owneruid))
	var userstr string
	if err == nil {
		userstr = fmt.Sprintf(" (%q)", u.Username)
	}
	owneds := ""
	var o []string
	for _, ownedbytype := range uns.ownedns {
		for _, owned := range ownedbytype {
			o = append(o, owned.(NamespaceStringer).TypeIDString())
		}
	}
	if len(o) != 0 {
		owneds = ", owning [" + strings.Join(o, ", ") + "]"
	}
	parentandchildren := uns.ParentChildrenString()
	leaders := uns.LeaderString()
	if leaders != "" {
		leaders = ", " + leaders
	}
	return fmt.Sprintf("%s, created by UID %d%s%s, %s%s",
		uns.TypeIDString(),
		uns.owneruid, userstr,
		leaders,
		parentandchildren,
		owneds)
}

// detectUIDs takes an open file referencing a user namespace to query its
// owner's UID and then stores it for this user namespace proxy.
func (uns *userNamespace) detectUID(nsf *ops.NamespaceFile) {
	uns.owneruid, _ = nsf.OwnerUID()
}

// ResolveOwner sets the owning user namespace reference based on the owning
// user namespace id discovered earlier. Yes, we're repeating us ourselves with
// this method, because Golang is self-inflicted pain when trying to emulate
// inheritance using embedding ... note: it doesn't work correctly. The reason
// is that we need the use the correct instance pointer and not a pointer to an
// embedded instance when setting the "owned" relationship.
func (uns *userNamespace) ResolveOwner(usernsmap NamespaceMap) {
	uns.resolveOwner(uns, usernsmap)
}
