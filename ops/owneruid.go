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

//go:build linux

package ops

import (
	"errors"
	"unsafe"

	"github.com/thediveo/lxkns/nsioctl"
	"github.com/thediveo/lxkns/ops/relations"
	"golang.org/x/sys/unix"
)

// ownerUID takes an open file descriptor which must reference a user namespace.
// It then returns the UID of the user "owning" this user namespace, or an
// error. The Relation reference is only needed in case of errors, to allow for
// returning meaningful wrapped errors.
func ownerUID(ref relations.Relation, fd int) (int, error) {
	// For the reason to use "int" to represent uid_t as the return value,
	// see: https://github.com/golang/go/issues/6495; however, we must be
	// careful with the Syscall(), giving it the correct uint32 -- even on
	// 64bit Linux. See also:
	// https://elixir.bootlin.com/linux/latest/source/include/linux/types.h#L32
	uid := ^uint32(0) - 42
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL, uintptr(fd),
		uintptr(nsioctl.NS_GET_OWNER_UID), uintptr(unsafe.Pointer(&uid))) // #nosec G103
	if errno != 0 {
		return 0, newInvalidNamespaceError(ref, errors.New(errno.Error()))
	}
	return int(uid), nil
}
