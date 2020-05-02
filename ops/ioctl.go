// Interfaces Go with the Linux ioctl()s related to discovering namespace
// relationships.

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

package ops

import (
	"errors"

	"golang.org/x/sys/unix"
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

// Internal convenience wrapper for calling a NSIO-related ioctl function of a
// file descriptor using only the particular NSIO command number.
func ioctl(fd int, nr uint) (uint, error) {
	nsfd, _, errno := unix.Syscall(unix.SYS_IOCTL,
		uintptr(fd), uintptr(_IO(_NSIO, nr)), uintptr(0))
	if errno != 0 {
		return ^uint(0), errors.New(errno.Error())
	}
	return uint(nsfd), nil
}
