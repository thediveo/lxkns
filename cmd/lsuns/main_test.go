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
	"os"
	"runtime"
	"time"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/safe"
	"github.com/thediveo/spacetest"
	"github.com/thediveo/spacetest/spacer"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("renders user namespaces", func() {

	var ourUsernsID, childUsernsID, ownedNetnsID uint64 // no dev ID necessary, just the ino's

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).WithPolling(100 * time.Millisecond).
				ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		ourUsernsID = spacetest.CurrentIno(unix.CLONE_NEWUSER)

		By("spinning up a local spacer service")
		ctx, cancel := context.WithCancel(context.Background())
		spacerClient := spacer.New(ctx, spacer.WithErr(GinkgoWriter))
		DeferCleanup(func() {
			cancel()
			spacerClient.Close()
		})

		By("creating a child user namespace")
		subspaceClient, subspc := spacerClient.Subspace(true, false)
		DeferCleanup(func() {
			subspaceClient.Close()
		})

		childUsernsID = spacetest.Ino(subspc.User, unix.CLONE_NEWUSER)

		By("creating a network namespace owned by child user namespace")
		netnsfd := subspaceClient.NewTransient(unix.CLONE_NEWNET)
		ownedNetnsID = spacetest.Ino(netnsfd, unix.CLONE_NEWNET)
	})

	It("fails for unknown CLI flag", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--foobar"})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).NotTo(Succeed())

		Expect(out.String()).To(MatchRegexp(`^Error: unknown flag: --foobar`))
	})

	It("renders just the user tree without any CLI args", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)
		Expect(cmd.Execute()).To(Succeed())
		output := out.String()

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			ourUsernsID)))

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^[├└]─ user:\[%d\] .*$`,
			childUsernsID)))
	})

	It("renders user tree with owned namespaces with CLI -d", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"-d", "--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		output := out.String()

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			ourUsernsID)))
		Expect(output).To(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ mnt:.*
[│ ]+⋄─ net:\[%d\] (referenced from )?(process|task|process/task) .*$`,
			childUsernsID, ownedNetnsID)))
	})

	It("renders user tree with loose threads with CLI -d", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		By("creating a stray task with its own namespace...")
		tidch := make(chan int)
		done := make(chan struct{})
		defer close(done)
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

		By("discovering from processes and tasks")
		cmd := newRootCmd()
		cmd.SetArgs([]string{"-d", "--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		output := out.String()

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			ourUsernsID)))
		Expect(output).To(MatchRegexp(fmt.Sprintf(`
[│ ]+⋄─ mnt:\[.*\] task ".*" \[%d\] of ".*" \(%d\)`,
			tid, os.Getpid())))

		By("discovering from processes only")
		cmd = newRootCmd()
		cmd.SetArgs([]string{"-d", "--task=false", "--" + turtles.NoContainersFlagName})
		var out2 safe.Buffer
		cmd.SetOut(&out2)
		cmd.SetErr(&out2)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		output = out2.String()

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			ourUsernsID)))
		Expect(output).NotTo(MatchRegexp(fmt.Sprintf(`
[│ ]+⋄─ mnt:\[.*\] task ".*" \[%d\] of ".*" \(%d\)`,
			tid, os.Getpid())))
	})

	It("filters owned namespaces with CLI -f", func() {
		cmd := newRootCmd()
		cmd.SetArgs([]string{"-d", "-f=pid", "--" + turtles.NoContainersFlagName})
		var out safe.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		output := out.String()

		Expect(output).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			ourUsernsID)))
		Expect(output).ToNot(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ .*$`,
			childUsernsID)))

		cmd = newRootCmd()
		cmd.SetArgs([]string{"-d", "-f=ipc,net,pid", "--" + turtles.NoContainersFlagName})
		var out2 safe.Buffer
		cmd.SetOut(&out2)
		cmd.SetErr(&out2)
		debug.SetWriter(cmd, GinkgoWriter)

		Expect(cmd.Execute()).To(Succeed())
		output = out2.String()
		Expect(output).To(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ net:\[%d\] (referenced from )?(process|task|process/task) .*$`,
			childUsernsID, ownedNetnsID)))
	})

})
