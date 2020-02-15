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

package lxkns

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/nstest"
	t "github.com/thediveo/lxkns/nstypes"
	"github.com/thediveo/testbasher"
)

var _ = Describe("Discover from bind-mounts", func() {

	It("finds hidden hierarchical user namespaces", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Common(`bm=/tmp/netbindmount`)
		scripts.Script("main", `
# Create new user and mount namespace to bind-mount things in. We need the
# user namespace in order to gain full capabilities, including the mount
# capability. While this will drive Debian Disciples into a frenzy, we
# rather run our tests. Creating a separate mount namespace is necessary as we
# don't have capabilities in the user namespace we started from. And it has the
# nice side-effect that this tests discovery in other mount namespaces than the
# initial mount namespace.
unshare -Umr $stage2
`)
		scripts.Script("stage2", `
umount $bm || /bin/true # remove stale bind mount.
touch $bm # make sure we have a thing to bind mount over.
unshare -n $stage2a # create new net namespace and bind-mount it.
read # wait for test to proceed()
umount $bm || /bin/true # clean up.
rm $bm || /bin/true
`)
		scripts.Script("stage2a", `
process_namespaceid net # prints the "current" net namespace ID.
mount --bind /proc/self/ns/net $bm
# That's it: the end of the script, continue in stage2...
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var netnsid t.NamespaceID
		cmd.Decode(&netnsid)
		opts := NoDiscovery
		opts.SkipBindmounts = false
		allns := Discover(FullDiscovery)
		Expect(allns.Namespaces[NetNS]).To(HaveKey(netnsid))
	})

})
