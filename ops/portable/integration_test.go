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

package portable

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/thediveo/spacetest"
	"github.com/thediveo/spacetest/netns"
	"golang.org/x/sys/unix"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("portable reference integration", func() {

	BeforeEach(func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))
	})

	It("opens portable (network) namespace reference and runs a sub-process in it", func(ctx context.Context) {
		netnsfd := netns.NewTransient()
		// Try to reference and lock the new network namespace created by the
		// test script, then try to enter it and run a separate process attached
		// to the new network namespace.
		lockednetns, netnsunlocker := Successful2R(
			PortableReference{ID: species.NamespaceIDfromInode(netns.Ino(netnsfd)), Type: species.CLONE_NEWNET}.Open())
		defer netnsunlocker()
		res, err := ops.Execute(
			func() any {
				cmd := exec.Command("ls", "-l", "/proc/self/ns/net")
				out, err := cmd.CombinedOutput()
				if err != nil {
					return err
				}
				return out
			},
			lockednetns)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeAssignableToTypeOf([]byte{}))
		b, _ := res.([]byte)
		Expect(string(b)).To(MatchRegexp(fmt.Sprintf(`net:\[%d\]`, netns.Ino(netnsfd))))
	})

	It("keeps an Open()ed portable namespace reference open/locked without any processes left", func() {
		netnsfd := spacetest.NewUnmanagedTransient(unix.CLONE_NEWNET)
		netnsid := species.NamespaceIDfromInode(netns.Ino(netnsfd))
		closenetns := sync.OnceFunc(func() {
			Expect(unix.Close(netnsfd)).To(Succeed())
		})
		defer closenetns()
		// Must keep the returned new namespace reference alive till the end, as
		// otherwise garbage collecting it will prematurely close the wrapped
		// *os.File ... and we don't want THAT.
		lockednetns, netnsunlocker := Successful2R(
			PortableReference{ID: species.NamespaceIDfromInode(netns.Ino(netnsfd)), Type: species.CLONE_NEWNET}.Open())
		defer netnsunlocker()
		// Remove our hold on the newly created network namespace.
		closenetns()
		// Wait a short time so that the network namespace could be garbage
		// collected, weren't it for the locking reference we're still keeping.
		// We check in two steps: first, the namespace must not be found anymore
		// in process references. Second, it must still be present in open file
		// descriptor references.
		time.Sleep(time.Second)
		netns := discover.Namespaces(discover.FromProcs(), discover.WithNamespaceTypes(species.CLONE_NEWNET))
		Expect(netns.Namespaces[model.NetNS]).NotTo(HaveKey(netnsid),
			"temporary network namespace still found in processes")
		netns = discover.Namespaces(discover.FromFds(), discover.WithNamespaceTypes(species.CLONE_NEWNET))
		Expect(netns.Namespaces[model.NetNS]).To(HaveKey(netnsid),
			"temporary network namespace missing from open fd references")
		// Unlock temporary network namespace and wait a short time for things
		// to settle. Then check.
		netnsunlocker()
		netns = discover.Namespaces(discover.SameAs(netns))
		Expect(netns.Namespaces[model.NetNS]).NotTo(HaveKey(netnsid),
			"this *#@!& temporary network namespace won't get away!")
		// The wrapping namespace reference can from now on be garbage
		// collected.
		runtime.KeepAlive(lockednetns)
	})

})
