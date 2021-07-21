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
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/testbasher"
)

var _ = Describe("maps PIDs", func() {

	It("doesn't translates non-existing PID/namespace", func() {
		allns := Discover(FromProcs(), WithHierarchy())
		pidmap := NewPIDMap(allns)
		Expect(pidmap.Translate(0, allns.InitialNamespaces[model.PIDNS], allns.InitialNamespaces[model.PIDNS])).To(BeZero())
	})

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
mount -t proc proc /proc
process_namespaceid pid # print ID of new PID namespace.
echo "$$"
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		pidnsid := nstest.CmdDecodeNSId(cmd)
		var leafpid model.PIDType
		cmd.Decode(&leafpid)

		allns := Discover(FromProcs(), WithHierarchy())
		pidns := allns.Namespaces[model.PIDNS][pidnsid]
		initialpidns := allns.PIDNSRoots[0]

		pidmap := NewPIDMap(allns)

		pid := pidmap.Translate(leafpid, pidns, initialpidns)
		Expect(pid).NotTo(BeZero())
		Expect(allns.Processes[pid].Name).To(Equal("pidxlas3.sh"))
	})

})
