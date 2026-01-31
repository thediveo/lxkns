// Copyright 2021 Harald Albrecht.
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
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("Discover mount points", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{})))
	})

	It("discovers from other mount namespace", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()

		bm := fmt.Sprintf("/tmp/bindmountpoint-%d", os.Getpid())
		scripts.Common(fmt.Sprintf(`bm=%s`, bm))
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Umr $stage2
`)
		scripts.Script("stage2", `
umount $bm || /bin/true # remove stale bind mount.
rmdir $bm-testdir || /bin/true
mkdir $bm-testdir
touch $bm-testdir/canary
mkdir $bm # make sure we have a thing to bind mount over.
mount --bind $bm-testdir $bm
process_namespaceid mnt # prints the "current" mount namespace ID.
read # wait for test to proceed()
umount $bm || /bin/true # clean up.
rmdir $bm || /bin/true
rm $bm-dir/canary || /bin/true
rmdir $bm-dir || /bin/true
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		netnsid := nstest.CmdDecodeNSId(cmd)
		allns := Namespaces(WithNamespaceTypes(species.CLONE_NEWNS), FromProcs(), WithMounts())

		namespacedmmap := allns.Mounts
		Expect(namespacedmmap).NotTo(BeNil())
		Expect(namespacedmmap).To(HaveKey(netnsid))
		mpmap := namespacedmmap[netnsid]
		Expect(mpmap).To(HaveKey(bm), func() string {
			keys := []string{}
			for key := range mpmap {
				keys = append(keys, key)
			}
			return fmt.Sprintf("keys: %s", strings.Join(keys, ", "))
		})
		initialmntnsid, err := ops.NewTypedNamespacePath("/proc/self/ns/mnt", species.CLONE_NEWNS).ID()
		Expect(err).NotTo(HaveOccurred())
		Expect(namespacedmmap[initialmntnsid]).NotTo(HaveKey(bm))
	})

})
