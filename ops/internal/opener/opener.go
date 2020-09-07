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

package opener

import r "github.com/thediveo/lxkns/ops/relations"

// Opener is a module-internal interface to namespace references, to get file
// descriptor references to namespaces regardless of the particular type of
// reference; be it path-based, file-based, or fd-based. Also optimizes those
// usecases where the type of namespace is known in advance.
type Opener interface {
	// OpenTypedReference returns a typed and "opened" namespace reference,
	// ready for reference in Linux syscalls by a file descriptor. When used on
	// a non-typed namespace reference, the correct type of namespace will
	// automatically be queried (requires a kernel 4.11+), otherwise the already
	// known type will be used instead, if available. If opening or type
	// inference fails, an error will be returned instead. If the call succeeds,
	// then the caller must make sure to call the returned ReferenceCloser
	// function in order to properly release process resources after the open
	// namespace reference isn't needed anymore.
	OpenTypedReference() (r.Relation, ReferenceCloser, error)

	// NsFd returns a file descriptor referencing the namespace indicated in a
	// namespace reference implementing the Opener interface. After the caller
	// is done using the returned file descriptor, the caller must call the
	// returned FdCloser function in order to properly release process
	// resources. In case of any error when opening the referenced namespace,
	// err will be non-nil, and might additionally wrap an underlying OS-level
	// error.
	//
	// In case of a NamespaceFile-based reference the caller must make sure that
	// the object does not prematurely get garbage collected before the file
	// descriptor is used, if in doubt, use runtime.KeepAlive(nsref), see also:
	// https://golang.org/pkg/runtime/#KeepAlive.
	//
	// This method is named NsFd() instead of Fd() on purpose, as to avoid name
	// clashes with the existing os.File.Fd() method (not least in our own
	// NamespaceFile type).
	NsFd() (int, FdCloser, error)
}

// ReferenceCloser is a module-internal function which needs to be called in
// order to properly release process resources when done with a TypedReference.
type ReferenceCloser func()

// FdCloser is a module-internal function which needs to be called in order to
// properly release process resources when done with a ops.Relation.
type FdCloser func()
