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
// file descriptor. Please note that a Linux kernel version 4.11 or later is
// required.
func (nsfd NamespaceFd) Type() (species.NamespaceType, error) {
	t, err := ioctl(int(nsfd), _NS_GET_NSTYPE)
	if err != nil {
		err = newInvalidNamespaceError(nsfd, err)
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
// NamespaceFile reference. For user namespaces, User() behaves identical to
// Parent(). A Linux kernel version 4.9 or later is required.
func (nsfd NamespaceFd) User() (*NamespaceFile, error) {
	fd, err := ioctl(int(nsfd), _NS_GET_USERNS)
	return namespaceFileFromFd(nsfd, fd, err)
}

// Parent returns the parent namespace of the Linux-kernel namespace referenced
// by this open file descriptor. The namespace references must be either of type
// PID or user. For user namespaces, Parent() and User() behave identical. A
// Linux kernel version 4.9 or later is required.
func (nsfd NamespaceFd) Parent() (*NamespaceFile, error) {
	fd, err := ioctl(int(nsfd), _NS_GET_USERNS)
	return namespaceFileFromFd(nsfd, fd, err)
}

// OwnerUID returns the user id (UID) of the user namespace referenced by this
// open file descriptor. A Linux kernel version 4.11 or later is required.
func (nsfd NamespaceFd) OwnerUID() (int, error) {
	return ownerUID(nsfd, int(nsfd))
}

// fdID stats the given file descriptor in order to get the dev and inode
// numbers, and returns it as a NamespaceID. This is an internal convenience
// function to avoid duplicate code and is used also by the NamespaceFile and
// NamespacePath reference types.
func fdID(ref Relation, fd int) (species.NamespaceID, error) {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return species.NoneID, newInvalidNamespaceError(ref, err)
	}
	return species.NamespaceID{Dev: stat.Dev, Ino: stat.Ino}, nil
}

// Ensures that NamespaceFd implements the Relation interface.
var _ Relation = (*NamespaceFd)(nil)

// Reference returns an open file descriptor which references the namespace.
// After the file descriptor is no longer needed, the caller must call the
// returned close function, in order to avoid wasting file descriptors.
//
// Please note that in case of a NamespaceFd reference, this returns the
// original open file descriptor (and doesn't make a copy of it). Aliasing a
// file descriptor into a NamespaceFd does not take ownership, so control of the
// lifetime of the aliased file descriptor is still up to its original creator.
// In consequence, the closer returned for a namespace file descriptor will
// leave original file descriptor untouched.
func (nsfd NamespaceFd) Reference() (fd int, closer CloseFunc, err error) {
	return int(nsfd), func() {}, nil
}

// Make sure that we've fully implemented the Referrer interface.
var _ Referrer = (*NamespaceFd)(nil)
