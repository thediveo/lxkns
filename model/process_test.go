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

package model

import (
	"os"
	"sort"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			Expect(newProcessFromStatline(badstat)).To(BeNil(), badstat)
		}
	})

	It("cannot be created for non-existing process/PID", func() {
		Expect(NewProcess(0)).To(BeNil())
	})

	It("skips broken process stat", func() {
		Expect(newProcess(1, "test/proctable/kaputt")).To(BeNil())
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

	It("stringifies descriptive properties", func() {
		me := NewProcess(PIDType(os.Getpid()))
		s := me.String()
		const startre = `(^|\s|[[:punct:]])`
		const endre = `($|\s|[[:punct:]])`
		Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getpid()) + endre))
		Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getppid()) + endre))
		Expect(s).To(MatchRegexp(startre + me.Name + endre))
	})

	It("gets basename and command line", func() {
		proc42 := newProcess(PIDType(42), "test/proctable/proc")
		Expect(proc42.Cmdline).To(HaveLen(3))
		Expect(proc42.Basename()).To(Equal("mumble.exe"))
		Expect(proc42.Cmdline[2], "arg2")

		// $0 doesn't contain any "/"
		proc667 := newProcess(PIDType(667), "test/proctable/kaputt")
		Expect(proc667.Basename()).To(Equal("mumble.exe"))
	})

	It("falls back on process name", func() {
		// Please note that our synthetic PID 1 has no command line, but only
		// a process name in its stat file.
		proc1 := newProcess(PIDType(1), "test/proctable/proc")
		Expect(proc1.Basename()).To(Equal("init"))
	})

	It("synthesizes basename if all else fails", func() {
		proc := newProcess(PIDType(666), "test/proctable/kaputt")
		Expect(proc.Basename()).To(Equal("process (666)"))
	})

})

var _ = Describe("ProcessTable", func() {

	It("reads synthetic /proc", func() {
		pt := newProcessTable("test/proctable/proc")
		Expect(pt).NotTo(BeNil())
		Expect(pt).To(HaveLen(2))

		proc1 := pt[1]
		proc42 := pt[42]
		Expect(proc1).NotTo(BeNil())
		Expect(proc1.Parent).To(BeNil())
		Expect(proc1.Children).To(HaveLen(1))
		Expect(proc1.Children[0]).To(BeIdenticalTo(proc42))
	})

	It("returns nil for inaccessible /proc", func() {
		Expect(newProcessTable("test/nirvana")).To(BeNil())
	})

	It("gathers from real /proc", func() {
		pt := NewProcessTable()
		Expect(pt).NotTo(BeNil())
		proc := pt[PIDType(os.Getpid())]
		Expect(proc).NotTo(BeZero())
		Expect(proc.Parent).NotTo(BeNil())
		Expect(proc.Parent.PID).To(Equal(PIDType(os.Getppid())))
	})

})

var _ = Describe("ProcessListByPID", func() {

	It("sorts Process lists", func() {
		pls := [][]*Process{
			{
				&Process{PID: 1, Name: "foo"},
				&Process{PID: 42, Name: "bar"},
			},
			{
				&Process{PID: 42, Name: "bar"},
				&Process{PID: 1, Name: "foo"},
			},
		}
		for _, pl := range pls {
			sort.Sort(ProcessListByPID(pl))
			Expect(pl[0].PID).To(Equal(PIDType(1)))
			Expect(pl[1].PID).To(Equal(PIDType(42)))
		}
	})

})
