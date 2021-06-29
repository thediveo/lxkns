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

package gmodel

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/errxpect"
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

var allns *lxkns.DiscoveryResult
var userns1 model.Namespace
var initproc *model.Process

var _ = BeforeSuite(func() {
	allns = lxkns.Discover(lxkns.WithStandardDiscovery())
	initproc = allns.Processes[model.PIDType(os.Getpid())]
	userns1 = initproc.Namespaces[model.UserNS]
})

var _ = Describe("Namespace", func() {

	It("matches same ID and Type", func() {
		ns1 := namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(1), "/foobar")
		ns2 := namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(1), "/foobar")
		Expect(ns1).To(BeSameNamespace(ns2))
		Expect(ns1).NotTo(
			BeSameNamespace(namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(1), "/foobar")))
		Expect(ns1).NotTo(
			BeSameNamespace(namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(666), "/foobar")))
		Expect(ns1).NotTo(
			BeSameNamespace(namespaces.New(species.CLONE_NEWNET, species.NamespaceIDfromInode(1), "/foo/bar")))
	})

	It("matches PID lists", func() {
		Expect(sameLeaders([]model.PIDType{}, []model.PIDType{1})).To(BeFalse())
		Expect(sameLeaders([]model.PIDType{1}, []model.PIDType{2})).To(BeFalse())
		Expect(sameLeaders([]model.PIDType{1, 2}, []model.PIDType{1, 2})).To(BeTrue())
		Expect(sameLeaders([]model.PIDType{1, 2}, []model.PIDType{2, 1})).To(BeTrue())
	})

	It("matches namespaces", func() {
		Expect(nil).To(BeSameNamespace(nil))
		Errxpect(BeSameNamespace("bar").Match(userns1)).To(HaveOccurred())

		// ID, Type, and ref path must be checked
		Expect(userns1).NotTo(BeSameNamespace(initproc.Namespaces[model.NetNS]))

		u2 := *(userns1.(*namespaces.UserNamespace))
		userns2 := model.Namespace(&u2)
		Expect(userns1).To(BeSameNamespace(userns2))

		// leader PIDs must be checked
		u2 = *(userns1.(*namespaces.UserNamespace))
		userns2.(namespaces.NamespaceConfigurer).AddLeader(initproc)
		Expect(userns1).NotTo(BeSameNamespace(userns2))

		// Guard ;) :p
		u2 = *(userns1.(*namespaces.UserNamespace))
		Expect(userns1).To(BeSameNamespace(userns2))

		// parents must be checked
		u2 = *(userns1.(*namespaces.UserNamespace))
		dummyuserns := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(1), "/roode")
		dummyuserns.(namespaces.HierarchyConfigurer).AddChild(model.Hierarchy(&u2))
		Expect(userns1).NotTo(BeSameNamespace(userns2))

		// uid must be checked
		u2 = *(userns1.(*namespaces.UserNamespace))
		userns2.(namespaces.UserConfigurer).SetOwnerUID(123)
		Expect(userns1).NotTo(BeSameNamespace(userns2))

		// children must be same...
		u0 := *(userns1.(*namespaces.UserNamespace))
		c1 := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(667), "/667")
		c2 := namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(668), "/668")
		(namespaces.HierarchyConfigurer)(&u0).AddChild(c1.(model.Hierarchy))
		(namespaces.HierarchyConfigurer)(&u0).AddChild(c2.(model.Hierarchy))
		Expect(&u0).NotTo(BeSameNamespace(userns1))

		u2 = u0
		Expect(&u2).To(BeSameNamespace(&u0))

		// add the "same" children to a clean copy without any children, then
		// expect both parent namespaces to still be the "same".
		u2 = *(userns1.(*namespaces.UserNamespace))
		c1 = namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(667), "/667")
		c2 = namespaces.New(species.CLONE_NEWUSER, species.NamespaceIDfromInode(668), "/668")
		(namespaces.HierarchyConfigurer)(&u2).AddChild(c1.(model.Hierarchy))
		Expect(&u2).NotTo(BeSameNamespace(&u0))
		(namespaces.HierarchyConfigurer)(&u2).AddChild(c2.(model.Hierarchy))
		Expect(&u2).To(BeSameNamespace(&u0))
	})

	It("handles errors", func() {
		Errxpect(BeSameNamespace(userns1).Match("bar")).To(
			MatchError(MatchRegexp(`expects a model.Namespace, not a string`)))

		Errxpect(BeSameNamespace("foo").Match(userns1)).To(
			MatchError(MatchRegexp(`must be passed a model.Namespace, not a string`)))
	})

})
