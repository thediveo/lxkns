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

	"github.com/thediveo/lxkns/species"
)

// NamespaceFile is an open os.File which references a Linux-kernel namespace.
// Please use NewNamespaceFile() to create a *NamespaceFile from an *os.File and
// an error (correctly deals with errors by returning a nil *NamespaceFile).
type NamespaceFile struct {
	// Please note that we embed an os.File instead of *os.File. Our rationale here
	// is that this is fine, as os.File has an indirection designed in anyway in
	// order to avoid users of os.File overwriting the file descriptors. With this
	// indirection in mind, we simply skip yet another level of indirection,
	// hopefully reducing pointer chasing.
	os.File
}

// NewNamespaceFile returns a new NamespaceFile given an *os.File and a nil
// error. In case of a non-nil error or of a nil *os.File it returns a nil
// *NamespaceFile instead, together with the error.
func NewNamespaceFile(f *os.File, err error) (*NamespaceFile, error) {
	if err == nil && f != nil {
		return &NamespaceFile{*f}, nil
	}
	return nil, newInvalidNamespaceError(&NamespaceFile{}, err)
}

// String returns the textual representation for a namespace reference by file.
// This does contain only the os.File, but not the referenced namespace (ID), as
// we're here dealing with the references themselves.
func (nsf NamespaceFile) String() (s string) {
	defer func() {
		if err := recover(); err != nil {
			s = "nil os.File"
		}
	}()
	return fmt.Sprintf("os.File %v, name %q", nsf.File, nsf.Name())
}

// Type returns the type of the Linux-kernel namespace referenced by this open
// file. Please note that a Linux kernel version 4.11 or later is required.
func (nsf NamespaceFile) Type() (species.NamespaceType, error) {
	t, err := ioctl(int(nsf.Fd()), _NS_GET_NSTYPE)
	if err != nil {
		err = newInvalidNamespaceError(nsf, err)
	}
	return species.NamespaceType(t), err
}

// ID returns the namespace ID in form of its inode number for any given
// Linux kernel namespace reference.
func (nsf NamespaceFile) ID() (species.NamespaceID, error) {
	return fdID(nsf, int(nsf.Fd()))
}

// User returns the owning user namespace of any namespace, as a NamespaceFile
// reference. For user namespaces, User() behaves identical to Parent(). A Linux
// kernel version 4.9 or later is required.
func (nsf NamespaceFile) User() (*NamespaceFile, error) {
	fd, err := ioctl(int(nsf.Fd()), _NS_GET_USERNS)
	return namespaceFileFromFd(nsf, fd, err)
}

// Parent returns the parent namespace of a hierarchical namespaces, that is, of
// PID and user namespaces. For user namespaces, Parent() and User() behave
// identical. A Linux kernel version 4.9 or later is required.
func (nsf NamespaceFile) Parent() (*NamespaceFile, error) {
	fd, err := ioctl(int(nsf.Fd()), _NS_GET_PARENT)
	return namespaceFileFromFd(nsf, fd, err)
}

// OwnerUID returns the user id (UID) of the user namespace referenced by this
// open file descriptor. A Linux kernel version 4.11 or later is required.
func (nsf NamespaceFile) OwnerUID() (int, error) {
	return ownerUID(nsf, int(nsf.Fd()))
}

// Internal convenience helper which takes a file descriptor and an error,
// returning a NamespaceFile reference if there is no error.
func namespaceFileFromFd(ref Relation, fd uint, err error) (*NamespaceFile, error) {
	if err != nil {
		return nil, newInvalidNamespaceError(ref, err)
	}
	if f := os.NewFile(uintptr(fd), ""); f != nil {
		return &NamespaceFile{*f}, nil
	}
	return nil, fmt.Errorf(
		"all I got was a nil file wrapper for: %w",
		newInvalidNamespaceError(ref, nil))
}

// Ensures that NamespaceFile implements the Relation interface.
var _ Relation = (*NamespaceFile)(nil)

// Reference returns an open file descriptor which references the namespace.
// After the file descriptor is no longer needed, the caller must call the
// returned close function, in order to avoid wasting file descriptors.
func (nsf NamespaceFile) Reference() (fd int, closer CloseFunc, err error) {
	return int(nsf.Fd()), func() {}, nil
}

var _ Referrer = (*NamespaceFile)(nil)
