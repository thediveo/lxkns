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

	o "github.com/thediveo/lxkns/ops/internal/opener"
	r "github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// TypedNamespacePath is an explicitly typed NamespacePath reference in the file
// system. Use this type in case you (1) need to use Visit(), AND (2) must
// support kernels pre-4.11 which lack support for the NS_GET_NSTYPE ioctl(),
// AND (3) you know already the specific type of namespace. You may also use
// TypedNamespacePath when using Visit() on newer kernels to slightly optimize
// things, but this isn't strictly necessary.
//
// ℹ️ Please note that User() and Parent() require a least a 4.9+ kernel.
// OwnerUID() requires at least a 4.11+ kernel.
type TypedNamespacePath struct {
	NamespacePath
	nstype species.NamespaceType
}

// NewTypedNamespacePath returns a new explicitly typed namespace path reference.
func NewTypedNamespacePath(path string, nstype species.NamespaceType) *TypedNamespacePath {
	return &TypedNamespacePath{NamespacePath(path), nstype}
}

// String returns the textual representation for a namespace reference by file
// descriptor. This does contain only the file descriptor, but not the
// referenced namespace (ID), as we're here dealing with the references
// themselves.
func (nsp TypedNamespacePath) String() string {
	return fmt.Sprintf("path %s, type %s",
		string(nsp.NamespacePath), nsp.nstype.Name())
}

// Type returns the foreknown type of the Linux-kernel namespace set when this
// namespace reference was created. This avoids having to call the corresponding
// namespace-type syscall, so it will work also on Linux kernels before 4.11,
// offering limited backward supported in those situations where the type of
// namespace is already known when establishing the namespace reference.
func (nsp TypedNamespacePath) Type() (species.NamespaceType, error) {
	return nsp.nstype, nil
}

// Parent returns the parent namespace of a hierarchical namespaces, that is, of
// PID and user namespaces. For user namespaces, Parent() and User() behave
// identical.
//
// ℹ️ A Linux kernel version 4.9 or later is required.
func (nsp TypedNamespacePath) Parent() (r.Relation, error) {
	fd, err := unix.Open(string(nsp.NamespacePath), unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)
	parentfd, err := ioctl(fd, _NS_GET_PARENT)
	// We already know what type the parent must be, so return the properly
	// typed parent namespace reference object.
	return typedNamespaceFileFromFd(nsp, parentfd, nsp.nstype, err)
}

// Ensures that TypedNamespacePath implements the Relation interface.
var _ r.Relation = (*TypedNamespacePath)(nil)

// Ensures that we've fully implemented the Opener interface.
var _ o.Opener = (*TypedNamespacePath)(nil)
