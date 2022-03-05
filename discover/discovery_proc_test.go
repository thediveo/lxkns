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

package discover

import (
	"fmt"
	"os"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("Discover from processes", func() {

	It("finds at least the namespaces lsns finds", func() {
		allns := Namespaces(FromProcs())
		alllsns := lsns()
		ignoreme := regexp.MustCompile(`^(unshare|/bin/bash|runc) (.+ )?/tmp/`)
		for _, ns := range alllsns {
			nsidx := model.TypeIndex(species.NameToType(ns.Type))
			discons := allns.Namespaces[nsidx][species.NamespaceIDfromInode(ns.NS)]
			// Try to squash false positives, which are resulting from our own
			// test scripts...
			if discons == nil {
				if ignoreme.MatchString(ns.Command) {
					fmt.Fprintf(os.Stderr,
						"NOTE: skipping false positive: %s:[%d] %q\n",
						ns.Type, ns.NS, ns.Command)
					continue
				}
			}
			// And now for the real assertion!
			Expect(discons).NotTo(BeNil(), func() string {
				// Dump details of what lsns has seen, versus what lxkns has
				// discovered. This should help diagnosing problems ... such
				// as the spurious false positives due to test basher scripts
				// spinning up and down with some delay, so lsns and lxkns
				// might see different system states.
				lsns := ""
				for _, entry := range alllsns {
					lsns += fmt.Sprintf("\t%v\n", entry)
				}
				lxns := ""
				for nstype := model.NamespaceTypeIndex(0); nstype < model.NamespaceTypesCount; nstype++ {
					for _, ns := range allns.Namespaces[nstype] {
						lxns += fmt.Sprintf("\t%s\n", ns.String())
					}
				}
				return fmt.Sprintf("missing %s namespace %d\nlsns:\n%slxkns:\n%s", ns.Type, ns.NS, lsns, lxns)
			})
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
