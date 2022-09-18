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

	"github.com/thediveo/lxkns/ops/internal/opener"
	"github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
)

// TypedNamespaceFd references a Linux-kernel namespace via an open file
// descriptor.
type TypedNamespaceFd struct {
	NamespaceFd                       // underlying file descriptor referencing the namespace.
	nstype      species.NamespaceType // type of namespace.
}

// NewTypedNamespaceFd wraps a OS-level file descriptor referencing a
// Linux-kernel namespace, as well as the type of namespace. NewTypedNamespaceFd
// can be used in those situations where the type of namespace is already known,
// where later access to the type is of reference required and the namespace
// type query ioctl() is to be avoided (such as to support 4.9 to pre-4.11 Linux
// kernels).
func NewTypedNamespaceFd(fd int, nstype species.NamespaceType) (*TypedNamespaceFd, error) {
	switch nstype {
	case species.CLONE_NEWCGROUP,
		species.CLONE_NEWIPC,
		species.CLONE_NEWNET,
		species.CLONE_NEWNS,
		species.CLONE_NEWPID,
		species.CLONE_NEWTIME,
		species.CLONE_NEWUSER,
		species.CLONE_NEWUTS:
		return &TypedNamespaceFd{
			NamespaceFd: NamespaceFd(fd),
			nstype:      nstype,
		}, nil
	}
	return nil, fmt.Errorf("invalid namespace type %x", nstype)
}

// String returns the textual representation for a typed namespace reference by
// file descriptor. This does contain only the type as well as the file
// descriptor, but not the referenced namespace (ID), as we're here dealing with
// the references themselves. If a dedicated reference was given at creation
// time (such as a filesystem path), then this is used instead of the fd number.
func (nsfd TypedNamespaceFd) String() string {
	return fmt.Sprintf("fd %d (type %s)", int(nsfd.NamespaceFd), nsfd.nstype.Name())
}

// Type returns the foreknown type of the Linux-kernel namespace set when this
// namespace reference was created. This avoids having to call the corresponding
// namespace-type syscall, so it will work also on Linux kernels before
// 4.11, offering limited backward supported in those situations where the type
// of namespace is already known when establishing the namespace reference.
func (nsfd TypedNamespaceFd) Type() (species.NamespaceType, error) {
	return nsfd.nstype, nil
}

// OpenTypedReference returns an open namespace reference, from which an
// OS-level file descriptor can be retrieved using [TypedNamespaceFd.NsFd].
//
// OpenTypeReference is also internally used to allow optimizing switching
// namespaces under the condition that additionally the type of namespace needs
// to be known at the same time.
func (nsfd *TypedNamespaceFd) OpenTypedReference() (relations.Relation, opener.ReferenceCloser, error) {
	return nsfd, func() {}, nil
}

// Ensures that TypedNamespaceFd implements the Relation interface.
var _ relations.Relation = (*TypedNamespaceFd)(nil)

// Ensures that we've fully implemented the Opener interface.
var _ opener.Opener = (*TypedNamespaceFd)(nil)
