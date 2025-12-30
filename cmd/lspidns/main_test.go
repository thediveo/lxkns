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
	"context"
	"fmt"
	"time"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/safe"
	"github.com/thediveo/spacetest"
	"github.com/thediveo/spacetest/spacer"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("renders pid namespaces", func() {

	var initusernsid, initpidnsid, usernsid, pidnsid species.NamespaceID

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		initusernsid = species.NamespaceIDfromInode(spacetest.CurrentIno(unix.CLONE_NEWUSER))
		initpidnsid = species.NamespaceIDfromInode(spacetest.CurrentIno(unix.CLONE_NEWPID))

		By("creating user and PID child namespaces")
		ctx, cancel := context.WithCancel(context.Background())
		spcclnt := spacer.New(ctx, spacer.WithErr(GinkgoWriter))
		DeferCleanup(func() {
			cancel()
			spcclnt.Close()
		})

		subclnt, subspc := spcclnt.Subspace(true, true)
		DeferCleanup(func() {
			_ = unix.Close(subspc.PID)
			_ = unix.Close(subspc.User)
			subclnt.Close()
		})

		usernsid = species.NamespaceIDfromInode(spacetest.Ino(subspc.User, unix.CLONE_NEWUSER))
		pidnsid = species.NamespaceIDfromInode(spacetest.Ino(subspc.PID, unix.CLONE_NEWPID))
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

	It("CLI w/o args renders pid tree", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		output := out.String()

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^pid:\[%d\] process .*$`,
			initpidnsid.Ino)))
		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^[├└]─ pid:\[%d\] process .*$`,
			pidnsid.Ino)))
	})

	It("CLI -u renders user/pid tree", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"-u", "--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		output := out.String()

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] process .*
[├└]─ pid:\[%d\] process .*$`,
			initusernsid.Ino, initpidnsid.Ino)))
		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^   [├└]─ user:\[%d\] process .*
   [│ ]  [├└]─ pid:\[%d\] process .*$`,
			usernsid.Ino, pidnsid.Ino)))
	})

})
