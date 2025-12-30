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
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	nonetns "github.com/thediveo/notwork/netns"
	"github.com/thediveo/spacetest/netns"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("Discover from fds", func() {

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

	It("finds fd-referenced namespaces", func() {
		By("creating a transient network namespace and keeping only an open fd to it")
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Urn $stage2 # set up the stage with a new user ns.
`)
		scripts.Script("stage2", `
process_namespaceid net # print ID of first new net ns.
exec unshare -n 3</proc/self/ns/net $stage3 # fd-ref net ns and then replace our shell.
`)
		scripts.Script("stage3", `
process_namespaceid net # print ID of second new net ns.
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		fdnetnsid := nstest.CmdDecodeNSId(cmd)
		netnsid := nstest.CmdDecodeNSId(cmd)
		Expect(fdnetnsid).ToNot(Equal(netnsid))

		By("missing fd referenced namespaces when discovering only from processes")
		allns := Namespaces(FromProcs())
		Expect(allns.Namespaces[model.NetNS]).To(HaveKey(netnsid))
		Expect(allns.Namespaces[model.NetNS]).ToNot(HaveKey(fdnetnsid))

		By("finding fd referenced namespaces")
		allns = Namespaces(FromFds())
		Expect(allns.Namespaces[model.NetNS]).To(HaveKey(fdnetnsid))
	})

	It("skips /proc/*/fd/* nonsense", func() {
		var stat unix.Stat_t
		Expect(unix.Stat("./test/fdscan/proc", &stat)).ToNot(HaveOccurred())
		Expect(stat.Dev).NotTo(BeZero())
		r := Result{
			Options: DiscoverOpts{
				ScanFds: true,
			},
			Processes: model.ProcessTable{
				1234: &model.Process{PID: 1234},
				5678: &model.Process{PID: 5678},
			},
		}
		r.Namespaces[model.NetNS] = model.NamespaceMap{}
		scanFd(0, "./test/fdscan/proc", true, &r)
		Expect(r.Namespaces[model.NetNS]).To(HaveLen(1))
		Expect(r.Namespaces[model.NetNS]).To(HaveKey(species.NamespaceID{Dev: stat.Dev, Ino: 12345678}))

		origns := r.Namespaces[model.NetNS][species.NamespaceID{Dev: stat.Dev, Ino: 12345678}]
		scanFd(0, "./test/fdscan/proc", true, &r)
		Expect(r.Namespaces[model.NetNS]).To(HaveLen(1))
		Expect(r.Namespaces[model.NetNS][species.NamespaceID{Dev: stat.Dev, Ino: 12345678}]).To(BeIdenticalTo(origns))
	})

	It("finds a network namespace a socket is connected to", func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		By("creating a transient new network namespace we only keep a socket connected to")
		netnsFd := netns.NewTransient()
		closeNetnsFd := sync.OnceFunc(func() { _ = unix.Close(netnsFd) })
		defer closeNetnsFd()

		netnsino := netns.Ino(netnsFd)

		nlh := nonetns.NewNetlinkHandle(netnsFd)
		defer nlh.Close()

		By("keeping only a socket as the last reference to the transient network namespace")
		closeNetnsFd()

		By("discovering the transient network namespace from the RTNETLINK socket")
		allns := Namespaces(FromFds())
		Expect(allns.Namespaces[model.NetNS]).To(
			HaveKey(species.NamespaceIDfromInode(netnsino)))
	})

	It("discovers the socket-to-process mapping", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		sockfd := Successful(unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0))
		defer func() { _ = unix.Close(sockfd) }()
		var sockstat unix.Stat_t
		Expect(unix.Fstat(sockfd, &sockstat)).To(Succeed())

		By("requesting scanning fds for socket network namespaces")
		allns := Namespaces(FromFds())
		Expect(allns.SocketProcessMap).To(HaveKeyWithValue(
			sockstat.Ino, ConsistOf(model.PIDType(os.Getpid()))))

		allns = Namespaces(WithSocketProcesses())
		Expect(allns.SocketProcessMap).To(HaveKeyWithValue(
			sockstat.Ino, ConsistOf(model.PIDType(os.Getpid()))))
	})

})
