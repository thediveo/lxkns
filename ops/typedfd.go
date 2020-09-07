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
)

// TypedNamespaceFd references a Linux-kernel namespace via an open file
// descriptor.
type TypedNamespaceFd struct {
	NamespaceFd                       // underlying file descriptor referencing the namespace.
	closer      o.ReferenceCloser     // optional closer for cleaning up the file descriptor.
	nstype      species.NamespaceType // type of namespace.
	oref        string                // optional original namespace reference for error reporting.
}

// NewTypedNamespaceFd FIXME: write doc
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
		return newTypedNamespaceFd(fd, nstype, nil, ""), nil
	}
	return nil, fmt.Errorf("invalid namespace type %x", nstype)
}

func newTypedNamespaceFd(fd int, nstype species.NamespaceType, closer o.ReferenceCloser, oref string) *TypedNamespaceFd {
	return &TypedNamespaceFd{
		NamespaceFd: NamespaceFd(fd),
		closer:      closer,
		nstype:      nstype,
		oref:        oref,
	}
}

// String returns the textual representation for a typed namespace reference by
// file descriptor. This does contain only the type as well as the file
// descriptor, but not the referenced namespace (ID), as we're here dealing with
// the references themselves. If a dedicated reference was given at creation
// time (such as a filesystem path), then this is used instead of the fd number.
func (nsfd TypedNamespaceFd) String() string {
	if nsfd.oref != "" {
		return fmt.Sprintf("%s (type %s)", nsfd.oref, nsfd.nstype.Name())
	}
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

// FIXME: write doc
func (nsfd TypedNamespaceFd) OpenTypedReference() (r.Relation, o.ReferenceCloser, error) {
	if nsfd.closer != nil {
		return nsfd, nsfd.closer, nil
	}
	return nsfd, func() {}, nil
}

// Ensures that TypedNamespaceFd implements the Relation interface.
var _ r.Relation = (*TypedNamespaceFd)(nil)

// Ensures that we've fully implemented the Opener interface.
var _ o.Opener = (*TypedNamespaceFd)(nil)
