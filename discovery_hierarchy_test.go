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
	r "github.com/thediveo/lxkns/relations"
	"github.com/thediveo/testbasher"
)

var _ = Describe("Discover hierarchy", func() {

	It("finds hidden hierarchical user namespaces", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Ur unshare -U $stage2 # create a user ns with another user ns inside.
`)
		scripts.Script("stage2", `
process_namespaceid user # prints the user namespace ID of "the" process.
read # wait for test to proceed()
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var usernsid t.NamespaceID
		cmd.Decode(&usernsid)
		allns := Discover(FullDiscovery)
		userns := allns.Namespaces[UserNS][usernsid].(Hierarchy)
		Expect(userns).NotTo(BeNil())
		ppusernsid, _ := r.ID("/proc/self/ns/user")
		Expect(userns.Parent().Parent().(Namespace).ID()).To(Equal(ppusernsid))
	})

	It("adds child namespaces only once", func() {
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Script("main", `
unshare -Umnr unshare -Un $stage2
`)
		scripts.Script("stage2", `
echo "\"ready\""
read
`)
		cmd := scripts.Start("main")
		defer cmd.Close()
		var ready string
		cmd.Decode(&ready)
		allns := Discover(FullDiscovery)
		for _, uns := range allns.Namespaces[UserNS] {
			children := uns.(Hierarchy).Children()
			for chidx, child := range children {
				for checkidx, checkchild := range children {
					if child == checkchild && chidx != checkidx {
						Fail("duplicate child")
					}
				}
			}
		}
	})

})
