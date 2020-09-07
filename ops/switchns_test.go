// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ops

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"
)

var _ = Describe("Set Namespaces", func() {

	It("Go()es with errors", func() {
		err := Go(func() {}, NamespacePath("foobar"))
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(MatchRegexp(`cannot reference namespace, .+invalid namespace path foobar`)))
	})

	It("Go()es with errors as non-root", func() {
		if os.Geteuid() == 0 {
			Skip("don't be roode.")
		}

		err := Go(func() {}, NamespacePath("/proc/1/ns/pid"))
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(MatchRegexp(`cannot reference namespace, .+invalid namespace path`)))

		err = Go(func() {}, NamespacePath("/proc/self/ns/pid"))
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(MatchRegexp(`cannot enter namespace path .+, operation not permitted`)))
	})

	It("Execute()s with errors", func() {
		_, err := Execute(func() interface{} { return nil },
			NamespacePath("foobar"))
		Expect(err).To(HaveOccurred())
	})

	It("Visit()s with errors", func() {
		err := Visit(func() {}, NamespacePath("foobar"))
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(MatchRegexp(`cannot reference namespace, .+invalid namespace path foobar`)))

		err = Visit(func() {}, NamespacePath("doc.go"))
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(MatchRegexp(`cannot determine type.+ioctl`)))

		err = Visit(func() {}, NewTypedNamespacePath("/proc/self/ns/net", 0))
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(MatchRegexp(`cannot determine type.+ioctl`)))
	})

	It("Visit()s with errors, part 2", func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		netns := NamespacePath("/proc/self/ns/net")
		_, closer, err := netns.Reference()
		Expect(err).NotTo(HaveOccurred())
		err = Visit(
			func() { closer() },
			netns)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(MatchRegexp(`hmpf`)))
	})

	It("Execute()s", func() {
		Expect(Execute(func() interface{} { return nil })).NotTo(HaveOccurred())
	})

	It("Go()es into other namespaces", func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -n $stage2
`)
		scripts.Script("stage2", `
echo "\"/proc/$$/ns/net\""
process_namespaceid net
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()

		var netnsref NamespacePath
		var netnsid species.NamespaceID
		cmd.Decode(&netnsref)
		cmd.Decode(&netnsid)

		result := make(chan species.NamespaceID)
		Expect(Go(func() {
			id, _ := NamespacePath(
				fmt.Sprintf("/proc/%d/ns/net", unix.Gettid())).
				ID()
			result <- id
		}, netnsref)).NotTo(HaveOccurred())
		Expect(<-result).To(Equal(netnsid))

		res, err := Execute(func() interface{} {
			id, _ := NamespacePath(
				fmt.Sprintf("/proc/%d/ns/net", unix.Gettid())).
				ID()
			return id
		}, netnsref)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.(species.NamespaceID)).To(Equal(netnsid))
	})

	It("Visit()s other namespaces and then returns", func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -n $stage2
`)
		scripts.Script("stage2", `
echo "\"/proc/$$/ns/net\""
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()

		var netnsref NamespacePath
		cmd.Decode(&netnsref)
		initID, err := netnsref.ID()
		Expect(err).ToNot(HaveOccurred())

		var beforeID, visitedID, afterID species.NamespaceID
		done := make(chan struct{})
		var locked bool
		// Don't do the Visit on the main go routine, mate!
		go func() {
			defer close(done)

			// Record the network namespace the process is in (don't care about
			// the specific OS thread here, as it isn't locked yet anyway).
			beforeID, err = NamespacePath("/proc/self/ns/net").ID()
			if err != nil {
				return
			}
			var innererr error
			err = Visit(func() {
				// We now should be switched into the new network namespace, but
				// only this locked OS thread is switched. Record the current
				// network namespace so we can later check that the OS thread
				// had switched into the correct network namespace.
				visitedID, innererr = NamespacePath(
					fmt.Sprintf("/proc/%d/ns/net", unix.Gettid())).ID()
			}, netnsref)
			if innererr != nil {
				err = innererr
				return
			}
			if err != nil {
				return
			}
			// Find out whether the OS thread has been correctly unlocked, ...
			// or not; this is an ugly hack, as there is no official API, so we
			// check what stack trace is going to tell us...
			locked = strings.Contains(string(debug.Stack()), ", locked to thread]")
			// Finally record the network namespace after the visit; we'll later
			// check that we're back in the process' network namespace.
			afterID, err = NamespacePath("/proc/self/ns/net").ID()
		}()
		// Wait for Visit to complete on separate go routine with a throw-away
		// OS thread.
		<-done

		Expect(err).ToNot(HaveOccurred())
		Expect(visitedID).ToNot(Equal(beforeID), "didn't switch network namespace")
		Expect(visitedID).To(Equal(initID), "switched into what???")
		Expect(afterID).To(Equal(beforeID), "didn't switch back")
		Expect(locked).To(BeFalse(), "didn't unlock OS thread")
	})

})
