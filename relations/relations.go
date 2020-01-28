// Provides the Linux ioctl()s related to discovering namespace relationships.

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

// +build linux

package relations

import (
	"errors"
	"fmt"
	"github.com/thediveo/lxkns/nstypes"
	"os"
	"syscall"
	"unsafe"
)

/*
   Ugly IOCTL stuff.

   ATTENTION: the following definitions hold only for the "asm-generic"
   platforms, such as x86, arm, and others. Currently the only platforms
   having a different ioctl request field mapping are: alpha, mips, powerpc,
   and sparc.

   Our const definitions here come from:
   https://elixir.bootlin.com/linux/latest/source/include/uapi/asm-generic/ioctl.h
*/
const _IOC_NRBITS = 8
const _IOC_TYPEBITS = 8
const _IOC_SIZEBITS = 14

const _IOC_NRSHIFT = 0
const _IOC_TYPESHIFT = _IOC_NRSHIFT + _IOC_NRBITS
const _IOC_SIZESHIFT = _IOC_TYPESHIFT + _IOC_TYPEBITS
const _IOC_DIRSHIFT = _IOC_SIZESHIFT + _IOC_SIZEBITS

const _IOC_NONE = uint(0)

// Returns an ioctl() request value, calculated from the specific ioctl call
// properties: parameter in/out direction, type of ioctl, command number, and
// finally parameter size.
func _IOC(dir, ioctype, nr, size uint) uint {
	return (dir << _IOC_DIRSHIFT) | (ioctype << _IOC_TYPESHIFT) | (nr << _IOC_NRSHIFT) | (size << _IOC_SIZESHIFT)
}

func _IO(ioctype, nr uint) uint {
	return _IOC(_IOC_NONE, ioctype, nr, 0)
}

// Linux kernel ioctl() command for namespace relationship queries
// https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/nsfs.h
const _NSIO = 0xb7

const (
	_NS_GET_USERNS    = 0x1 // Returns a file descriptor that refers to an owning user namespace.
	_NS_GET_PARENT    = 0x2 // Returns a file descriptor that refers to a parent namespace.
	_NS_GET_NSTYPE    = 0x3 // Returns the type of namespace CLONE_NEW* value referred to by a file descriptor.
	_NS_GET_OWNER_UID = 0x4 // Get owner UID (in the caller's user namespace) for a user namespace
)

// Type returns the type of a given Linux namespace, as one of the
// CLONE_NEW... constants, or 0. The namespace can be specified (or, rather
// referenced) either giving (1) a file system path string, (2) a file
// descriptor (int), or (3) an os.File. A Linux kernel version 4.11 or later
// is required.
func Type(ref interface{}) (nstypes.NamespaceType, error) {
	//nolint: S1034
	switch ref.(type) {
	// The namespace reference given is a filesystem path (string). Thus, we
	// need to open the (pseudo) file for this path, better avoiding any go io
	// library call that can cause offloading the IO operation to another go
	// routine -- which then would be sooner or later a receipe for desaster
	// in case anyone calls this function from a different set of Linux kernel
	// namespaces.
	case string:
		fd, err := syscall.Open(ref.(string), syscall.O_RDONLY, 0)
		if err != nil {
			return 0, err
		}
		defer syscall.Close(fd)
		t, err := ioctl(fd, _NS_GET_NSTYPE)
		return nstypes.NamespaceType(t), err
	// The namespace reference is an (open) file descriptor.
	case int:
		t, err := ioctl(ref.(int), _NS_GET_NSTYPE)
		return nstypes.NamespaceType(t), err
	// The namespace reference is an open os.File.
	case *os.File:
		t, err := ioctl(int(ref.(*os.File).Fd()), _NS_GET_NSTYPE)
		return nstypes.NamespaceType(t), err
	default:
		return ^nstypes.NamespaceType(0), fmt.Errorf("Linux kernel namespace reference must be of type string, int, or *os.File, but not %T", ref)
	}
}

// ID returns the namespace ID in form of its inode number for any given
// Linux kernel namespace reference.
func ID(ref interface{}) (nstypes.NamespaceID, error) {
	var fd int
	var err error
	switch ref.(type) {
	case string: // namespace reference is a filesystem path
		fd, err = syscall.Open(ref.(string), syscall.O_RDONLY, 0)
		if err != nil {
			return 0, err
		}
		defer syscall.Close(fd)
	case int: // namespace reference is an open file descriptor
		fd = ref.(int)
	case *os.File: // namespace reference is an open file
		fd = int(ref.(*os.File).Fd())
	default:
		return 0, fmt.Errorf("Linux kernel namespace reference must be of type string, int, or ..., but not %T", ref)
	}
	var stat syscall.Stat_t
	err = syscall.Fstat(fd, &stat)
	if err != nil {
		return 0, err
	}
	return nstypes.NamespaceID(stat.Ino), nil
}

// User returns the owning user namespace of any namespace, as a file. For
// user namespaces, User() behaves identical to Parent(). A Linux kernel
// version 4.9 or later is required.
func User(ref interface{}) (*os.File, error) {
	return relationship(ref, _NS_GET_USERNS)
}

// Parent returns the parent namespace of a hierarchical namespaces, that is,
// of PID and user namespaces. For user namespaces, Parent() and User() behave
// identical. A Linux kernel version 4.9 or later is required.
func Parent(ref interface{}) (*os.File, error) {
	return relationship(ref, _NS_GET_PARENT)
}

// OwnerUID returns the user id (uid) of the specified user namespace. A Linux
// kernel version 4.11 or later is required.
func OwnerUID(ref interface{}) (int, error) {
	var fd int
	var err error
	switch ref.(type) {
	case string: // namespace reference is a filesystem path
		fd, err = syscall.Open(ref.(string), syscall.O_RDONLY, 0)
		if err != nil {
			return 0, err
		}
		defer syscall.Close(fd)
	case int: // namespace reference is an open file descriptor
		fd = ref.(int)
	case *os.File: // namespace reference is an open file
		fd = int(ref.(*os.File).Fd())
	default:
		return 0, fmt.Errorf("Linux kernel namespace reference must be of type string, int, or ..., but not %T", ref)
	}
	// For the reason to use "int" to represent uid_t as the return value,
	// see: https://github.com/golang/go/issues/6495; however, we must be
	// careful with the Syscall(), giving it the correct uint32 -- even on
	// 64bit Linux. See also:
	// https://elixir.bootlin.com/linux/latest/source/include/linux/types.h#L32
	var uid uint32 = ^uint32(0) - 42
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(_IO(_NSIO, _NS_GET_OWNER_UID)), uintptr(unsafe.Pointer(&uid)))
	if errno != 0 {
		return 0, errors.New(errno.Error())
	}
	return int(uid), nil
}

// Internal convenience function for use with those namespace ioctl's that
// given a file descriptor then return a new file descriptor on the basis of
// the specific namespace relationship requested via nr.
func relationship(ref interface{}, nr uint) (*os.File, error) {
	switch ref.(type) {
	// The namespace reference given is a filesystem path (string). Thus, we
	// need to open the (pseudo) file for this path, better avoiding any go io
	// library call that can cause offloading the IO operation to another go
	// routine -- which then would be sooner or later a receipe for desaster
	// in case anyone calls this function from a different set of Linux kernel
	// namespaces.
	case string:
		fd, err := syscall.Open(ref.(string), syscall.O_RDONLY, 0)
		if err != nil {
			return nil, err
		}
		defer syscall.Close(fd)
		return fileFromFd(ioctl(fd, nr))
	// The namespace reference is an (open) file descriptor.
	case int:
		return fileFromFd(ioctl(ref.(int), nr))
	// The namespace reference is an open os.File.
	case *os.File:
		return fileFromFd(ioctl(int(ref.(*os.File).Fd()), nr))
	default:
		return nil, fmt.Errorf("Linux kernel namespace reference must be of type string, int, or *os.File, but not %T", ref)
	}
}

// Internal convenience helper returning an os.File given only a (syscall) file
// descriptor.
func fileFromFd(fd uint, err error) (*os.File, error) {
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), ""), nil
}

// Internal convenience wrapper for calling a NSIO-related ioctl function of a
// file descriptor using only the particular NSIO command number.
func ioctl(fd int, nr uint) (uint, error) {
	nsfd, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), uintptr(_IO(_NSIO, nr)), uintptr(0))
	if errno != 0 {
		return ^uint(0), errors.New(errno.Error())
	}
	return uint(nsfd), nil
}
