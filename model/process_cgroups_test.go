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
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/thediveo/go-mntinfo"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/lxkns/test/success"
)

var _ = Describe("Freezer", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("reads v1 freezer state", func() {
		Expect(frozenV1("test/cgroupies/v1")).To(BeFalse())
		Expect(frozenV1("test/cgroupies/v1/thawed")).To(BeFalse())
		Expect(frozenV1("test/cgroupies/v1/somethingelse")).To(BeFalse())
		Expect(frozenV1("test/cgroupies/v1/freezing")).To(BeTrue())
		Expect(frozenV1("test/cgroupies/v1/frozen")).To(BeTrue())
	})

	It("reads v2 freezer state", func() {
		Expect(frozenV2("test/cgroupies/v2")).To(BeFalse())
		Expect(frozenV2("test/cgroupies/v2/thawed")).To(BeFalse())
		Expect(frozenV2("test/cgroupies/v2/gnawed")).To(BeFalse())
		Expect(frozenV2("test/cgroupies/v2/frozen")).To(BeTrue())
	})

})

var _ = Describe("cgrouping", func() {

	It("finds control groups of processes", func() {
		procs := NewProcessTable(false)
		Expect(procs).To(ContainElement(And(
			HaveField("CpuCgroup", Not(BeEmpty())),
			HaveField("FridgeCgroup", Not(BeEmpty())),
		)))
	})

	It("gets fridge status", func() {
		// since we're going to mess around with control groups, we need to be
		// root (well, simplified constraint).
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		By("detecting the freezer cgroup path")
		// Pick up the path of the freezer v1 cgroup root; this allows this test
		// to automatically adjust. However it requires that when we run inside
		// a test container we got full cgroup access by bind-mounting the
		// cgroups root into our container. Otherwise, we won't be able to
		// create our own freezer (sub) cgroup controller :(
		cgroupsv2 := false
		mount, ok := lo.Find(mntinfo.MountsOfType(-1, "cgroup"),
			func(m mntinfo.Mountinfo) bool {
				return lo.Contains(strings.Split(m.SuperOptions, ","), "freezer")
			})
		var fridgeroot string
		if ok {
			fridgeroot = mount.MountPoint
		} else {
			// If we couldn't find a v1 freezer then there must be a unified v2
			// hierarchy, so let's take that instead.
			fridgeroot = mntinfo.MountsOfType(-1, "cgroup2")[0].MountPoint
			cgroupsv2 = true
		}
		Expect(fridgeroot).NotTo(BeZero(), "not detecting freezer cgroup root")
		By("freezer cgroup path: " + fridgeroot)

		// We unfortunately can't use a thread of our own process due to the
		// limitations in cgroups v2.
		By("spawning process as specimen")
		session := Successful(gexec.Start(exec.Command("/bin/sleep", "60"), nil, nil))
		defer session.Kill() // can be called multiple times.
		sleepypid := PIDType(session.Command.Process.Pid)

		By("creating a new freezer controller and putting the specimen under its control")
		freezerctrl := path.Join(fridgeroot, fmt.Sprintf("lxkns%d", os.Getpid()))
		Expect(os.Mkdir(freezerctrl, 0o755)).To(Succeed())
		Expect(os.WriteFile(path.Join(freezerctrl, "cgroup.procs"),
			[]byte(strconv.Itoa(int(sleepypid))), fs.ModePerm)).To(Succeed())

		By("cross-checking")
		undercontrol := string(Successful(os.ReadFile(path.Join(freezerctrl, "cgroup.procs"))))
		Expect(strings.Count(undercontrol, "\n")).To(Equal(1))

		sleepyproc := func() *Process {
			p := NewProcessTable(true)
			proc := p[sleepypid]
			return proc
		}
		sleepytask := func() *Task {
			p := NewProcessTableWithTasks(true)
			proc, ok := p[sleepypid]
			if !ok {
				return nil
			}
			task, _ := lo.Find(proc.Tasks,
				func(t *Task) bool { return t.TID == sleepypid })
			return task
		}

		Expect(sleepyproc()).To(And(
			HaveField("FridgeFrozen", false),
			HaveField("FridgeCgroup", path.Join("/", path.Base(freezerctrl))),
		))
		Expect(sleepytask()).To(And(
			HaveField("FridgeFrozen", false),
			HaveField("FridgeCgroup", path.Join("/", path.Base(freezerctrl))),
		))

		By("freezing the specimen")
		if cgroupsv2 {
			Expect(os.WriteFile(path.Join(freezerctrl, "cgroup.freeze"),
				[]byte("1\n"), os.ModePerm)).To(Succeed())
		} else {
			Expect(os.WriteFile(path.Join(freezerctrl, "freezer.state"),
				[]byte("FROZEN\n"), os.ModePerm)).To(Succeed())
		}
		Eventually(sleepyproc).Within(5 * time.Second).ProbeEvery(500 * time.Millisecond).
			Should(HaveField("FridgeFrozen", true))
		Eventually(sleepytask).Within(1 * time.Second).ProbeEvery(500 * time.Millisecond).
			Should(HaveField("FridgeFrozen", true))

		By("thawing the specimen")
		if cgroupsv2 {
			Expect(os.WriteFile(path.Join(freezerctrl, "cgroup.freeze"),
				[]byte("0\n"), os.ModePerm)).To(Succeed())
		} else {
			Expect(os.WriteFile(path.Join(freezerctrl, "freezer.state"),
				[]byte("THAWED\n"), os.ModePerm)).To(Succeed())
		}
		Eventually(sleepyproc).Within(5 * time.Second).ProbeEvery(500 * time.Millisecond).
			Should(HaveField("FridgeFrozen", false))
		Eventually(sleepytask).Within(1 * time.Second).ProbeEvery(500 * time.Millisecond).
			Should(HaveField("FridgeFrozen", false))

		By("getting rid of the specimen")
		session.Kill()
		Eventually(sleepyproc).Within(5 * time.Second).ProbeEvery(500 * time.Millisecond).
			Should(BeNil())
		Eventually(sleepytask).Within(1 * time.Second).ProbeEvery(500 * time.Millisecond).
			Should(BeNil())
	})

})
