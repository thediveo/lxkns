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
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("Discover owning user namespaces", func() {

	It("finds owners", func() {
		allns := Discover(FromProcs(), WithHierarchy(), WithOwnership())

		myusernsid, err := ops.NamespacePath("/proc/self/ns/user").ID()
		Expect(err).To(Succeed())
		Expect(allns.Namespaces[model.UserNS]).To(HaveKey(myusernsid))
		userns := allns.Namespaces[model.UserNS][myusernsid]
		for _, nst := range []string{"cgroup", "ipc", "mnt", "net", "pid", "uts"} {
			mynsid, err := ops.NamespacePath("/proc/self/ns/" + nst).ID()
			Expect(err).To(Succeed())
			Expect(allns.Namespaces[model.TypeIndex(species.NameToType(nst))]).To(HaveKey(mynsid))
			owneruserns := allns.Namespaces[model.TypeIndex(species.NameToType(nst))][mynsid].Owner()
			Expect(owneruserns).To(BeIdenticalTo(userns))
		}
	})

})
