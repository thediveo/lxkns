// Definitions of data types and constants related to Linux kernel namespaces.

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
// +build linux

package species

import (
	"strconv"

	"golang.org/x/sys/unix"
)

//revive:disable:var-naming

// NamespaceType mirrors the data type used in the Linux kernel for the
// namespace type constants. These constants are actually part of the clone()
// syscall options parameter.
type NamespaceType uint64

// The 8 type of Linux namespaces defined at this time (sic!). Please note that
// the 8th namespace is only supported since Kernel 5.6+.
//
// These constants ([Linux source]) are used with several of the
// namespace-related functions, such as [clone(7)] in particular, but also
// [setns(2)], [unshare(2)], and the [NS_GET_NSTYPE ioctl(2)].
//
//   - Oh, forgo golint with its “helicopter parents” attitude patronizing us about
//     how names of Linux kernel definitions have to look like. Go for something
//     grown up, such as golangci-lint, and many more, which hide the totally
//     childish behavior of golint.
//
// [clone(7)]: https://man7.org/linux/man-pages/man2/clone.2.html
// [setns(2)]: https://man7.org/linux/man-pages/man2/setns.2.html
// [unshare(2)]: https://man7.org/linux/man-pages/man2/unshare.2.html
// [NS_GET_NSTYPE ioctl(2)]: https://man7.org/linux/man-pages/man2/ioctl_ns.2.html
// [Linux source]: https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/sched.h
const (
	CLONE_NEWNS     = NamespaceType(unix.CLONE_NEWNS)
	CLONE_NEWCGROUP = NamespaceType(unix.CLONE_NEWCGROUP)
	CLONE_NEWUTS    = NamespaceType(unix.CLONE_NEWUTS)
	CLONE_NEWIPC    = NamespaceType(unix.CLONE_NEWIPC)
	CLONE_NEWUSER   = NamespaceType(unix.CLONE_NEWUSER)
	CLONE_NEWPID    = NamespaceType(unix.CLONE_NEWPID)
	CLONE_NEWNET    = NamespaceType(unix.CLONE_NEWNET)
	CLONE_NEWTIME   = NamespaceType(unix.CLONE_NEWTIME)
)

// NaNS identifies an invalid namespace type.
const NaNS NamespaceType = 0

// AllNS is the OR-ed bitmask of all currently defined (8) Linux-kernel
// namespace type constants.
const AllNS = CLONE_NEWNS | CLONE_NEWCGROUP | CLONE_NEWUTS | CLONE_NEWIPC |
	CLONE_NEWUSER | CLONE_NEWPID | CLONE_NEWNET | CLONE_NEWTIME

// Name returns the type name string (such as "mnt", "net", ...) of a
// namespace type value.
func (nstype NamespaceType) Name() string {
	name := typeNames[nstype]
	return name
}

// String returns the Linux kernel namespace constant name for a given
// namespace type value.
func (nstype NamespaceType) String() string {
	switch nstype {
	case NaNS:
		return "NaNS"
	case CLONE_NEWNS:
		return "CLONE_NEWNS"
	case CLONE_NEWCGROUP:
		return "CLONE_NEWCGROUP"
	case CLONE_NEWUTS:
		return "CLONE_NEWUTS"
	case CLONE_NEWIPC:
		return "CLONE_NEWIPC"
	case CLONE_NEWUSER:
		return "CLONE_NEWUSER"
	case CLONE_NEWPID:
		return "CLONE_NEWPID"
	case CLONE_NEWNET:
		return "CLONE_NEWNET"
	case CLONE_NEWTIME:
		return "CLONE_NEWTIME"
	default:
		return "NamespaceType(" + strconv.FormatInt(int64(nstype), 10) + ")"
	}
}

// Maps Linux namespace constants to their "short" type names, as used in the
// proc filesystem.
var typeNames = map[NamespaceType]string{
	CLONE_NEWNS:     "mnt",
	CLONE_NEWCGROUP: "cgroup",
	CLONE_NEWUTS:    "uts",
	CLONE_NEWIPC:    "ipc",
	CLONE_NEWUSER:   "user",
	CLONE_NEWPID:    "pid",
	CLONE_NEWNET:    "net",
	CLONE_NEWTIME:   "time",
}

// NameToType returns the namespace type value (constant [CLONE_NEWNS], ...)
// corresponding to the specified namespace type name (such as "mnt", "net", et
// cetera).
func NameToType(name string) NamespaceType {
	t := nameTypes[name]
	return t
}

// Maps Linux namespace type names (as used in the proc filesystem) to their
// Linux kernel constants.
var nameTypes = map[string]NamespaceType{}

func init() {
	for species, name := range typeNames {
		nameTypes[name] = species
	}
}
