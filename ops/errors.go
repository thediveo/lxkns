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

// newInvalidNamespaceError returns a descriptive error, also wrapping an
// underlying (OS-level) error giving more details when desired.
func newInvalidNamespaceError(nsref r.Relation, err error) *InvalidNamespaceError {
	return &InvalidNamespaceError{nsref.(fmt.Stringer).String(), err}
}

// Error returns a textual description of this invalid namespace error.
func (e *InvalidNamespaceError) Error() string {
	s := "lxkns: invalid namespace " + e.Ref
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

// Unwrap returns the error underlying an invalid namespace error.
func (e *InvalidNamespaceError) Unwrap() error { return e.Err }
