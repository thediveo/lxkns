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

// Package relations gives access to properties of and relationships between
// Linux-kernel namespaces, such as type and ID of a namespace, its owning user
// namespace, parent namespace in case of hierarchical namespaces, et cetera.
//
// While the Relation interface strongly relates to the ops package, it is
// separate in order to avoid import cycles, because:
//   - the [github.com/thediveo/lxkns/ops/Opener] interface needs to reference the
//     [Relation] interface
//   - and we want to keep the Opener interface internal (due
//     to the knowledge required to correctly handle the object and file descriptor
//     lifecycles correctly).
package relations

import "github.com/thediveo/lxkns/species"

// Relation defines query operations on Linux-kernel namespaces for discovering
// relationships between namespaces as well as some properties, when only a
// namespace reference is given. Such references can be in form of a filesystem
// path, an open file descriptor, or an open file.
type Relation interface {
	// Type of the referenced namespace, such as CLONE_NEWNET, et cetera.
	// Returns an error in case of an invalid namespace reference (closed file
	// descriptor, invalid path, et cetera).
	//
	// ℹ️ A Linux kernel version 4.11 or later is required.
	Type() (species.NamespaceType, error)

	// ID (inode number) of the referenced namespace. Returns an error in case
	// of an invalid namespace reference.
	ID() (species.NamespaceID, error)

	// User namespace owning the referenced namespace. The owning user namespace
	// is returned in form of a NamespaceFile reference when there was no error
	// in retrieving the information.
	//
	// ℹ️ A Linux kernel version 4.9 or later is required.
	User() (Relation, error)

	// Parent namespace of the referenced PID or user namespace. Returns an
	// error if the parent doesn't exist, if the caller hasn't capabilities in
	// the parent namespace, or if the referenced namespace is neither a PID nor
	// a user namespace.
	//
	// ℹ️ A Linux kernel version 4.9 or later is required.
	Parent() (Relation, error)

	// User ID of the process originally creating the referenced user namespace.
	// Returns an error, if the referenced namespace is not a user namespace.
	//
	// ℹ️ A Linux kernel version 4.11 or later is required.
	OwnerUID() (int, error)
}
