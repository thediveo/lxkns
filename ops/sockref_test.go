// Copyright 2026 Harald Albrecht.
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
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/thediveo/spacetest"
	"github.com/thediveo/spacetest/netns"
	"golang.org/x/sys/unix"

	"github.com/thediveo/lxkns/ops/relations"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("socket fd references", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	DescribeTable("parsing procfs fd link paths",
		func(path string, expectedPID, expectedFd int, expectedOk bool) {
			pid, fdno, ok := parseProcFdPath(path)
			Expect(ok).To(Equal(expectedOk))
			Expect(pid).To(Equal(expectedPID))
			Expect(fdno).To(Equal(expectedFd))
		},
		Entry(nil, "/proc/1/fd/3", 1, 3, true),
		Entry(nil, "/proc/42/task/43/fd/0", 42, 0, true),
		Entry(nil, "/host/proc/42/fd/7", 42, 7, true),
		Entry(nil, "/proc/self/fd/3", 0, 0, false),
		Entry(nil, "/proc/thread-self/fd/3", 0, 0, false),
		Entry(nil, "/proc/1/ns/net", 0, 0, false),
		Entry(nil, "/proc/1/fd/x", 0, 0, false),
		Entry(nil, "/proc/1/fd/-1", 0, 0, false),
		Entry(nil, "/proc/0/fd/3", 0, 0, false),
		Entry(nil, "/proc/4294967296/fd/3", 0, 0, false),
		Entry(nil, "/proc/42/task/x/fd/3", 0, 0, false),
		Entry(nil, "fd/3", 0, 0, false),
	)

	It("keeps ENXIO for sockets outside procfs fd links", func() {
		tmpdir := Successful(os.MkdirTemp("", "lxkns-test-"))
		DeferCleanup(os.RemoveAll, tmpdir)
		l := Successful(net.Listen("unix", filepath.Join(tmpdir, "sock")))
		DeferCleanup(l.Close)

		Expect(NamespacePath(filepath.Join(tmpdir, "sock")).ID()).Error().To(MatchError(unix.ENXIO))
	})

	It("keeps ENXIO when the socket's network namespace is inaccessible", func() {
		if os.Getuid() == 0 {
			Skip("needs non-root")
		}
		sockfd := Successful(unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0))
		DeferCleanup(unix.Close, sockfd)

		Expect(NamespacePath(fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), sockfd)).ID()).Error().
			To(MatchError(unix.ENXIO))
	})

	When("a network namespace is only kept alive by a socket", func() {

		var sockfd int
		var netnsid species.NamespaceID

		BeforeEach(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			netnsfd := spacetest.NewUnmanagedTransient(unix.CLONE_NEWNET)
			netnsid = Successful(NamespaceFd(netnsfd).ID())
			netns.Execute(netnsfd, func() {
				sockfd = Successful(unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0))
			})
			DeferCleanup(unix.Close, sockfd)
			Expect(unix.Close(netnsfd)).To(Succeed())
		})

		It("queries the socket's network namespace", func() {
			nsp := NamespacePath(fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), sockfd))
			Expect(nsp.ID()).To(Equal(netnsid))
			Expect(nsp.Type()).To(Equal(species.CLONE_NEWNET))

			userns := Successful(nsp.User())
			Expect(userns.(io.Closer).Close()).To(Succeed())

			fd, closer := Successful2R(nsp.NsFd())
			Expect(NamespaceFd(fd).ID()).To(Equal(netnsid))
			closer()

			ref, refcloser := Successful2R(nsp.OpenTypedReference())
			Expect(ref.ID()).To(Equal(netnsid))
			refcloser()
		})

		DescribeTable("visits the socket's network namespace",
			func(newref func(path string) relations.Relation) {
				var visited species.NamespaceID
				Expect(Visit(func() {
					visited, _ = NamespacePath("/proc/thread-self/ns/net").ID()
				}, newref(fmt.Sprintf("/proc/%d/fd/%d", os.Getpid(), sockfd)))).To(Succeed())
				Expect(visited).To(Equal(netnsid))
			},
			Entry("NamespacePath", func(path string) relations.Relation {
				return NamespacePath(path)
			}),
			Entry("TypedNamespacePath", func(path string) relations.Relation {
				return NewTypedNamespacePath(path, species.CLONE_NEWNET)
			}),
		)

		It("queries the network namespace of another process's socket", func() {
			dupfd := Successful(unix.FcntlInt(uintptr(sockfd), unix.F_DUPFD_CLOEXEC, 0))
			sockf := os.NewFile(uintptr(dupfd), "socket")
			DeferCleanup(sockf.Close)

			sleepy := exec.Command("sleep", "600")
			sleepy.ExtraFiles = []*os.File{sockf} // becomes fd 3 in the child.
			Expect(sleepy.Start()).To(Succeed())
			DeferCleanup(func() {
				_ = sleepy.Process.Kill()
				_ = sleepy.Wait()
			})

			Expect(NamespacePath(fmt.Sprintf("/proc/%d/fd/3", sleepy.Process.Pid)).ID()).
				To(Equal(netnsid))
		})

	})

})
