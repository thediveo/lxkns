// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gmodel

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
)

var (
	pt1  = model.ProcessTable{proc1.PID: proc1}
	pt11 = model.ProcessTable{proc1.PID: proc1}
	pt12 = model.ProcessTable{proc1.PID: proc1, proc2.PID: proc2}
	pt2  = model.ProcessTable{proc2.PID: proc2}
)

var _ = Describe("ProcessTable", func() {

	It("handles mistakes", func() {
		_, err := BeSameProcessTable(nil).Match(nil)
		Expect(err).To(MatchError(MatchRegexp(`use BeNil()`)))
		_, err = BeSameProcessTable(pt1).Match("foo")
		Expect(err).To(MatchError(MatchRegexp(`expects a model.ProcessTable, not a string`)))
		_, err = BeSameProcessTable("foo").Match(pt1)
		Expect(err).To(MatchError(MatchRegexp(`must be passed a model.ProcessTable, not a string`)))
	})

	It("matches, or not", func() {
		Expect(pt1).To(BeSameProcessTable(pt11))
		Expect(pt1).NotTo(BeSameProcessTable(pt2))
		Expect(pt1).NotTo(BeSameProcessTable(pt12))
	})

})
