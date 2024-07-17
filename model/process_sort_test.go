// Copyright 2024 Harald Albrecht.
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

package model

import (
	deco "github.com/onsi/ginkgo/v2/dsl/decorators"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/ginkgo/v2/dsl/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mr Wehrli sorting processes", func() {

	It("detects the system's PID wrap-around", func() {
		mask, dist := pidWrapping(0)
		Expect(mask).NotTo(BeZero())
		Expect(dist).NotTo(BeZero())
		Expect((dist << 1) - 1).To(Equal(mask))
	})

	Context("interval arithmetic", deco.Ordered, func() {

		BeforeAll(func() {
			pidMaxMask, pidMaxDist = 8-1, 8>>1
			DeferCleanup(func() {
				pidMaxMask, pidMaxDist = pidWrapping((uint64(1) << 22))
			})
		})

		DescribeTable("sorting by age and PID distance",
			func(ageA int, pidA int, ageB int, pidB int, expect int) {
				delta := SortProcessByAgeThenPIDDistance(
					&Process{PID: PIDType(pidA), ProTaskCommon: ProTaskCommon{Starttime: uint64(ageA)}},
					&Process{PID: PIDType(pidB), ProTaskCommon: ProTaskCommon{Starttime: uint64(ageB)}})
				Expect(delta).To(Equal(expect),
					"(%d-%d)&%x=%x ?? %x", pidA, pidB, pidMaxMask, (pidB-pidA)&int(pidMaxMask), pidMaxDist)
			},
			Entry("a older than b", 100, 1, 200, 2, -1),
			Entry("a younger than b", 200, 1, 100, 2, 1),
			Entry("a same age as b, PID a before PID b, nowrap", 100, 4, 100, 5, -1),
			Entry("a same age as b, PID b before PID a, nowrap", 100, 5, 100, 4, 1),
			Entry("a same age as b, PID a before PID b, wrap", 100, 7, 100, 1, -1),
			Entry("a same age as b, PID b before PID a, wrap", 100, 1, 100, 7, 1),
		)

	})

})
