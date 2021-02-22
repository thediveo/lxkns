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
	"os"
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

		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Script("main", fmt.Sprintf(`
sleep 1d &
PID=$!
# create new freezer controller and put the sleep task under its control,
# then freeze it.
mkdir %[1]s/lxkns
echo $PID > %[1]s/lxkns/tasks
echo $PID
read # wait to proceed() and only then freeze the process.
echo "FROZEN" > %[1]s/lxkns/freezer.state
read # wait to proceed() and then thaw the process again.
echo "THAWED" > %[1]s/lxkns/freezer.state
read # wait to proceed().
kill $PID
rmdir %[1]s/lxkns
`, fridgeroot))
		cmd := scripts.Start("main")
		defer cmd.Close()
		var pid PIDType
		cmd.Decode(&pid)

		f := func() *Process {
			p := NewProcessTable()
			return p[pid]
		}
		Expect(f()).Should(PointTo(MatchFields(IgnoreExtras, Fields{
			"Fridge":       Equal(ProcessThawed),
			"FridgeCgroup": Equal("/lxkns"),
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
