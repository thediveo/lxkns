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
	"golang.org/x/sys/unix"
)

// NamespaceFd references a Linux-kernel namespace via an open file descriptor.
// Following Unix tradition for file descriptors, NamespaceFd is an alias for an
// int (and not an uintptr, as in some cross-platform parts of the Golang
// packages). Please note that a NamespaceFd reference aliases a file
// descriptor, but it does not take ownership of it.
type NamespaceFd int

// String returns the textual representation for a namespace reference by file
// descriptor. This does contain only the file descriptor, but not the
// referenced namespace (ID), as we're here dealing with the references
// themselves.
func (nsfd NamespaceFd) String() string {
	return fmt.Sprintf("fd %d", int(nsfd))
}

// Type returns the type of the Linux-kernel namespace referenced by this open
// file descriptor.
//
// 🛈 A Linux kernel version 4.11 or later is required.
func (nsfd NamespaceFd) Type() (species.NamespaceType, error) {
	t, err := ioctl(int(nsfd), _NS_GET_NSTYPE)
	if err != nil {
		return 0, newInvalidNamespaceError(nsfd, err)
	}
	return species.NamespaceType(t), err
}

// ID returns the namespace ID in form of its inode number of the Linux-kernel
// namespace referenced by this open file descriptor. Please be aware that ID
// even returns an inode number if the file descriptor doesn't reference a
// namespace but instead some other open file.
func (nsfd NamespaceFd) ID() (species.NamespaceID, error) {
	return fdID(nsfd, int(nsfd))
}

// User returns the owning user namespace the namespace referenced by this open
// file descriptor. The owning user namespace is returned in form of a
// [TypedNamespaceFile] reference. For user namespaces, User (mostly) behaves
// identical to [NamespaceFd.Parent]. The only difference is that User returns a
// [TypedNamespaceFile], whereas [NamespaceFd.Parent] returns an untyped
// [NamespaceFile] reference.
//
// 🛈 A Linux kernel version 4.9 or later is required.
func (nsfd NamespaceFd) User() (relations.Relation, error) {
	userfd, err := ioctl(int(nsfd), _NS_GET_USERNS)
	// From the Linux namespace architecture, we already know that the owning
	// namespace must be a user namespace (otherwise there is something really
	// seriously broken), so we return the properly typed parent namespace
	// reference object. And we're returning an os.File-based namespace
	// reference, as this allows us to reuse the lifecycle control over the
	// newly gotten file descriptor implemented in os.File.
	return typedNamespaceFileFromFd(nsfd, "NS_GET_USERNS", userfd, species.CLONE_NEWUSER, err)
}

// Parent returns the parent namespace of the Linux-kernel namespace referenced
// by this open file descriptor. The namespace references must be either of type
// PID or user. For user namespaces, [NamespaceFd.Parent] and [NamespaceFd.User]
// mostly behave identical, except that [NamespaceFd.Parent] returns an untyped
// [NamespaceFile] reference, whereas [NamespaceFd.User] returns a
// [TypedNamespaceFile] instead.
//
// ℹ️ A Linux kernel version 4.9 or later is required.
func (nsfd NamespaceFd) Parent() (relations.Relation, error) {
	fd, err := ioctl(int(nsfd), _NS_GET_PARENT)
	// We don't know the proper type, so return the parent namespace reference
	// as an un-typed os.File-based reference, so we can reuse the lifecycle
	// management of os.File.
	return namespaceFileFromFd(nsfd, fd, err)
}

// OwnerUID returns the user id (UID) of the user namespace referenced by this
// open file descriptor.
//
// 🛈 A Linux kernel version 4.11 or later is required.
func (nsfd NamespaceFd) OwnerUID() (int, error) {
	return ownerUID(nsfd, int(nsfd))
}

// fdID stats the given file descriptor in order to get the dev and inode
// numbers, and returns it as a NamespaceID. This is an internal convenience
// function to avoid duplicate code and is used also by the NamespaceFile and
// NamespacePath reference types.
func fdID(ref relations.Relation, fd int) (species.NamespaceID, error) {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return species.NoneID, newInvalidNamespaceError(ref, err)
	}
	return species.NamespaceID{Dev: stat.Dev, Ino: stat.Ino}, nil
}

// OpenTypedReference returns an open namespace reference, from which an
// OS-level file descriptor can be retrieved using [NamespaceFd.NsFd].
// OpenTypeReference is internally used to allow optimizing switching namespaces
// under the condition that additionally the type of namespace needs to be known
// at the same time.
func (nsfd NamespaceFd) OpenTypedReference() (relations.Relation, opener.ReferenceCloser, error) {
	t, err := ioctl(int(nsfd), _NS_GET_NSTYPE)
	if err != nil {
		return nil, nil, newNamespaceOperationError(nsfd, "NS_GET_NSTYPE", err)
	}
	openref, err := NewTypedNamespaceFd(int(nsfd), species.NamespaceType(t))
	return openref, func() {}, err
}

// NsFd returns an open file descriptor which references the namespace.
// After the file descriptor is no longer needed, the caller must call the
// returned close function, in order to avoid wasting file descriptors.
//
// Please note that in case of a NamespaceFd reference, this returns the
// original open file descriptor (and doesn't make a copy of it). Aliasing a
// file descriptor into a NamespaceFd does not take ownership, so control of the
// lifetime of the aliased file descriptor is still up to its original creator.
// In consequence, the closer returned for a namespace file descriptor will
// leave the original file descriptor untouched.
func (nsfd NamespaceFd) NsFd() (fd int, closer opener.FdCloser, err error) {
	return int(nsfd), func() {}, nil
}

// Ensures that NamespaceFd implements the Relation interface.
var _ relations.Relation = (*NamespaceFd)(nil)

// Make also sure that we've fully implemented the Opener interface. Golang
// would really be great if at the same time it could ensure that we've also had
// it implemented *correctly* :p
var _ opener.Opener = (*NamespaceFd)(nil)
