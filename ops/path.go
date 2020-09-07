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
	"golang.org/x/sys/unix"
)

// NamespacePath references a Linux-kernel namespace via a filesystem path.
type NamespacePath string

// String returns the textual representation for a namespace reference by file
// descriptor. This does contain only the file descriptor, but not the
// referenced namespace (ID), as we're here dealing with the references
// themselves.
func (nsp NamespacePath) String() string {
	return fmt.Sprintf("path %s", string(nsp))
}

// Type returns the type of the Linux-kernel namespace referenced by this file
// path. Please note that a Linux kernel version 4.11 or later is required.
func (nsp NamespacePath) Type() (species.NamespaceType, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return 0, err
	}
	defer unix.Close(fd)
	t, err := ioctl(int(fd), _NS_GET_NSTYPE)
	return species.NamespaceType(t), err
}

// ID returns the namespace ID in form of its inode number for any given
// Linux kernel namespace reference.
func (nsp NamespacePath) ID() (species.NamespaceID, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return species.NoneID, err
	}
	defer unix.Close(fd)
	return fdID(nsp, int(fd))
}

// User returns the owning user namespace of any namespace, as a NamespaceFile
// reference. For user namespaces, User() behaves identical to Parent().
//
// ℹ️ A Linux kernel version 4.9 or later is required.
func (nsp NamespacePath) User() (r.Relation, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)
	userfd, err := ioctl(fd, _NS_GET_USERNS)
	return typedNamespaceFileFromFd(nsp, userfd, species.CLONE_NEWUSER, err)
}

// Parent returns the parent namespace of a hierarchical namespaces, that is, of
// PID and user namespaces. For user namespaces, Parent() and User() behave
// identical.
//
// ℹ️ A Linux kernel version 4.9 or later is required.
func (nsp NamespacePath) Parent() (r.Relation, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)
	parentfd, err := ioctl(fd, _NS_GET_PARENT)
	return namespaceFileFromFd(nsp, parentfd, err)
}

// OwnerUID returns the user id (UID) of the user namespace referenced by this
// open file descriptor.
//
// ℹ️ A Linux kernel version 4.11 or later is required.
func (nsp NamespacePath) OwnerUID() (int, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return 0, err
	}
	defer unix.Close(fd)
	return ownerUID(nsp, fd)
}

// OpenTypedReference returns an open namespace reference, from which an
// OS-level file descriptor can be retrieved using NsFd(). OpenTypeReference is
// internally used to allow optimizing switching namespaces under the condition
// that additionally the type of namespace needs to be known at the same time.
func (nsp NamespacePath) OpenTypedReference() (r.Relation, o.ReferenceCloser, error) {
	f, err := os.Open(string(nsp))
	if err != nil {
		return nil, nil, newInvalidNamespaceError(nsp, err)
	}
	openref, err := NewTypedNamespaceFile(f, 0)
	if err != nil {
		return nil, nil, newInvalidNamespaceError(nsp, err)
	}
	return openref, func() { openref.Close() }, nil
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
func (nsp NamespacePath) NsFd() (int, o.FdCloser, error) {
	var fdi int
	fdi, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return fdi, nil, newInvalidNamespaceError(nsp, err)
	}
	return int(fdi), func() { unix.Close(int(fdi)) }, nil
}

// Ensures that NamespacePath implements the Relation interface.
var _ r.Relation = (*TypedNamespacePath)(nil)

// Ensures that we've fully implemented the Opener interface.
var _ o.Opener = (*TypedNamespacePath)(nil)
