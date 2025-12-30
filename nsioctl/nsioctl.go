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

package nsioctl

import "github.com/thediveo/ioctl"

// Linux kernel [ioctl(2)] command for [namespace relationship queries].
//
// [ioctl(2)]: https://man7.org/linux/man-pages/man2/ioctl.2.html
// [namespace relationship queries]: https://elixir.bootlin.com/linux/v6.2.11/source/include/uapi/linux/nsfs.h
const _NSIO = 0xb7

var (
	// NS_GET_USERNS returns a file descriptor that refers to an owning user
	// namespace.
	NS_GET_USERNS = ioctl.IO(_NSIO, 0x1)
	// NS_GET_PARENT returns a file descriptor that refers to a parent
	// namespace.
	NS_GET_PARENT = ioctl.IO(_NSIO, 0x2)
	// NS_GET_NSTYPE returns the type of namespace CLONE_NEW* value referred to
	// by a file descriptor.
	NS_GET_NSTYPE = ioctl.IO(_NSIO, 0x3)
	// NS_GET_OWNER_UID gets the owner UID (in the caller's user namespace) for
	// a user namespace.
	NS_GET_OWNER_UID = ioctl.IO(_NSIO, 0x4)

	// TUNGETDEVNETNS returns a file descriptor that refers the the network
	// namespace of the TAP/TUN netdev.
	TUNGETDEVNETNS = ioctl.IO('T', 227)
)
