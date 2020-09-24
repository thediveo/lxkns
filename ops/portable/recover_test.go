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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("recovering namespaces", func() {

	It("locates a namespace by ID only", func() {
		netns := LocateNamespace(mynetnsid, 0)
		Expect(netns).NotTo(BeNil())
		Expect(netns.ID()).To(Equal(mynetnsid))
	})

	It("locates a namespace by ID and type", func() {
		netns := LocateNamespace(mynetnsid, species.CLONE_NEWNET)
		Expect(netns).NotTo(BeNil())
		Expect(netns.ID()).To(Equal(mynetnsid))
	})

	It("fails to locate a namespace with wrong type", func() {
		netns := LocateNamespace(mynetnsid, species.CLONE_NEWUSER)
		Expect(netns).To(BeNil())
	})

	It("fails to locate a namespace with wrong ID", func() {
		netns := LocateNamespace(species.NamespaceIDfromInode(666), 0)
		Expect(netns).To(BeNil())
		netns = LocateNamespace(species.NoneID, 0)
		Expect(netns).To(BeNil())
	})

})
