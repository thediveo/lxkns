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
	"runtime"
	"time"

	"github.com/thediveo/lxkns/cmd/internal/test/getstdout"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("renders user namespaces", func() {

	var initusernsid, usernsid, netnsid species.NamespaceID

	BeforeEach(func() {
		osExit = func(int) {}
		DeferCleanup(func() { osExit = os.Exit })
	})

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		scripts := testbasher.Basher{}
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
process_namespaceid user
unshare -Unfr $stage2
`)
		scripts.Script("stage2", `
process_namespaceid user
process_namespaceid net
read
`)
		cmd := scripts.Start("main")
		initusernsid = nstest.CmdDecodeNSId(cmd)
		usernsid = nstest.CmdDecodeNSId(cmd)
		netnsid = nstest.CmdDecodeNSId(cmd)

		DeferCleanup(func() {
			if cmd != nil {
				cmd.Close()
			}
			scripts.Done()
		})
	})

	It("fails for unknown CLI flag", func() {
		oldExit := osExit
		defer func() { osExit = oldExit }()
		exit := 0
		osExit = func(code int) { exit = code }
		os.Args = append(os.Args[:1], "--foobar")
		out := getstdout.Stdouterr(main)
		Expect(exit).To(Equal(1))
		Expect(out).To(MatchRegexp(`^Error: unknown flag: --foobar`))
	})

	It("renders just the user tree without any CLI args", func() {
		os.Args = os.Args[:1]
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^[├└]─ user:\[%d\] .*$`,
			usernsid.Ino)))
	})

	It("renders user tree with owned namespaces with CLI -d", func() {
		os.Args = append(os.Args[:1], "-d")
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).To(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ net:\[%d\] process .*$`,
			usernsid.Ino, netnsid.Ino)))
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
		os.Args = append(os.Args[:1], "-d")
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).To(MatchRegexp(fmt.Sprintf(`
[│ ]+⋄─ mnt:\[.*\] task ".*" \[%d\] of ".*" \(%d\)`,
			tid, os.Getpid())))

		By("discovering from processes only")
		os.Args = append(os.Args[:1], "-d", "--task=false")
		out = getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).NotTo(MatchRegexp(fmt.Sprintf(`
[│ ]+⋄─ mnt:\[.*\] task ".*" \[%d\] of ".*" \(%d\)`,
			tid, os.Getpid())))
	})

	It("filters owned namespaces with CLI -f", func() {
		os.Args = append(os.Args[:1], "-d", "-f=pid")
		out := getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`(?m)^user:\[%d\] .*$`,
			initusernsid.Ino)))
		Expect(out).ToNot(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ .*$`,
			usernsid.Ino)))

		os.Args = append(os.Args[:1], "-d", "-f=ipc,net,pid")
		out = getstdout.Stdouterr(main)
		Expect(out).To(MatchRegexp(fmt.Sprintf(`
(?m)^[├└]─ user:\[%d\] process .*
[│ ]+⋄─ net:\[%d\] process .*$`,
			usernsid.Ino, netnsid.Ino)))
	})

})
