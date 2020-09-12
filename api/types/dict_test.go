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

package types

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var _ = Describe("NamespacesDict namespace references dictionary", func() {

	It("always gets namespaces", func() {
		allns := NewNamespacesDict()
		uns := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(123), "/foobar")
		allns[model.UserNS][uns.ID()] = uns

		ns := allns.Get(uns.ID(), uns.Type())
		Expect(ns).To(BeIdenticalTo(uns))

		ns = allns.Get(species.NamespaceIDfromInode(666), species.CLONE_NEWNET)
		Expect(ns).NotTo(BeNil())
		Expect(ns.ID()).To(Equal(species.NamespaceIDfromInode(666)))
		Expect(ns.Type()).To(Equal(species.CLONE_NEWNET))
		Expect(ns.Ref()).To(BeZero())
	})

})
