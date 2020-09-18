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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("Discover from processes", func() {

	It("finds at least the namespaces lsns finds", func() {
		opts := NoDiscovery
		opts.SkipProcs = false
		allns := Discover(opts)
		for _, ns := range lsns() {
			nsidx := model.TypeIndex(species.NameToType(ns.Type))
			discons := allns.Namespaces[nsidx][species.NamespaceIDfromInode(ns.NS)]
			Expect(discons).NotTo(BeNil(),
				"missing %s namespace %d", ns.Type, ns.NS)
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
				pids := []model.PIDType{}
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

})
