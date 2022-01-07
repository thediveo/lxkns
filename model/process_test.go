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

	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = g.Describe("Process", func() {

	g.It("stringifies", func() {
		var p *Process
		Expect(p.String()).To(MatchRegexp(`<nil>`))

		Expect(NewProcess(1)).To(MatchRegexp(`PID 1.+PPID 0`))
	})

	g.It("rejects invalid /proc/[PID] status lines", func() {
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

	g.It("cannot be created for non-existing process/PID", func() {
		Expect(NewProcess(0)).To(BeNil())
	})

	g.It("skips broken process stat", func() {
		Expect(NewProcessInProcfs(1, "test/proctable/kaputt")).To(BeNil())
	})

	g.It("properties are read from /proc/[PID]", func() {
		pid := PIDType(os.Getpid())
		me := NewProcess(pid)
		Expect(me).NotTo(BeNil())
		Expect(me.PID).To(Equal(pid))
		Expect(me.PPID).To(Equal(PIDType(os.Getppid())))
	})

	g.It("validates it exists or exits", func() {
		me := NewProcess(PIDType(os.Getpid()))
		Expect(me.Valid()).To(BeTrue())
		me.PID = PIDType(1)
		Expect(me.Valid()).NotTo(BeTrue())
	})

	g.It("stringifies descriptive properties", func() {
		me := NewProcess(PIDType(os.Getpid()))
		s := me.String()
		const startre = `(^|\s|[[:punct:]])`
		const endre = `($|\s|[[:punct:]])`
		Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getpid()) + endre))
		Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getppid()) + endre))
		Expect(s).To(MatchRegexp(startre + me.Name + endre))
	})

	g.It("gets basename and command line", func() {
		proc42 := NewProcessInProcfs(PIDType(42), "test/proctable/proc")
		Expect(proc42.Cmdline).To(HaveLen(3))
		Expect(proc42.Basename()).To(Equal("mumble.exe"))
		Expect(proc42.Cmdline[2], "arg2")

		// $0 doesn't contain any "/"
		proc667 := NewProcessInProcfs(PIDType(667), "test/proctable/kaputt")
		Expect(proc667.Basename()).To(Equal("mumble.exe"))
	})

	g.It("falls back on process name", func() {
		// Please note that our synthetic PID 1 has no command line, but only
		// a process name in its stat file.
		proc1 := NewProcessInProcfs(PIDType(1), "test/proctable/proc")
		Expect(proc1.Basename()).To(Equal("init"))
	})

	g.It("synthesizes basename if all else fails", func() {
		proc := NewProcessInProcfs(PIDType(666), "test/proctable/kaputt")
		Expect(proc.Basename()).To(Equal("process (666)"))
	})

})

var _ = g.Describe("ProcessTable", func() {

	g.It("reads synthetic /proc", func() {
		pt := NewProcessTableFromProcfs(false, "test/proctable/proc")
		Expect(pt).NotTo(BeNil())
		Expect(pt).To(HaveLen(2))

		proc1 := pt[1]
		proc42 := pt[42]
		Expect(proc1).NotTo(BeNil())
		Expect(proc1.Parent).To(BeNil())
		Expect(proc1.Children).To(HaveLen(1))
		Expect(proc1.Children[0]).To(BeIdenticalTo(proc42))
	})

	g.It("returns nil for inaccessible /proc", func() {
		Expect(NewProcessTableFromProcfs(false, "test/nirvana")).To(BeNil())
	})

	g.It("gathers from real /proc", func() {
		pt := NewProcessTable(false)
		Expect(pt).NotTo(BeNil())
		proc := pt[PIDType(os.Getpid())]
		Expect(proc).NotTo(BeZero())
		Expect(proc.Parent).NotTo(BeNil())
		Expect(proc.Parent.PID).To(Equal(PIDType(os.Getppid())))
	})

})

var _ = g.Describe("ProcessListByPID", func() {

	g.It("sorts Process lists", func() {
		p1 := &Process{PID: 1, Name: "foo"}
		p42 := &Process{PID: 42, Name: "bar"}
		pls := [][]*Process{
			{p1, p42},
			{p42, p1},
		}
		for _, pl := range pls {
			sort.Sort(ProcessListByPID(pl))
			Expect(pl[0].PID).To(Equal(PIDType(1)))
			Expect(pl[1].PID).To(Equal(PIDType(42)))
		}
	})

})
