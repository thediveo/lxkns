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
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"
	"golang.org/x/sys/unix"
)

var _ = Describe("Discover from fds", func() {

	It("finds fd-referenced namespaces", func() {
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
		var fdnetnsid, netnsid species.NamespaceID
		cmd.Decode(&fdnetnsid)
		cmd.Decode(&netnsid)
		Expect(fdnetnsid).ToNot(Equal(netnsid))
		// correctly misses fd-referenced namespaces without proper discovery
		// method enabled.
		opts := NoDiscovery
		opts.SkipProcs = false
		allns := Discover(opts)
		Expect(allns.Namespaces[model.NetNS]).To(HaveKey(netnsid))
		Expect(allns.Namespaces[model.NetNS]).ToNot(HaveKey(fdnetnsid))
		// correctly finds fd-referenced namespaces now.
		opts = NoDiscovery
		opts.SkipFds = false
		allns = Discover(opts)
		Expect(allns.Namespaces[model.NetNS]).To(HaveKey(fdnetnsid))
	})

	It("skips /proc/*/fd/* nonsense", func() {
		var stat unix.Stat_t
		Expect(unix.Stat("./test/fdscan/proc", &stat)).ToNot(HaveOccurred())
		Expect(stat.Dev).NotTo(BeZero())
		r := DiscoveryResult{
			Options: NoDiscovery,
			Processes: model.ProcessTable{
				1234: &model.Process{PID: 1234},
				5678: &model.Process{PID: 5678},
			},
		}
		r.Options.SkipFds = false
		r.Namespaces[model.NetNS] = model.NamespaceMap{}
		scanFd(0, "./test/fdscan/proc", true, &r)
		Expect(r.Namespaces[model.NetNS]).To(HaveLen(1))
		Expect(r.Namespaces[model.NetNS]).To(HaveKey(species.NamespaceID{Dev: stat.Dev, Ino: 12345678}))

		origns := r.Namespaces[model.NetNS][species.NamespaceID{Dev: stat.Dev, Ino: 12345678}]
		scanFd(0, "./test/fdscan/proc", true, &r)
		Expect(r.Namespaces[model.NetNS]).To(HaveLen(1))
		Expect(r.Namespaces[model.NetNS][species.NamespaceID{Dev: stat.Dev, Ino: 12345678}]).To(BeIdenticalTo(origns))
	})

})
