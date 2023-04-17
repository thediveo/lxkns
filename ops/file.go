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

	"github.com/thediveo/ioctl"
	"github.com/thediveo/lxkns/nsioctl"
	"github.com/thediveo/lxkns/ops/internal/opener"
	"github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// NamespaceFile is an open [os.File] which references a Linux-kernel namespace.
// Please use [NewNamespaceFile] to create a *NamespaceFile from an *os.File and
// an error (correctly deals with errors by returning a nil *[NamespaceFile]).
type NamespaceFile struct {
	// Please note that we embed(!) an os.File instead of embedding an *os.File.
	// Our rationale here is that this is fine, as os.File has an indirection
	// designed in anyway in order to avoid users of os.File overwriting the
	// file descriptors. With this indirection in mind, we simply skip yet
	// another level of indirection, hopefully reducing pointer chasing. Hold my
	// beer!
	os.File
}

// NewNamespaceFile returns a new [NamespaceFile] given an *[os.File] and a nil
// error. In case of a non-nil error or of a nil *os.File it returns a nil
// *NamespaceFile instead, together with the error. This “constructor” is
// intended to be conveniently used with [os.Open] by directly accepting its two
// return values.
func NewNamespaceFile(f *os.File, err error) (*NamespaceFile, error) {
	if err == nil && f != nil {
		return &NamespaceFile{*f}, nil
	}
	return nil, newInvalidNamespaceError(nil, err)
}

// Internal convenience helper that takes a file descriptor and an error,
// returning a NamespaceFile reference if there is no error.
func namespaceFileFromFd(ref relations.Relation, fd uint, err error) (*NamespaceFile, error) {
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

// String returns the textual representation for a namespace reference by file.
// This does contain only the os.File, but not the referenced namespace (ID), as
// we're here dealing with the references themselves.
func (nsf NamespaceFile) String() (s string) {
	// We might end up with a "dummy" NamespaceFile object which isn't attached
	// to an os.File; mainly when NewNamespaceFile() gets called on the results
	// of a failed os.Open() call. Unfortunately, we cannot detect this
	// situation beforehand, but only after the crash, as there is no way to
	// detect a zero os.File (as opposed to a zero *os.File). But as this is the
	// rare exceptional code path, we can live with the penality and leave out
	// the unnecessary double file references. (Hold my beer!)
	defer func() {
		if err := recover(); err != nil {
			s = "zero os.File"
		}
	}()
	return fmt.Sprintf("os.File %v, name %q", nsf.File, nsf.Name())
}

// Type returns the type of the Linux-kernel namespace referenced by this open
// file.
//
// 🛈 A Linux kernel version 4.11 or later is required.
func (nsf NamespaceFile) Type() (species.NamespaceType, error) {
	t, err := unix.IoctlRetInt(int(nsf.Fd()), nsioctl.NS_GET_NSTYPE)
	if err != nil {
		return 0, newInvalidNamespaceError(nsf, err)
	}
	return species.NamespaceType(t), err
}

// ID returns the namespace ID in form of its inode number for any given
// Linux kernel namespace reference.
func (nsf NamespaceFile) ID() (species.NamespaceID, error) {
	return fdID(nsf, int(nsf.Fd()))
}

// User returns the owning user namespace of any namespace, as a NamespaceFile
// reference. For user namespaces, [NamespaceFile.User] mostly behaves identical
// to [NamespaceFile.Parent], except that the latter returns an untyped
// [NamespaceFile] instead of a [TypedNamespaceFile].
//
// 🛈 A Linux kernel version 4.9 or later is required.
func (nsf NamespaceFile) User() (relations.Relation, error) {
	userfd, err := ioctl.RetFd(int(nsf.Fd()), nsioctl.NS_GET_USERNS)
	// From the Linux namespace architecture, we already know that the owning
	// namespace must be a user namespace (otherwise there is something really
	// seriously broken), so we return the properly typed parent namespace
	// reference object. And we're returning an os.File-based namespace
	// reference, as this allows us to reuse the lifecycle control over the
	// newly gotten file descriptor implemented in os.File.
	return typedNamespaceFileFromFd(nsf, "NS_GET_USERNS", uint(userfd), species.CLONE_NEWUSER, err)
}

// Parent returns the parent namespace of a hierarchical namespaces, that is, of
// PID and user namespaces. For user namespaces, [NamespaceFile.Parent] and
// [NamespaceFile.User] mostly behave identical, except that the latter returns
// a [TypedNamespaceFile], while Parent returns an untyped [NamespaceFile] instead.
//
// 🛈 A Linux kernel version 4.9 or later is required.
func (nsf NamespaceFile) Parent() (relations.Relation, error) {
	fd, err := ioctl.RetFd(int(nsf.Fd()), nsioctl.NS_GET_PARENT)
	// We don't know the proper type, so return the parent namespace reference
	// as an un-typed os.File-based reference, so we can reuse the lifecycle
	// management of os.File.
	return namespaceFileFromFd(nsf, uint(fd), err)
}

// OwnerUID returns the user id (UID) of the user namespace referenced by this
// open file descriptor.
//
// 🛈 A Linux kernel version 4.11 or later is required.
func (nsf NamespaceFile) OwnerUID() (int, error) {
	return ownerUID(nsf, int(nsf.Fd()))
}

// OpenTypedReference returns an open namespace reference, from which an
// OS-level file descriptor can be retrieved using [NamespaceFile.NsFd].
//
// OpenTypeReference is also internally used to allow optimizing switching
// namespaces under the condition that additionally the type of namespace needs
// to be known at the same time.
func (nsf NamespaceFile) OpenTypedReference() (relations.Relation, opener.ReferenceCloser, error) {
	openref, err := NewTypedNamespaceFile(&nsf.File, 0)
	if err != nil {
		return nil, nil, err
	}
	return openref, func() {}, nil
}

// NsFd returns a file descriptor referencing the namespace indicated in a
// namespace reference implementing the [opener.Opener] interface.
//
// ⚠️ After the caller is done using the returned file descriptor, the caller
// must call the returned [opener.FdCloser] function in order to properly
// release process resources. In case of any error when opening the referenced
// namespace, err will be non-nil, and might additionally wrap an underlying
// OS-level error.
//
// ⚠️ The caller must make sure that the namespace reference object doesn't get
// prematurely garbage collected, while the file descriptor returned by NsFd is
// still in use.
func (nsf NamespaceFile) NsFd() (int, opener.FdCloser, error) {
	return int(nsf.Fd()), func() {}, nil
}

// Ensures that NamespaceFile implements the Relation interface.
var _ relations.Relation = (*NamespaceFile)(nil)

// Ensures that NamespaceFile also implements the Opener interface.
var _ opener.Opener = (*NamespaceFile)(nil)
