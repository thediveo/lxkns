// Copyright 2021 Harald Albrecht.
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

package mountineer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
)

var _ = Describe("mountineer", func() {

	Context("basic functionality", func() {

		AfterEach(func() {
			Eventually(Goroutines).ShouldNot(HaveLeaked())
		})

		It("does not accept empty references", func() {
			Expect(New(nil, nil)).Error().To(HaveOccurred())
			Expect(New([]string{""}, nil)).Error().To(HaveOccurred())
			Expect(New([]string{"foobar"}, nil)).Error().To(HaveOccurred())
			Expect(New([]string{"/proc/self/ns/mnt", "/proc/self/ns/mnt"}, nil)).Error().To(HaveOccurred())
			Expect(New([]string{"/proc/self"}, nil)).Error().To(HaveOccurred())
			Expect(New([]string{"/proc/self/"}, nil)).Error().To(HaveOccurred())
		})

		It("resolves paths", func() {
			pid := os.Getpid()
			m, err := New([]string{fmt.Sprintf("/proc/%d/ns/mnt", pid)}, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(m.contentsRoot).To(Equal(fmt.Sprintf("/proc/%d/root", pid)))
			pwd, err := filepath.Abs("")
			Expect(err).NotTo(HaveOccurred())
			Expect(m.Resolve("")).To(Equal(fmt.Sprintf("/proc/%d/root%s", pid, pwd)))
		})

		It("opens", func() {
			pid := os.Getpid()
			m, err := New([]string{fmt.Sprintf("/proc/%d/ns/mnt", pid)}, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(m.PID()).To(Equal(model.PIDType(os.Getpid())))

			f, err := m.Open("mountineer_test.go")
			Expect(err).NotTo(HaveOccurred())
			f.Close()
			Expect(m.Open("foobar.go")).Error().To(HaveOccurred())
		})

	})

	Context("accessing bind-mounted mount namespace", Ordered, func() {

		bindmountpoint := "/tmp/lxkns-unittest-bindmountpoint"
		testdata := "/tmp/lxkns-unittest-data"
		canary := testdata + "/killroy.was.here"

		BeforeAll(func() {
			if os.Getegid() != 0 {
				// This unit test cannot be run inside a user namespace :(
				Skip("needs root")
			}

			// This test harness is admittedly involved: we create a new mount
			// namespace and then bind-mount it. Unfortunately, bind-mounting mount
			// namespaces is more restricted and tricky than bind-mounting a network
			// namespace, see also: https://unix.stackexchange.com/a/473819. And we
			// also want to be able to correctly identify the use of the proper
			// mount namespace by placing a file into it not visible from the mount
			// namespace the test runs in.
			scripts := testbasher.Basher{}
			defer scripts.Done()

			scripts.Common(fmt.Sprintf(`bm=%s
td=%s
canary=%s`, bindmountpoint, testdata, canary))
			scripts.Common(nstest.NamespaceUtilsScript)

			scripts.Script("main1", `
umount $bm || /bin/true
umount $bm || /bin/true
touch $bm
mount --bind $bm $bm
mount --make-private $bm

umount $td || /bin/true
umount $td || /bin/true
mkdir -p $td
mount --bind $td $td
mount --make-private $td

echo "\"\""

read PID # wait for test to proceed()
mount --bind /proc/$PID/ns/mnt $bm

echo "\"\""

read # wait for test to proceed()
umount $bm || /bin/true
umount $bm || /bin/true
rm $bm
umount $td || /bin/true
umount $td || /bin/true
rmdir $td
`)

			scripts.Script("main2", `
unshare -m $stage2
`)
			scripts.Script("stage2", `
mount -t tmpfs -o size=1m tmpfs $td
touch $canary
echo $$
process_namespaceid mnt # prints the "current" mount namespace ID.
read # wait for test to proceed()
`)

			By("creating a bind-mounted mount namespace")
			cmd := scripts.Start("main1")
			DeferCleanup(func() { cmd.Close() })

			var dummy string
			cmd.Decode(&dummy)

			By("creating a canary file inside the bind-mounted mount namespace")
			// create new mount namespace, mount a tmpfs into it and create the
			// canary file.
			cmd2 := scripts.Start("main2")
			defer cmd2.Close()
			var pid int
			cmd2.Decode(&pid)
			mntnsid := nstest.CmdDecodeNSId(cmd2)

			// tell the first script to bind-mount the new mount namespace.
			cmd.Tell(strconv.Itoa(pid))
			cmd.Decode(&dummy)

			// we don't need to keep the second script anymore, as the mount
			// namespace is now kept alive via the bind-mount. Note that we're
			// already defer'ed closing cmd2 anyway.

			By("checking the bind-mounted mount namespace test harness")
			// sanity check that the bind-mount points to the expected mount namespace.
			bmnsid, err := ops.NamespacePath(bindmountpoint).ID()
			Expect(err).NotTo(HaveOccurred())
			Expect(bmnsid).To(Equal(mntnsid))

			// canary must not be visible in this mount namespace
			Expect(canary).NotTo(Or(BeADirectory(), BeAnExistingFile()))
		})

		AfterAll(func() {
			// my oh my, we ARE pedantic today, innit?
			Eventually(Goroutines).ShouldNot(HaveLeaked())
		})

		When("using a mountineer", func() {

			var m *Mountineer

			BeforeAll(func() {
				// tell the mountineer to sandbox the newly created mount namespace via
				// the bind-mount reference.
				var err error
				m, err = New([]string{bindmountpoint}, nil)
				Expect(err).NotTo(HaveOccurred())
				DeferCleanup(func() { m.Close() })
			})

			It("created a sandbox/pause process that survives", func() {
				Expect(m.sandbox).NotTo(BeNil())
				// And the sandbox must not have terminated even if waiting a few
				// moments.
				Consistently(func() *os.ProcessState {
					return m.sandbox.ProcessState
				}).WithTimeout(3 * time.Second).WithPolling(250 * time.Millisecond).Should(BeNil())
			})

			It("sets the contentsroot to the sandbox process", func() {
				Expect(m.contentsRoot).To(Equal(
					fmt.Sprintf("/proc/%d/root", m.sandbox.Process.Pid)))
				Expect(m.PID()).To(Equal(model.PIDType(m.sandbox.Process.Pid)))
			})

			It("correctly resolves and opens a path", func() {
				path, err := m.Resolve(canary)
				Expect(err).NotTo(HaveOccurred())
				Expect(path).To(Equal(
					fmt.Sprintf("/proc/%d/root%s", m.sandbox.Process.Pid, canary)))

				f, err := os.Open(path)
				Expect(err).NotTo(HaveOccurred())
				f.Close()

				f, err = m.Open(canary)
				Expect(err).NotTo(HaveOccurred())
				f.Close()
			})

			It("shuts down correctly and doesn't leak sandboxes", func() {
				sandbox := m.sandbox
				m.Close()
				Eventually(func() *os.ProcessState {
					return sandbox.ProcessState
				}).WithTimeout(1 * time.Second).WithPolling(250 * time.Millisecond).ShouldNot(BeNil())
				Expect(m.sandbox).To(BeNil())
			})

		})

	})

})
