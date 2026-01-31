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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/safe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("renders branches", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		// As we're keeping a harness script running in the background we'll
		// have additionally "background" goroutines running that would
		// otherwise cause false positives, so we take a snapshot here.
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("CLI --foobar fails correctly", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--foobar"})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: unknown flag: --foobar`))
	})

	It("CLI rejects invalid target namespaces", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"foo:[666]"})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: not a valid namespace:`))
	})

	It("CLI rejects invalid --ns", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--ns", "net:[666]", "net:[12345678]"})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: not a valid PID namespace:`))
	})

	It("CLI rejects valid --ns ID without --pid", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--ns", "666", "net:[12345678]"})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: --ns requires --pid`))
	})

	It("CLI rejects non-existing --ns ID", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"--ns", "666",
			"--pid", "666",
			"--" + turtles.NoContainersFlagName,
			"net:[12345678]",
		})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: unknown PID namespace`))
	})

	It("CLI rejects non-existing PID", func() {
		mypidns, err := ops.NamespacePath("/proc/self/ns/pid").ID()
		Expect(err).ToNot(HaveOccurred())

		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"--ns", fmt.Sprintf("%d", mypidns.Ino),
			"--pid", fmt.Sprintf("%d", ^uint32(0)),
			"--" + turtles.NoContainersFlagName,
			"net:[12345678]",
		})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: unknown process PID .* in`))

		cmd.SetArgs([]string{
			"--pid", fmt.Sprintf("%d", ^uint32(0)),
			"--" + turtles.NoContainersFlagName,
			"net:[12345678]",
		})
		var out2 safe.Buffer
		cmd.SetOut(&out2)
		cmd.SetErr(&out2)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: unknown process PID .*`))
	})

	It("CLI rejects non-existing target namespace", func() {
		mypidns, err := ops.NamespacePath("/proc/self/ns/pid").ID()
		Expect(err).ToNot(HaveOccurred())

		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"--ns", fmt.Sprintf("%d", mypidns.Ino),
			"--pid", fmt.Sprintf("%d", os.Getpid()),
			"--" + turtles.NoContainersFlagName,
			"net:[12345678]",
		})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: unknown namespace net:`))
	})

	It("CLI w/o args fails", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())
		Expect(out.String()).To(MatchRegexp(`^Error: expects 1 arg, received 0`))
	})

	It("CLI with target non-user namespace below process in owned user namespace", func() {
		if os.Geteuid() == 0 {
			Skip("only non-root")
		}

		mynetnsid, err := ops.NamespacePath("/proc/self/ns/net").ID()
		Expect(err).NotTo(HaveOccurred())

		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"--" + turtles.NoContainersFlagName,
			fmt.Sprintf("net:[%d]", mynetnsid.Ino),
		})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(fmt.Sprintf(`(?m)^⛛ user:\[%d\] process .*
├─ process .*
│     ⋄─ \(no effective capabilities\)
└─ target net:\[%d\] process .*
      ⋄─ \(no effective capabilities\)$`,
			initialUsernsID.Ino, mynetnsid.Ino)))
	})

	It("CLI with target non-user namespace below process in owned user namespace", func() {
		if os.Geteuid() == 0 {
			Skip("only non-root")
		}

		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"--" + turtles.NoContainersFlagName,
			fmt.Sprintf("net:[%d]", targetNetnsID.Ino),
		})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(fmt.Sprintf(`(?m)^⛛ user:\[%d\] process .*
├─ process .*
│     ⋄─ \(no effective capabilities\)
└─ ✓ user:\[%d\] process .*
   └─ target net:\[%d\] (referenced from )?(process|task|process/task) .*
         ⋄─ cap_audit_control .*$`,
			initialUsernsID.Ino, targetUsernsID.Ino, targetNetnsID.Ino)))
	})

	It("CLI with target non-user namespace at process", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"-p", fmt.Sprintf("%d", targetPID),
			"--" + turtles.NoContainersFlagName,
			fmt.Sprintf("net:[%d]", targetNetnsID.Ino),
		})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(fmt.Sprintf(`(?m)^⛔ user:\[%d\] process .*
└─ ⛛ user:\[%d\] process .*
   ├─ process .*
   │     ⋄─ cap_audit_control .*
(   │     ⋄─ .*
)*   └─ target net:\[%d\] (referenced from )?(process|task|process/task) .*
         ⋄─ cap_audit_control .*$`,
			initialUsernsID.Ino, targetUsernsID.Ino, targetNetnsID.Ino)))
	})

	It("CLI with process in other user namespace branch than target non-user namespace", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{
			"-p", fmt.Sprintf("%d", someProcPID),
			"--" + turtles.NoContainersFlagName,
			fmt.Sprintf("net:[%d]", targetNetnsID.Ino),
		})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(fmt.Sprintf(`(?m)^⛔ user:\[%d\] process .*
├─ ⛛ user:\[%d\] process .*
│  └─ process .*
│        ⋄─ cap_audit_control .*
(│        ⋄─ .*
)*└─ ⛔ user:\[%d\] process .*
   └─ target net:\[%d\] (referenced from )?(process|task|process/task) .*
         ⋄─ \(no capabilities\)$`,
			initialUsernsID.Ino, someProcUsernsID.Ino, targetUsernsID.Ino, targetNetnsID.Ino)))
	})

})
