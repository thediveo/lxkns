// Copyright 2023 Harald Albrecht.
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

package helpers

import (
	. "github.com/onsi/gomega"
)

// Successful takes a return value together with an additional error return
// value, returning only the value and at the same time asserting that there
// error return value is nil.
func Successful[V any](v V, err error) V {
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	return v
}
