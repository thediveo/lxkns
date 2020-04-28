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

import "github.com/thediveo/lxkns/nstypes"

// Relation defines query operations on Linux-kernel namespaces for discovering
// relationships between namespaces as well as some properties, when only a
// namespace reference is given. Such references can be in form of a filesystem
// path, an open file descriptor, or an open file.
type Relation interface {
	// Type of the referenced namespace, such as CLONE_NEWNET, et cetera.
	// Returns an error in case of an invalid namespace reference (closed file
	// descriptor, invalid path, et cetera).
	Type() (nstypes.NamespaceType, error)

	// ID (inode number) of the referenced namespace. Returns an error in case
	// of an invalid namespace reference.
	ID() (nstypes.NamespaceID, error)

	// User namespace owning the referenced namespace. The owning user namespace
	// is returned in form of a NamespaceFile reference when there was no error
	// in retrieving the information.
	User() (*NamespaceFile, error)

	// Parent namespace of the referenced PID or user namespace. Returns an
	// error if the parent doesn't exist, if the caller hasn't capabilities in
	// the parent namespace, or if the referenced namespace is neither a PID nor
	// a user namespace.
	Parent() (*NamespaceFile, error)

	// User ID of the process originally creating the referenced user namespace.
	// Returns an error, if the referenced namespace is not a user namespace.
	OwnerUID() (int, error)
}
