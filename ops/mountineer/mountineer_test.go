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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/testbasher"
)

var _ = Describe("mountineer", func() {

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

	It("opens a mount namespace path in initial context", func() {
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

		bm := "/tmp/lxkns-unittest-bindmountpoint"
		td := "/tmp/lxkns-unittest-data"
		canary := td + "/killroy.was.here"

		scripts.Common(fmt.Sprintf(`bm=%s
td=%s
canary=%s`, bm, td, canary))
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
		cmd := scripts.Start("main1")
		defer cmd.Close() // ...just in case of slight panic
		var dummy string
		cmd.Decode(&dummy)

		// create new mount namespace, mount a tmpfs into it and create the
		// canary file.
		cmd2 := scripts.Start("main2")
		defer cmd.Close()
		var pid int
		cmd2.Decode(&pid)
		mntnsid := nstest.CmdDecodeNSId(cmd2)

		// tell the first script to bind-mount the new mount namespace.
		cmd.Tell(strconv.Itoa(pid))
		cmd.Decode(&dummy)

		// we don't need to keep the second script anymore, as the mount
		// namespace is now kept alive via the bind-mount.
		cmd2.Close()

		// sanity check that the bind-mount points to the expected mount namespace.
		bmnsid, err := ops.NamespacePath(bm).ID()
		Expect(err).NotTo(HaveOccurred())
		Expect(bmnsid).To(Equal(mntnsid))

		// tell the mountineer to sandbox the newly created mount namespace via
		// the bind-mount reference.
		m, err := New([]string{bm}, nil)
		Expect(err).NotTo(HaveOccurred())
		defer m.Close()

		// canary must not be visible in this mount namespace
		Expect(canary).NotTo(Or(BeADirectory(), BeAnExistingFile()))

		// It must have created a sandbox/pause process.
		Expect(m.sandbox).NotTo(BeNil())
		// And the sandbox must not have terminated even if waiting a few
		// moments.
		Consistently(func() *os.ProcessState {
			return m.sandbox.ProcessState
		}, "1s", "250ms").Should(BeNil())
		// the contentsroot must be set to the sandbox process.
		Expect(m.contentsRoot).To(Equal(
			fmt.Sprintf("/proc/%d/root", m.sandbox.Process.Pid)))

		Expect(m.PID()).To(Equal(model.PIDType(m.sandbox.Process.Pid)))

		// Content path resolution must be correct.
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

		// Correctly stops the sandbox process -- no sandbox leaks, please.
		m.Close()
		Eventually(func() *os.ProcessState {
			return m.sandbox.ProcessState
		}, "1s", "250ms").ShouldNot(BeNil())
	})

})
