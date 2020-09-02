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

package ops

import (
	"fmt"

	"github.com/thediveo/lxkns/species"
)

// TypedNamespacePath is an explicitly typed NamespacePath reference in the file
// system. Use this type only in case you (1) need to use Visit(), AND (2) must
// support kernels pre-4.11 which lack support for the NS_GET_NSTYPE ioctl(),
// AND (3) you know already the specific type of namespace. Please note that
// User() and Parent() require a least a 4.9+ kernel. OwnerUID() requires at
// least a 4.11+ kernel.
type TypedNamespacePath struct {
	NamespacePath
	NamespaceType species.NamespaceType
}

// NewTypedNamespacePath returns a new explicitly typed namespace path reference.
func NewTypedNamespacePath(path string, typ species.NamespaceType) *TypedNamespacePath {
	return &TypedNamespacePath{NamespacePath(path), typ}
}

// String returns the textual representation for a namespace reference by file
// descriptor. This does contain only the file descriptor, but not the
// referenced namespace (ID), as we're here dealing with the references
// themselves.
func (nsp TypedNamespacePath) String() string {
	return fmt.Sprintf("path %s, type %s",
		string(nsp.NamespacePath), nsp.NamespaceType.Name())
}

// Type returns the (explicitly given) type of the Linux-kernel namespace
// referenced.
func (nsp TypedNamespacePath) Type() (species.NamespaceType, error) {
	return nsp.NamespaceType, nil
}

// Ensures that NamespacePath implements the Relation and Referrer interfaces.
var _ Relation = (*TypedNamespacePath)(nil)
var _ Referrer = (*TypedNamespacePath)(nil)
