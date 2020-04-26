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

package relations

import (
	"errors"
	"os"

	"github.com/thediveo/lxkns/nstypes"
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
	return nil, err
}

// Type returns the type of the Linux-kernel namespace referenced by this open
// file. Please note that a Linux kernel version 4.11 or later is required.
func (nsf NamespaceFile) Type() (nstypes.NamespaceType, error) {
	t, err := ioctl(int(nsf.Fd()), _NS_GET_NSTYPE)
	return nstypes.NamespaceType(t), err
}

// ID returns the namespace ID in form of its inode number for any given
// Linux kernel namespace reference.
func (nsf NamespaceFile) ID() (nstypes.NamespaceID, error) {
	return fdID(int(nsf.Fd()))
}

// User returns the owning user namespace of any namespace, as a NamespaceFile
// reference. For user namespaces, User() behaves identical to Parent(). A Linux
// kernel version 4.9 or later is required.
func (nsf NamespaceFile) User() (*NamespaceFile, error) {
	return namespaceFileFromFd(ioctl(int(nsf.Fd()), _NS_GET_USERNS))
}

// Parent returns the parent namespace of a hierarchical namespaces, that is, of
// PID and user namespaces. For user namespaces, Parent() and User() behave
// identical. A Linux kernel version 4.9 or later is required.
func (nsf NamespaceFile) Parent() (*NamespaceFile, error) {
	return namespaceFileFromFd(ioctl(int(nsf.Fd()), _NS_GET_PARENT))
}

// OwnerUID returns the user id (UID) of the user namespace referenced by this
// open file descriptor. A Linux kernel version 4.11 or later is required.
func (nsf NamespaceFile) OwnerUID() (int, error) {
	return ownerUID(int(nsf.Fd()))
}

// Internal convenience helper which takes a file descriptor and an error,
// returning a NamespaceFile reference if there is no error.
func namespaceFileFromFd(fd uint, err error) (*NamespaceFile, error) {
	if err != nil {
		return nil, err
	}
	if f := os.NewFile(uintptr(fd), ""); f != nil {
		return &NamespaceFile{*f}, nil
	}
	return nil, errors.New("nil namespace file descriptor")
}
