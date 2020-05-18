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
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// NamespacePath references a Linux-kernel namespace via a filesystem path.
type NamespacePath string

// Type returns the type of the Linux-kernel namespace referenced by this open
// file descriptor. Please note that a Linux kernel version 4.11 or later is
// required.
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
	return fdID(int(fd))
}

// User returns the owning user namespace of any namespace, as a NamespaceFile
// reference. For user namespaces, User() behaves identical to Parent(). A Linux
// kernel version 4.9 or later is required.
func (nsp NamespacePath) User() (*NamespaceFile, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)
	return namespaceFileFromFd(ioctl(fd, _NS_GET_USERNS))
}

// Parent returns the parent namespace of a hierarchical namespaces, that is, of
// PID and user namespaces. For user namespaces, Parent() and User() behave
// identical. A Linux kernel version 4.9 or later is required.
func (nsp NamespacePath) Parent() (*NamespaceFile, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)
	return namespaceFileFromFd(ioctl(fd, _NS_GET_PARENT))
}

// OwnerUID returns the user id (UID) of the user namespace referenced by this
// open file descriptor. A Linux kernel version 4.11 or later is required.
func (nsp NamespacePath) OwnerUID() (int, error) {
	fd, err := unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return 0, err
	}
	defer unix.Close(fd)
	return ownerUID(fd)
}

// Ensures that NamespacePath implements the Relation interface.
var _ Relation = (*NamespacePath)(nil)

// Reference returns an open file descriptor which references the namespace.
// After the file descriptor is no longer needed, the caller must call the
// returned close function, in order to avoid wasting file descriptors.
func (nsp NamespacePath) Reference() (fd int, closer CloseFunc, err error) {
	var fdi int
	fdi, err = unix.Open(string(nsp), unix.O_RDONLY, 0)
	if err != nil {
		return fdi, nil, err
	}
	return int(fdi), func() { unix.Close(int(fdi)) }, nil
}

var _ Referrer = (*NamespacePath)(nil)
