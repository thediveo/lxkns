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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
)

var _ = Describe("Discover mount points", func() {

	It("from other mount namespace", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()

		bm := "/tmp/bindmountpoint"
		scripts.Common(fmt.Sprintf(`bm=%s`, bm))
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Umr $stage2
`)
		scripts.Script("stage2", `
umount $bm || /bin/true # remove stale bind mount.
mkdir $bm # make sure we have a thing to bind mount over.
mount --bind /tmp $bm
process_namespaceid mnt # prints the "current" mount namespace ID.
read # wait for test to proceed()
umount $bm || /bin/true # clean up.
rmdir $bm || /bin/true
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		netnsid := nstest.CmdDecodeNSId(cmd)
		allns := Namespaces(WithNamespaceTypes(species.CLONE_NEWNS), FromProcs(), WithMounts())

		namespacedmmap := allns.Mounts
		Expect(namespacedmmap).NotTo(BeNil())
		Expect(namespacedmmap).To(HaveKey(netnsid))
		mpmap := namespacedmmap[netnsid]
		Expect(mpmap).To(HaveKey(bm))
		initialmntnsid, err := ops.NewTypedNamespacePath("/proc/self/ns/mnt", species.CLONE_NEWNS).ID()
		Expect(err).NotTo(HaveOccurred())
		Expect(namespacedmmap[initialmntnsid]).NotTo(HaveKey(bm))
	})

})
