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
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/testbasher"
)

var _ = Describe("cgrouping", func() {

	It("finds control groups of processes", func() {
		procs := NewProcessTable()
		Expect(procs).To(ContainElement(PointTo(MatchFields(IgnoreExtras, Fields{
			"Controlgroup": Not(BeEmpty()),
		}))))
	})

	It("gets fridge status", func() {
		// since we're going to mess around with control groups, we need to be
		// root (well, simplified constraint).
		if os.Geteuid() != 0 {
			Skip("needs root")
		}
		// Pick up the path of the freezer cgroup root; this allows this test to
		// automatically adjust. However it requires that when we run inside a
		// test container we got full cgroup access by bind-mounting the cgroups
		// root into our container. Otherwise, we won't be able to create our
		// own freezer (sub) cgroup controller :(
		fridgeroot := ""
	Fridge:
		for _, mountinfo := range mntinfo.MountsOfType(-1, "cgroup") {
			for _, sopt := range strings.Split(mountinfo.SuperOptions, ",") {
				if sopt == "freezer" {
					fridgeroot = mountinfo.MountPoint
					break Fridge
				}
			}
		}
		Expect(fridgeroot).NotTo(BeZero(), "detecting freezer cgroup root")

		freezercgname := fmt.Sprintf("lxkns%d", rand.Intn(100000)+100000)

		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Script("main", fmt.Sprintf(`
set -e
CTRL=%s
sleep 1d &
PID=$!
# create new freezer controller and put the sleep task under its control,
# then freeze it.
mkdir $CTRL 2>&1 # crash reading PID with error message from mkdir if it failed
echo $PID > $CTRL/cgroup.procs
# Safety guard: check that there's exactly one process under control.
cat $CTRL/cgroup.procs | wc -l
cat $CTRL/cgroup.procs
read # wait to proceed() and only then freeze the process.
echo "FROZEN" > $CTRL/freezer.state
read # wait to proceed() and then thaw the process again.
echo "THAWED" > $CTRL/freezer.state
read # wait to proceed().
kill $PID
rmdir $CTRL
`, filepath.Join(fridgeroot, freezercgname)))
		cmd := scripts.Start("main")
		defer cmd.Close()

		var controlleds int
		cmd.Decode(&controlleds)
		Expect(controlleds).To(Equal(1), "oh, no! Fridge %q is empty.", filepath.Join(fridgeroot, freezercgname))

		var pid PIDType
		cmd.Decode(&pid)
		Expect(pid).NotTo(BeZero())

		f := func() *Process {
			p := NewProcessTable()
			return p[pid]
		}
		Expect(f()).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Fridge":       Equal(ProcessThawed),
			"FridgeCgroup": Equal(filepath.Join("/", freezercgname)),
			"Selffridge":   Equal(ProcessThawed),
			"Parentfridge": Equal(ProcessThawed),
		})))

		// Freeze
		cmd.Proceed()
		Eventually(f, "4s", "500ms").Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Fridge":       Equal(ProcessFrozen),
			"Selffridge":   Equal(ProcessFrozen),
			"Parentfridge": Equal(ProcessThawed),
		})))

		// Thaw
		cmd.Proceed()
		Eventually(f, "4s", "500ms").Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Fridge":     Equal(ProcessThawed),
			"Selffridge": Equal(ProcessThawed),
		})))

		cmd.Proceed()
		Eventually(f, "4s", "500ms").Should(BeNil())
	})

})
