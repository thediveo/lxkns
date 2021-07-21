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

package pidmap

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var _ = Describe("maps PIDs", func() {

	It("returns empty PID slice for non-existing PID 0", func() {
		Expect(NSpid(&model.Process{})).To(BeEmpty())
	})

	It("panics on invalid process status", func() {
		p := model.NewProcessInProcfs(model.PIDType(668), "test/proc")
		Expect(p).NotTo(BeNil())
		Expect(func() { nspid(p, "test/proc") }).To(PanicWith(MatchRegexp(`filesystem broken`)))
	})

	It("ignores invalid NSpid entries", func() {
		p := model.NewProcessInProcfs(model.PIDType(669), "test/proc")
		Expect(p).NotTo(BeNil())
		Expect(nspid(p, "test/proc")).To(HaveLen(0))
	})

	It("reads namespaced PIDs of process", func() {
		pt := model.NewProcessTableFromProcfs(false, "test/proc")
		Expect(pt).NotTo(BeNil())
		Expect(pt).To(HaveLen(4))
		Expect(pt).To(HaveKey(model.PIDType(4200)))

		pids := nspid(pt[model.PIDType(4200)], "test/proc")
		Expect(pids).To(ConsistOf(model.PIDType(4200), model.PIDType(1)))
	})

	/*
		It("doesn't translates non-existing PID/namespace", func() {
			opts := NoDiscovery
			opts.SkipProcs = false
			opts.SkipHierarchy = false
			allns := Discover(opts)
			pidmap := NewPIDMap(allns)
			Expect(pidmap.Translate(0, allns.InitialNamespaces[model.PIDNS], allns.InitialNamespaces[model.PIDNS])).To(BeZero())
		})
	*/

	It("translates PIDs", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
		unshare -Umr $stage2
		`)
		scripts.Script("stage2", `
		unshare -pf $pidxlas3
		`)
		scripts.Script("pidxlas3", `
		# in order to get the PID in the parent PID namespace we rely on
		# procfs not remounted yet and /proc/self pointing to us in the
		# parent PID namespace. We must not spawn a new child process in
		# trying to get our PID, so bas built-ins to the rescue.
		read </proc/self/stat PID _ && echo "$PID" # "outer" PID
		# remount procfs to pick up the new PID namespace.
		mount -t proc proc /proc
		process_namespaceid pid # print ID of new PID namespace.
		echo "$$" # PID in new/leaf PID namespace.
		read # wait for test to proceed()
		`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var initleafpid model.PIDType
		cmd.Decode(&initleafpid)
		Expect(initleafpid).NotTo(Equal(model.PIDType(1)))
		pidnsid := nstest.CmdDecodeNSId(cmd)
		Expect(pidnsid).NotTo(BeZero())
		var leafpid model.PIDType
		cmd.Decode(&leafpid)
		Expect(leafpid).NotTo(Equal(initleafpid))

		// Cannot run discovery in test here due to circular dependency. So we
		// need to do emulate some things. Yes, this is getting ugly.
		pt := model.NewProcessTable(false)
		Expect(pt).NotTo(BeZero())
		selfpid := model.PIDType(os.Getpid())
		initpidnsref := ops.NamespacePath(fmt.Sprintf("/proc/%d/ns/pid", selfpid))
		initpidnsid, err := initpidnsref.ID()
		Expect(err).NotTo(HaveOccurred())

		initialpidns := namespaces.New(species.CLONE_NEWPID, initpidnsid, string(initpidnsref))
		Expect(initialpidns).NotTo(BeNil())
		pt[selfpid].Namespaces[model.PIDNS] = initialpidns

		leafpidnsref := ops.NamespacePath(fmt.Sprintf("/proc/%d/ns/pid", initleafpid))
		leafpidnsid, err := leafpidnsref.ID()
		Expect(err).NotTo(HaveOccurred())
		leafpidns := namespaces.New(species.CLONE_NEWPID, leafpidnsid, string(leafpidnsref))
		Expect(leafpidns).NotTo(BeNil())
		pt[initleafpid].Namespaces[model.PIDNS] = leafpidns

		initialpidns.(namespaces.HierarchyConfigurer).AddChild(leafpidns.(model.Hierarchy))

		pm := NewPIDMap(pt)
		Expect(pm).NotTo(BeEmpty())

		Expect(pm.NamespacedPIDs(model.PIDType(0), initialpidns)).To(BeEmpty())

		// We should see the test process from the initial PID namespace...
		Expect(pm.NamespacedPIDs(leafpid, leafpidns)).To(ConsistOf(
			model.NamespacedPID{PIDNS: initialpidns, PID: initleafpid},
			model.NamespacedPID{PIDNS: leafpidns, PID: leafpid},
		))

		// ...and also from its own child PID namespace.
		Expect(pm.NamespacedPIDs(initleafpid, initialpidns)).To(ConsistOf(
			model.NamespacedPID{PIDNS: initialpidns, PID: initleafpid},
			model.NamespacedPID{PIDNS: leafpidns, PID: leafpid},
		))

		// Translate forth and back. That's forth, not Forth, argh!
		Expect(pm.Translate(initleafpid, initialpidns, leafpidns)).To(Equal(leafpid))
		Expect(pm.Translate(leafpid, leafpidns, initialpidns)).To(Equal(initleafpid))

		Expect(pm.Translate(0, initialpidns, initialpidns)).To(Equal(model.PIDType(0)))

	})

})
