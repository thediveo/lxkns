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
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strconv"
	"time"

	"github.com/samber/lo"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("processes and tasks", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	Context("processes", func() {

		It("stringifies", func() {
			var p *Process
			Expect(p.String()).To(MatchRegexp(`<nil>`))

			Expect(NewProcess(1, false)).To(MatchRegexp(`PID 1.+PPID 0`))
		})

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
			Expect(NewProcess(0, false)).To(BeNil())
		})

		It("skips broken process stat", func() {
			Expect(NewProcessInProcfs(1, false, "test/proctable/kaputt")).To(BeNil())
		})

		It("properties are read from /proc/[PID]", func() {
			pid := PIDType(os.Getpid())
			me := NewProcess(pid, false)
			Expect(me).NotTo(BeNil())
			Expect(me.PID).To(Equal(pid))
			Expect(me.PPID).To(Equal(PIDType(os.Getppid())))
		})

		It("validates it exists or exits", func() {
			me := NewProcess(PIDType(os.Getpid()), false)
			Expect(me.Valid()).To(BeTrue())
			me.PID = PIDType(1)
			Expect(me.Valid()).NotTo(BeTrue())
		})

		It("stringifies descriptive properties", func() {
			me := NewProcess(PIDType(os.Getpid()), false)
			s := me.String()
			const startre = `(^|\s|[[:punct:]])`
			const endre = `($|\s|[[:punct:]])`
			Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getpid()) + endre))
			Expect(s).To(MatchRegexp(startre + strconv.Itoa(os.Getppid()) + endre))
			Expect(s).To(MatchRegexp(startre + me.Name + endre))
		})

		It("gets basename and command line", func() {
			proc42 := NewProcessInProcfs(PIDType(42), false, "test/proctable/proc")
			Expect(proc42.Cmdline).To(HaveLen(3))
			Expect(proc42.Basename()).To(Equal("mumble.exe"))
			Expect(proc42.Cmdline[2]).To(Equal("arg2"))

			// $0 doesn't contain any "/"
			proc667 := NewProcessInProcfs(PIDType(667), false, "test/proctable/kaputt")
			Expect(proc667.Basename()).To(Equal("mumble.exe"))
		})

		It("falls back on process name", func() {
			// Please note that our synthetic PID 1 has no command line, but only
			// a process name in its stat file.
			proc1 := NewProcessInProcfs(PIDType(1), false, "test/proctable/proc")
			Expect(proc1.Basename()).To(Equal("init"))
		})

		It("synthesizes basename if all else fails", func() {
			proc := NewProcessInProcfs(PIDType(666), false, "test/proctable/kaputt")
			Expect(proc.Basename()).To(Equal("process (666)"))
		})

	})

	Context("tasks", func() {

		It("rejects invalid status lines", func() {
			Expect(newTaskFromStatline("42 (gru) mpf) R", nil)).To(BeNil())
		})

		It("discovers the tasks of a process", func() {
			done := make(chan struct{})
			tidch := make(chan int)
			defer close(done)
			defer close(tidch)
			for i := 1; i <= 2; i++ {
				go func(i int) {
					runtime.LockOSThread()
					tidch <- unix.Gettid()
					<-done
				}(i)
			}
			proc := NewProcess(PIDType(os.Getpid()), true)
			Expect(proc).NotTo(BeNil())

			Expect(proc.Tasks).NotTo(BeEmpty())
			tids := lo.Map(proc.Tasks,
				func(task *Task, _ int) int { return int(task.TID) })

			for i := 1; i <= 2; i++ {
				Eventually(tidch).Should(Receive(
					BeElementOf(tids)))
			}

			for _, task := range proc.Tasks {
				Expect(task.TID).NotTo(BeZero())
				Expect(task.Process).To(Equal(proc))
				Expect(task.Name).NotTo(BeEmpty())
				Expect(task.Starttime).To(BeNumerically(">=", proc.Starttime))
			}

			var taskgroupleader *Task
			Expect(proc.Tasks).To(ContainElement(
				HaveField("TID", proc.PID),
				&taskgroupleader))
			lo.ForEach(proc.Tasks, func(task *Task, _ int) {
				Expect(task.MainTask()).To(Equal(task == taskgroupleader))
			})
		})

	})

})

var _ = Describe("process table", func() {

	It("reads synthetic /proc", func() {
		pt := NewProcessTableFromProcfs(false, false, "test/proctable/proc")
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
		Expect(NewProcessTableFromProcfs(false, false, "test/nirvana")).To(BeNil())
	})

	It("gathers from real /proc", func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		pt := NewProcessTable(false)
		Expect(pt).NotTo(BeNil())
		proc := pt[PIDType(os.Getpid())]
		Expect(proc).NotTo(BeZero())
		Expect(proc.Parent).NotTo(BeNil())
		Expect(proc.Parent.PID).To(Equal(PIDType(os.Getppid())))
	})

	It("returns Process objects for PIDs", func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		pt := NewProcessTable(false)
		Expect(pt).NotTo(BeNil())
		procs := pt.ProcessesByPIDs(PIDType(os.Getpid()))
		Expect(procs).To(HaveLen(1))
		Expect(procs[0].PID).To(Equal(PIDType(os.Getpid())))
	})

})

var _ = Describe("process lists", func() {

	It("sorts Process slices numerically by PID", func() {
		p1 := &Process{PID: 1, ProTaskCommon: ProTaskCommon{Name: "foo"}}
		p42 := &Process{PID: 42, ProTaskCommon: ProTaskCommon{Name: "bar"}}
		pls := [][]*Process{
			{p1, p42},
			{p42, p1},
		}
		for _, pl := range pls {
			slices.SortFunc(pl, SortProcessByPID)
			Expect(pl[0].PID).To(Equal(PIDType(1)))
			Expect(pl[1].PID).To(Equal(PIDType(42)))
		}
	})

})

var _ = Describe("cpu affinity", func() {

	It("retrieves cpu affinities of processes and tasks", func() {
		proc := NewProcess(PIDType(os.Getpid()), true)
		Expect(proc).NotTo(BeNil())
		Expect(proc.RetrieveAffinity()).To(Succeed())
		Expect(proc.Affinity).NotTo(BeEmpty())
		Expect(proc.Nice).To(Equal(0))
		Expect(proc.Priority).To(Equal(0))

		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		var task *Task
		Expect(proc.Tasks).To(ContainElement(HaveField("TID", PIDType(unix.Gettid())), &task))
		Expect(task.RetrieveAffinity()).To(Succeed())
		Expect(task.Affinity).NotTo(BeEmpty())
		Expect(task.Affinity).To(Equal(proc.Affinity))
		Expect(proc.Nice).To(Equal(0))
		Expect(proc.Priority).To(Equal(0))
	})

	It("has no fun without scheduling risk", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		runtime.LockOSThread()

		tid := unix.Gettid()
		proc := NewProcess(PIDType(os.Getpid()), true)
		var task *Task
		Expect(proc.Tasks).To(ContainElement(HaveField("TID", PIDType(tid)), &task))

		func() {
			oldschedattr := Successful(unix.SchedGetAttr(0, 0))
			Expect(oldschedattr.Size).NotTo(BeZero())
			defer func() {
				Expect(unix.SchedSetAttr(0, oldschedattr, 0)).To(Succeed())
			}()

			newschedattr := *oldschedattr
			newschedattr.Flags = unix.SCHED_FLAG_RESET_ON_FORK
			newschedattr.Policy = unix.SCHED_BATCH
			newschedattr.Nice = -20
			Expect(unix.SchedSetAttr(0, &newschedattr, 0)).To(Succeed())

			proc = NewProcess(PIDType(os.Getpid()), true)
		}()

		runtime.UnlockOSThread()

		Expect(proc).NotTo(BeNil())
		Expect(proc.Tasks).To(ContainElement(HaveField("TID", PIDType(tid)), &task))
		Expect(task.RetrieveAffinity()).To(Succeed())

		Expect(task.Policy).To(Equal(unix.SCHED_BATCH))
		Expect(task.Nice).To(Equal(-20))
	})

})
