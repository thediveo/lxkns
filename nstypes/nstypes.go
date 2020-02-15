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

// +build linux

package nstypes

import (
	"strconv"
	"strings"
)

// Unfortunately, Go's syscall package for whatever reason lacks the const
// definition for CLONE_NEWCGROUP. So we need to roll our own definitions
// anyway.

// Following are Linux namespace type constants; these are used with several
// of the namespace-related functions, such as clone() in particular, but also
// setns(), unshare(), and the NS_GET_NSTYPE ioctl(). The origin of our const
// definitions is:
// https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/sched.h

// Oh, forgo golint with its helicopter parents patronizing about how names of
// Linux kernel definitions have to look like. Go for something grown up, such
// as golangci-lint, et cetera.

// NamespaceID represents a Linux kernel namespace identifier. While namespace
// identifiers currently use only 32bit values, we're playing safe here and
// keep with the 64bit-ness of inode numbers, as which they originally appear.
type NamespaceID uint64

// NoneID is a convenience constant for signalling an invalid or non-existing
// namespace identifier.
const NoneID NamespaceID = 0

// String returns the namespace identifier as text, or "NoneID", if it is
// invalid.
func (nsid NamespaceID) String() string {
	if nsid != NoneID {
		return "NamespaceID(" + strconv.FormatInt(int64(nsid), 10) + ")"
	}
	return "NoneID"
}

// NamespaceType mirrors the data type used in the Linux kernel for the
// namespace type constants. These constants are actually part of the clone()
// syscall options parameter.
type NamespaceType uint64

// The 7 type of Linux namespaces defined at this time.
const (
	CLONE_NEWNS     NamespaceType = 0x00020000 // identifies Linux mount namespaces.
	CLONE_NEWCGROUP NamespaceType = 0x02000000 // identifies Linux cgroup namespaces.
	CLONE_NEWUTS    NamespaceType = 0x04000000 // identifies Linux UTS (*nix timesharing system) namespaces.
	CLONE_NEWIPC    NamespaceType = 0x08000000 // identifies Linux inter-process communication namespaces.
	CLONE_NEWUSER   NamespaceType = 0x10000000 // identifies Linux user namespaces.
	CLONE_NEWPID    NamespaceType = 0x20000000 // identifies Linux PID namespaces.
	CLONE_NEWNET    NamespaceType = 0x40000000 // identifies Linux network namespaces.
)

// NaNS identifies an invalid namespace type.
const NaNS NamespaceType = 0

// Name returns the type name string (such as "mnt", "net", ...) of a
// namespace type value.
func (nstype NamespaceType) Name() string {
	name, _ := typeNames[nstype]
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
}

// NameToType returns the namespace type value (constant CLONE_NEWNS, ...)
// corresponding to the specified namespace type name (such as "mnt", "net",
// et cetera).
func NameToType(name string) NamespaceType {
	t, _ := nameTypes[name]
	return t
}

// Maps Linux namespace type names (as used in the proc filesystem) to their
// Linux kernel constants.
var nameTypes = map[string]NamespaceType{
	"mnt":    CLONE_NEWNS,
	"cgroup": CLONE_NEWCGROUP,
	"uts":    CLONE_NEWUTS,
	"ipc":    CLONE_NEWIPC,
	"user":   CLONE_NEWUSER,
	"pid":    CLONE_NEWPID,
	"net":    CLONE_NEWNET,
}

// IDwithType takes a string representation of a namespace instance, such as
// "net:[1234]", and returns the ID together with the type of the namespace.
// In case the string is malformed or contains an unknown namespace type, IDwithType
// returns (NoneID, NaNS).
func IDwithType(s string) (id NamespaceID, t NamespaceType) {
	colon := strings.IndexRune(s, ':')
	if colon < 3 || s[colon+1] != '[' || s[len(s)-1] != ']' {
		return
	}
	t, ok := nameTypes[s[0:colon]]
	if !ok {
		return
	}
	value, err := strconv.ParseUint(s[colon+2 : len(s)-1], 10, 64)
	if err != nil || value <= 0 {
		return 0, 0
	}
	return NamespaceID(value), t
}
