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

import (
	"fmt"

	r "github.com/thediveo/lxkns/ops/relations"
)

// InvalidNamespaceError wraps an underlying OS-related error when dealing with
// Linux-kernel namespaces. Due to Golang's attempt at abstracting things, this
// might often be an os.PathError, in its turn wrapping a syscall error, such as
// syscall.EBADF, syscall.EINVAL, syscall.EPERM, et cetera.
type InvalidNamespaceError struct {
	Ref string // textual representation of a namespace reference.
	Err error  // wrapped OS-level error.
}

// NamespaceOperationError wraps an invalid namespace operation, giving
// information about the failed operation both on a high level, as well as the
// underlying invalid namespace and OS-level errors.
type NamespaceOperationError struct {
	InvalidNamespaceError
	Op string // failed namespace ioctl operation
}

// newInvalidNamespaceError returns a descriptive error, also wrapping an
// underlying (OS-level) error giving more details when desired.
func newInvalidNamespaceError(nsref r.Relation, err error) *InvalidNamespaceError {
	if nsref == nil {
		return &InvalidNamespaceError{"", err}
	}
	return &InvalidNamespaceError{nsref.(fmt.Stringer).String(), err}
}

// newNamespaceOperationError returns a descriptive error, also wrapping an
// underlying (OS-level) error giving more details when desired.
func newNamespaceOperationError(nsref r.Relation, op string, err error) *NamespaceOperationError {
	return &NamespaceOperationError{
		InvalidNamespaceError{nsref.(fmt.Stringer).String(), err},
		op,
	}
}

// Error returns a textual description of this invalid namespace error.
func (e *InvalidNamespaceError) Error() string {
	s := "lxkns: invalid namespace"
	if e.Ref != "" {
		s += " " + e.Ref
	}
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

// Unwrap returns the error underlying an invalid namespace error.
func (e *InvalidNamespaceError) Unwrap() error { return e.Err }

// Error returns a textual description of this invalid namespace error.
func (e *NamespaceOperationError) Error() string {
	s := fmt.Sprintf("lxkns: invalid namespace operation %s on %s",
		e.Op, e.Ref)
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}
