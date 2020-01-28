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

package lxkns

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strconv"
)

var _ = Describe("Process", func() {

	It("rejects invalid /proc/[PID] status lines", func() {
		// Test various invalid field combinations
		for _, badstat := range []string{
			"42",
			"X (something)",
			"42 42",
			"42 (grmpf",
			"42 (grumpf)",
			"42 (gru) mpf) ",
			"42 (gru) mpf) R",
			//             3 4  5    6   7   8   9   10  11  12  13  14  15  16  17  18  19  20  21  22
			"42 (gru) mpf) R 1 1234 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123",
			"42 (gru) mpf) R x 1234 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 1",
			"42 (gru) mpf) R -1 1234 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 1",
			"42 (gru) mpf) R 1 1234 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 -1",
			"42 (gru) mpf) R 1 1234 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 123 x",
		} {
			Expect(newProcess(badstat)).To(BeNil(), badstat)
		}
	})

	It("cannot be created for non-existing process/PID", func() {
		Expect(NewProcess(0)).To(BeNil())
	})

	It("properties are read from /proc/[PID]", func() {
		pid := PIDType(os.Getpid())
		me := NewProcess(pid)
		Expect(me).NotTo(BeNil())
		Expect(me.PID).To(Equal(pid))
		Expect(me.PPID).To(Equal(PIDType(os.Getppid())))
	})

	It("validates it exists or exits", func() {
		me := NewProcess(PIDType(os.Getpid()))
		Expect(me.Valid()).To(BeTrue())
		me.PID = PIDType(1)
		Expect(me.Valid()).NotTo(BeTrue())
	})

	It("stringifies some properties", func() {
		me := NewProcess(PIDType(os.Getpid()))
		s := me.String()
		const startre = `(^|\s|[[:punct:]])`
		const endre = `($|\s|[[:punct:]])`
		Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getpid()) + endre))
		Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getppid()) + endre))
		Expect(s).To(MatchRegexp(startre + me.Name + endre))
	})

})

var _ = Describe("ProcessTable", func() {

	It("gathered from /proc", func() {
		pt := NewProcessTable()
		Expect(pt).NotTo(BeNil())
		pid := PIDType(os.Getpid())
		Expect(pt[pid]).NotTo(BeZero())
		Expect(pt[pid].Parent).NotTo(BeNil())
		Expect(pt[pid].Parent.PID).To(Equal(PIDType(os.Getppid())))
	})

})
