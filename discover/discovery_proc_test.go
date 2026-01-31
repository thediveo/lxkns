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
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("Discover from processes", Ordered, func() {

	BeforeEach(func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))
	})

	It("finds at least the namespaces lsns finds", func() {
		// hear, hear ... lsns finally upped its game :D
		allns := Namespaces(FromProcs(), FromBindmounts())
		alllsns := lsns()
		ignoreme := regexp.MustCompile(`^(unshare|/bin/bash|runc) (.+ )?/tmp/`)
		for _, lsns := range alllsns {
			nsidx := model.TypeIndex(species.NameToType(lsns.Type))
			discons := allns.Namespaces[nsidx][species.NamespaceIDfromInode(lsns.NS)]
			// Try to squash false positives, which are resulting from our own
			// test scripts...
			if discons == nil {
				if ignoreme.MatchString(lsns.Command) {
					fmt.Fprintf(os.Stderr,
						"NOTE: skipping false positive: %s:[%d] %q\n",
						lsns.Type, lsns.NS, lsns.Command)
					continue
				}
			}
			// And now for the real assertion!
			Expect(discons).NotTo(BeNil(), func() string {
				// Dump details of what lsns has seen, versus what lxkns has
				// discovered. This should help diagnosing problems ... such as
				// the spurious false positives due to test basher scripts
				// spinning up and down with some delay, so lsnslist and lxkns
				// might see different system states.
				var lsnslist strings.Builder
				for _, entry := range alllsns {
					fmt.Fprintf(&lsnslist, "\t%v\n", entry)
				}
				var lxnslist strings.Builder
				for nstype := range model.NamespaceTypesCount {
					for _, ns := range allns.Namespaces[nstype] {
						fmt.Fprintf(&lxnslist, "\t%s\n", ns.String())
					}
				}
				return fmt.Sprintf("missing %s namespace %d\nlsns:\n%slxkns:\n%s", lsns.Type, lsns.NS, lsnslist.String(), lxnslist.String())
			})
			// As of lsns util-linux 2.39.1 we now get bind-mounted namespaces
			// as well, so we need to cover this case especially.
			if lsns.PID == 0 {
				tidx, ok := model.NamespaceTypeIndexByName(lsns.Type)
				Expect(ok).To(BeTrue(), "unknown namespace type %s", lsns.Type)
				Expect(allns.Namespaces[tidx]).To(HaveKey(species.NamespaceIDfromInode(lsns.NS)))
				continue
			}
			// rats ... lsns seems to take the numerically lowest PID number
			// instead of the topmost PID in a namespace. This makes
			// Expect(dns.LeaderPIDs()).To(ContainElement(PIDType(ns.PID))) to
			// give false negatives, so we need to check the processes along
			// the hierarchy which are still in the same namespace to be
			// tested for.
			p, ok := allns.Processes[lsns.PID]
			Expect(ok).To(BeTrue(), "unknown PID %d", lsns.PID)
			leaders := discons.LeaderPIDs()
			func() {
				pids := []model.PIDType{}
				for p != nil {
					pids = append(pids, p.PID)
					if slices.Contains(leaders, p.PID) {
						return
					}
					p = p.Parent
				}
				Fail(fmt.Sprintf("PIDs %v not found in leaders %v", pids, leaders))
			}()
		}
	})

	It("finds a task-held namespace", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		By("setting up a stray task with its own namespace...")
		tidch := make(chan int)
		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			runtime.LockOSThread() // never unlock, as this task is going to be tainted.

			// I owe Micheal Kerrisk several beers for opening my eyes to this
			// twist: a task can create its own new mount namespace after it has
			// declared itself independent of the effects of CLONE_FS when it
			// was created as a task (=thread) inside a process. And yes, this
			// allows the mountineers to work without the separate pause process
			// and instead using a throw-away thread/task.
			Expect(unix.Unshare(unix.CLONE_FS | unix.CLONE_NEWNS)).To(Succeed())

			tidch <- unix.Gettid()
			<-done
		}()

		var tid int
		Eventually(tidch).Should(Receive(&tid))
		defer close(done)

		By("scanning all processes and tasks...")
		allns := Namespaces(FromProcs(), FromTasks())
		var tasknetns model.Namespace
		Expect(allns.Namespaces[model.MountNS]).To(ContainElement(
			HaveField("LooseThreadIDs()", ConsistOf(model.PIDType(tid))), &tasknetns))
		loose := tasknetns.LooseThreads()
		Expect(loose).To(HaveLen(1))
		Expect(loose[0].Namespaces[model.MountNS].ID()).NotTo(Equal(
			loose[0].Process.Namespaces[model.MountNS].ID()))

		By("scanning only processes")
		allns = Namespaces(FromProcs())
		Expect(allns.Namespaces[model.MountNS]).NotTo(ContainElement(
			HaveField("LooseThreadIDs()", ConsistOf(model.PIDType(tid)))))
	})

})
