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
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops/internal/opener"
	"github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

type brokenref struct{ NamespacePath }

func (b *brokenref) OpenTypedReference() (relations.Relation, opener.ReferenceCloser, error) {
	return b, func() {}, nil
}

func (b brokenref) NsFd() (int, opener.FdCloser, error) {
	return 0, nil, errors.New("broken reference")
}

var _ opener.Opener = (*brokenref)(nil)

var _ = Describe("Set Namespaces", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	DescribeTable("describes the error when switching or restoring a namespace",
		func(err error, msg string, as interface{}) {
			Expect(err.Error()).To(Equal(msg))
			Expect(errors.As(err, &as)).To(BeTrue())
			Expect(as.(error).Error()).To(Equal(msg))
		},
		Entry("switching error",
			&SwitchNamespaceErr{msg: "foo"}, "foo", &RestoreNamespaceErr{}),
		Entry("restoring error",
			&RestoreNamespaceErr{msg: "bar"}, "bar", &RestoreNamespaceErr{}),
	)

	It("Go()es with errors", func() {
		Expect(Go(func() {}, NamespacePath("foobar"))).Error().To(
			MatchError(MatchRegexp(`cannot reference namespace, .+invalid namespace path "foobar"`)))
	})

	It("Go()es with errors as non-root", func() {
		if os.Geteuid() == 0 {
			Skip("don't be roode.")
		}

		Expect(Go(func() {}, NamespacePath("/proc/1/ns/pid"))).Error().To(
			MatchError(MatchRegexp(`cannot reference namespace, .+invalid namespace path`)))

		Expect(Go(func() {}, NamespacePath("/proc/self/ns/pid"))).Error().To(
			MatchError(MatchRegexp(`cannot enter namespace path .+, operation not permitted`)))
	})

	It("Execute()s with errors", func() {
		Expect(Execute(func() interface{} { return nil }, NamespacePath("foobar"))).Error().To(HaveOccurred())
	})

	DescribeTable("Visit()s with errors when attempting to use...",
		func(nsref relations.Relation, expected string) {
			// Run the visitation on a separate locked goroutine in order to not
			// lock the current Ginkgo goroutine due to errors.
			ch := make(chan error)
			go func() {
				ch <- Visit(func() {}, nsref)
			}()
			Eventually(ch).WithTimeout(2 * time.Second).WithPolling(100 * time.Millisecond).
				Should(Receive(MatchError(MatchRegexp(expected))))
		},
		Entry("non-existing file as namespace reference",
			NamespacePath("foobar"),
			`cannot reference namespace, .+invalid namespace path "foobar"`),
		Entry("existing ordinary file as namespace reference",
			NamespacePath("doc.go"),
			`cannot reference namespace.+NS_GET_NSTYPE.+ioctl`),
		Entry("unspecified typed reference",
			NewTypedNamespacePath("/proc/self/ns/net", ^species.NamespaceType(0)),
			`cannot determine type`),
		Entry("non-enterable namespace",
			NamespacePath("/proc/self/ns/mnt"),
			`cannot enter namespace, (operation not permitted|invalid argument)`),
		Entry("broken reference",
			&brokenref{NamespacePath("/proc/self/ns/net")},
			`cannot reference namespace, broken reference`),
	)

	It("Execute()s", func() {
		Expect(Execute(func() interface{} { return nil })).To(Succeed())
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
		cmd.Decode(&netnsref)
		netnsid := nstest.CmdDecodeNSId(cmd)

		result := make(chan species.NamespaceID)
		Expect(Go(func() {
			id, _ := NamespacePath(
				fmt.Sprintf("/proc/%d/ns/net", unix.Gettid())).
				ID()
			result <- id
		}, netnsref)).To(Succeed())
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
			afterID, err = NamespacePath(fmt.Sprintf("/proc/%d/ns/net", unix.Gettid())).ID()
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
