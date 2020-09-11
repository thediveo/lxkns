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
	"os"

	o "github.com/thediveo/lxkns/ops/internal/opener"
	r "github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
)

// TypedNamespaceFile is a NamespaceFile (wrapping an open os.File) with a
// foreknown namespace type, optionally to be used in those use cases where the
// type of namespace referenced is known in advance. In such cases, ioctl()
// round trips to infer the type of namespace (when required) can be avoided,
// using the foreknown type instead.
type TypedNamespaceFile struct {
	NamespaceFile
	nstype species.NamespaceType // foreknown type of Linux kernel namespace
}

// NewTypedNamespaceFile takes an open(!) os.File plus the type of namespace
// referenced and returns a new typed namespace reference object. If the
// namespace type is left zero, then this convenience helper will auto-detect
// it, unless when on a pre-4.11 kernel, where auto-detection is impossible
// due to the missing specific ioctl().
func NewTypedNamespaceFile(f *os.File, nstype species.NamespaceType) (*TypedNamespaceFile, error) {
	if f != nil && nstype == 0 {
		t, err := ioctl(int(f.Fd()), _NS_GET_NSTYPE)
		if err != nil {
			return nil, newNamespaceOperationError(&NamespaceFile{*f}, "NS_GET_NSTYPE", err)
		}
		nstype = species.NamespaceType(t)
	}
	return &TypedNamespaceFile{
		NamespaceFile: NamespaceFile{*f},
		nstype:        nstype,
	}, nil
}

// Internal convenience helper that checks that there was no error and then
// returns a new TypedNamespaceFile object wrapping the underlying OS-level file
// descriptor to a namespace. The lifetime of the file descriptor then is under
// control of the TypedNamespaceFile object, and its embedded os.File object in
// particular. If there was an error instead, then this convenience functions
// returns a suitable namespace-related error, additionally wrapping the
// underlying OS-level error.
func typedNamespaceFileFromFd(ref r.Relation, op string, fd uint, nstype species.NamespaceType, err error) (*TypedNamespaceFile, error) {
	if err != nil {
		if op != "" {
			return nil, newNamespaceOperationError(ref, op, err)
		}
		return nil, newInvalidNamespaceError(ref, err)
	}
	if f := os.NewFile(uintptr(fd), ""); f != nil {
		return &TypedNamespaceFile{
				NamespaceFile: NamespaceFile{*f},
				nstype:        nstype},
			nil
	}
	return nil, fmt.Errorf(
		"xlkns ops.TypedNamespaceFile: invalid file descriptor %d: %w",
		int(fd), newInvalidNamespaceError(ref, nil))
}

// Type returns the foreknown type of the Linux-kernel namespace set when this
// namespace reference was created. This avoids having to call the corresponding
// namespace-type syscall, so it will work also on Linux kernels before 4.11,
// offering limited backward supported in those situations where the type of
// namespace is already known when establishing the namespace reference.
func (nsf TypedNamespaceFile) Type() (species.NamespaceType, error) {
	return nsf.nstype, nil
}

// Parent returns the parent namespace of a hierarchical namespaces, that is, of
// PID and user namespaces. For user namespaces, Parent() and User() behave
// identical.
//
// ℹ️ A Linux kernel version 4.9 or later is required.
func (nsf TypedNamespaceFile) Parent() (r.Relation, error) {
	fd, err := ioctl(int(nsf.Fd()), _NS_GET_PARENT)
	// We already know what type the parent must be, so return the properly
	// typed parent namespace reference object.
	return typedNamespaceFileFromFd(nsf, "NS_GET_PARENT", fd, nsf.nstype, err)
}

// OpenTypedReference returns an open and typed namespace reference, from which
// an OS-level file descriptor can be retrieved using NsFd(). OpenTypeReference
// is internally used to allow optimizing switching namespaces under the
// condition that additionally the type of namespace needs to be known at the
// same time.
func (nsf TypedNamespaceFile) OpenTypedReference() (r.Relation, o.ReferenceCloser, error) {
	return nsf, func() {}, nil
}

// NsFd returns a file descriptor referencing the namespace indicated in a
// namespace reference implementing the Opener interface.
//
// ⚠️ After the caller is done using the returned file descriptor, the caller
// must call the returned FdCloser function in order to properly release process
// resources. In case of any error when opening the referenced namespace, err
// will be non-nil, and might additionally wrap an underlying OS-level error.
//
// ⚠️ The caller must make sure that the namespace reference object doesn't get
// prematurely garbage collected, while the file descriptor returned by NsFd()
// is still in use.
func (nsf TypedNamespaceFile) NsFd() (int, o.FdCloser, error) {
	return int(nsf.Fd()), func() {}, nil
}

// Nota bene: we still "inherit" several of the other functionality by force of
// "syntactic sugar" from Golang's compiler, so no need to rehash it here.

// Ensures that TypedNamespaceFile implements the Relation interface.
var _ r.Relation = (*TypedNamespaceFile)(nil)

// Ensures that we've fully implemented the Opener interface.
var _ o.Opener = (*TypedNamespaceFile)(nil)
