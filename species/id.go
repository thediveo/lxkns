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

package species

import (
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// NamespaceID represents a Linux kernel namespace identifier. NamespaceIDs can
// be compared for equality or inequality using Golang's "==" and "!="
// operators. However, namespaceIDs are not ordered, so they cannot compared
// according to their order (they don't possess) using "<", et cetera.
//
// While namespace identifiers currently use only 32bit values, we're playing
// safe here and keep with the 64bit-ness of inode numbers, as which they
// originally appear. Additionally, we also adhere to
// http://man7.org/linux/man-pages/man7/namespaces.7.html and also take the
// device a namespace inode lives on into consideration, to cover for a
// potential future with multiple namespace filesystems, as opposed to the
// single "nsfs" namespace filesystem of today.
//
// Note, there are some caveats to watch for, such as that the current textual
// format used by the Linux kernel when rendering namespaces (references) as
// text does not cater for the device ID, but only a namespace's inode. The
// textual conversions in this package work around these limitations.
type NamespaceID struct {
	Dev uint64 // device ID maintaining the namespace (Golang insists on uint64)
	Ino uint64 // inode number of this namespace.
}

// NoneID is a convenience for signalling an invalid or non-existing namespace
// identifier.
var NoneID = NamespaceID{}

// String returns the namespace identifier in form of "NamespaceID(dev,#no)" as
// text, or "NoneID", if it is invalid. Please note that String on purpose does
// not use the text format used in the Linux kernel, as a namespace identifier
// has no namespace type information attached to it. Besides, not least it is
// used by Golang debuggers when rendering values, so we here follow Golang
// (tooling) convention.
func (nsid NamespaceID) String() string {
	if nsid != NoneID {
		return "NamespaceID(" + strconv.FormatUint(uint64(nsid.Dev), 10) + "," +
			strconv.FormatUint(uint64(nsid.Ino), 10) + ")"
	}
	return "NoneID"
}

// IDwithType takes a string representation of a namespace instance, such as
// "net:[1234]", and returns the ID together with the type of the namespace (but
// see note below). In case the string is malformed or contains an unknown
// namespace type, IDwithType returns (NoneID, NaNS).
//
// There is an important gotcha to be aware of: the Linux kernel only uses a
// namespace's inode number in its textual format, dropping the device ID where
// the namespace is located on. To work around this oversight and to allow for
// namespace IDs to be comparable using "==", IDwithType adds the missing device
// ID by guessing it from the net namespace of the current process at the time
// of its startup.
func IDwithType(s string) (id NamespaceID, t NamespaceType) {
	// There must be a colon, immediately followed by an opening square bracket,
	// as well as a terminating closing square bracket.
	colon := strings.IndexRune(s, ':')
	if colon < 3 || s[colon+1] != '[' || s[len(s)-1] != ']' {
		return
	}
	// Look up the type of namespaces (which goes before the ":").
	t, ok := nameTypes[s[0:colon]]
	if !ok {
		return
	}
	value, err := strconv.ParseUint(s[colon+2:len(s)-1], 10, 64)
	if err != nil {
		// As t might have been correctly set already, make sure to not leak it
		// when bailing out with an error...
		return NoneID, 0
	}
	// In the current sorry state of affairs, we need to sneak in the device ID
	// of the nsfs filesystem in order to get complete namespace identifiers.
	if dev := nsfsDev(); dev != 0 {
		return NamespaceID{Dev: dev, Ino: value}, t
	}
	return NoneID, 0
}

// NamespaceIDfromInode is an inconvenience helper that caters for the current
// chaos in that several sources of inodes, such as the kernel's own textual
// references and 3rd party CLI tools such as "lsns", only give a namespace's
// inode number, but not its device ID. It does so by glancing the missing
// device ID from one of our own process' net namespace (at the startup time)
// and then adds that in the hope that things still work correctly for the
// moment.
func NamespaceIDfromInode(ino uint64) NamespaceID {
	if dev := nsfsDev(); dev != 0 {
		return NamespaceID{Dev: dev, Ino: ino}
	}
	return NoneID
}

// In order return correct NamespaceIDs given only the kernel's currently
// incomplete textual format, we need to pick up the device ID of the kernel's
// special nsfs filesystem. nsfs manages namespace identifiers. In order to
// avoid a circual dependency, we cannot use ops.NamespacePath, but instead have
// to use the underlying query directly to get the required device ID.
var nsfsdev = ^uint64(0)

// nsfsDev returns the device ID of the nsfs filesystem, to be used to fix
// incomplete textual namespace references. This function dynamically discovers
// the device ID and then caches it. It relies on a properly mounted /proc.
func nsfsDev() uint64 {
	if nsfsdev == ^uint64(0) {
		var stat unix.Stat_t
		if err := unix.Stat("/proc/self/ns/net", &stat); err != nil {
			nsfsdev = 0
		} else {
			nsfsdev = stat.Dev
		}
	}
	return nsfsdev
}

// Prime the "cache" for the missing device ID in textual namespace IDs during
// startup.
var _ = nsfsDev()
