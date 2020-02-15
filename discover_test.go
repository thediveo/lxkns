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
	"fmt"

	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/nstypes"
	t "github.com/thediveo/lxkns/nstypes"
	r "github.com/thediveo/lxkns/relations"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Discover", func() {

	It("discovers no namespaces unless told so", func() {
		allns := Discover(NoDiscovery)
		for _, nsmap := range allns.Namespaces {
			Expect(nsmap).To(HaveLen(0))
		}
	})

	It("sorts namespaces", func() {
		nsmap := NamespaceMap{
			5678: NewNamespace(nstypes.CLONE_NEWNET, 5678, ""),
			1234: NewNamespace(nstypes.CLONE_NEWNET, 1234, ""),
		}
		dr := DiscoveryResult{}
		dr.Namespaces[NetNS] = nsmap
		sortedns := dr.SortedNamespaces(NetNS)
		Expect(sortedns).To(HaveLen(2))
		Expect(sortedns[0].ID()).To(Equal(nstypes.NamespaceID(1234)))
		Expect(sortedns[1].ID()).To(Equal(nstypes.NamespaceID(5678)))
	})

	It("finds at least the namespaces lsns finds", func() {
		allns := Discover(FullDiscovery)
		for _, ns := range lsns() {
			nsidx := TypeIndex(t.NameToType(ns.Type))
			discons := allns.Namespaces[nsidx][ns.NS]
			Expect(discons).NotTo(BeZero())
			// rats ... lsns seems to take the numerically lowest PID number
			// instead of the topmost PID in a namespace. This makes
			// Expect(dns.LeaderPIDs()).To(ContainElement(PIDType(ns.PID))) to
			// give false negatives, so we need to check the processes along
			// the hierarchy which are still in the same namespace to be
			// tested for.
			p, ok := allns.Processes[ns.PID]
			Expect(ok).To(BeTrue(), "unknown PID %d", ns.PID)
			leaders := discons.LeaderPIDs()
			func() {
				pids := []PIDType{}
				for p != nil {
					pids = append(pids, p.PID)
					for _, lPID := range leaders {
						if lPID == p.PID {
							return
						}
					}
					p = p.Parent
				}
				Fail(fmt.Sprintf("PIDs %v not found in leaders %v", pids, leaders))
			}()
		}
	})

	It("finds hidden hierarchical user namespaces", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Ur unshare -U $stage2 # create a user ns with another user ns inside.
`)
		scripts.Script("stage2", `
process_namespaceid user # prints the user namespace ID of "the" process.
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var usernsid t.NamespaceID
		cmd.Decode(&usernsid)
		allns := Discover(FullDiscovery)
		userns := allns.Namespaces[UserNS][usernsid].(Hierarchy)
		Expect(userns).NotTo(BeNil())
		ppusernsid, _ := r.ID("/proc/self/ns/user")
		Expect(userns.Parent().Parent().(Namespace).ID()).To(Equal(ppusernsid))
	})

	It("finds fd-referenced namespaces", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Urn $stage2 # set up the stage with a new user ns.
`)
		scripts.Script("stage2", `
process_namespaceid net # print ID of first new net ns.
exec unshare -n 3</proc/self/ns/net $stage3 # fd-ref net ns and then replace our shell.
`)
		scripts.Script("stage3", `
process_namespaceid net # print ID of second new net ns.
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var fdnetnsid, netnsid t.NamespaceID
		cmd.Decode(&fdnetnsid)
		cmd.Decode(&netnsid)
		Expect(fdnetnsid).ToNot(Equal(netnsid))
		// correctly misses fd-referenced namespaces without proper discovery
		// method enabled.
		opts := NoDiscovery
		opts.SkipProcs = false
		allns := Discover(opts)
		Expect(allns.Namespaces[NetNS]).To(HaveKey(netnsid))
		Expect(allns.Namespaces[NetNS]).ToNot(HaveKey(fdnetnsid))
		// correctly finds fd-referenced namespaces now.
		opts = NoDiscovery
		opts.SkipFds = false
		allns = Discover(opts)
		Expect(allns.Namespaces[NetNS]).To(HaveKey(fdnetnsid))
	})

	It("rejects finding roots for plain namespaces", func() {
		// We only need to run a simplified discovery on processes, but
		// nothing else.
		opts := NoDiscovery
		opts.SkipProcs = false
		opts.NamespaceTypes = t.CLONE_NEWNET
		allns := Discover(opts)
		Expect(func() { rootNamespaces(allns.Namespaces[NetNS]) }).To(Panic())
	})

})
