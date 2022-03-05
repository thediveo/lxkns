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

package main

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/internal/namespaces"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

func last(n node) node {
	Expect(n).NotTo(BeNil())
	Expect(n).To(BeAssignableToTypeOf(&nsnode{}))
	for len(n.Children()) == 1 {
		Expect(n).To(BeAssignableToTypeOf(&nsnode{}))
		n = n.Children()[0]
	}
	return n
}

var _ = Describe("(in)capability", func() {

	It("get euid for arbitrary process", func() {
		euid := os.Geteuid()
		Expect(processEuid(&model.Process{PID: model.PIDType(os.Getpid())})).To(Equal(euid))
	})

	It("fails correctly", func() {
		_, _, err := caps(&model.Process{PID: 1}, nil)
		Expect(err).To(HaveOccurred())
		_, _, err = caps(allns.Processes[model.PIDType(os.Getpid())],
			namespaces.NewWithSimpleRef(species.CLONE_NEWNET, species.NoneID, ""))
		Expect(err).To(HaveOccurred())
	})

	It("rule #1: process inside target's user namespace", func() {
		tcaps, euid, err := caps(
			allns.Processes[tpid],
			allns.Namespaces[model.UserNS][tuserid])
		Expect(err).ToNot(HaveOccurred())
		Expect(tcaps).To(Equal(effcaps))
		Expect(euid).To(Equal(processEuid(allns.Processes[tpid])))

	})

	It("missed rule #2: process inside target's sibling user namespace", func() {
		tcaps, _, err := caps(
			allns.Processes[procpid],
			allns.Namespaces[model.NetNS][tnsid])
		Expect(err).ToNot(HaveOccurred())
		Expect(tcaps).To(Equal(incapable))
	})

	It("rules #2+#3: this process outside target's user namespace hierarchy", func() {
		tcaps, euid, err := caps(
			allns.Processes[model.PIDType(os.Getpid())],
			allns.Namespaces[model.NetNS][tnsid])
		Expect(err).ToNot(HaveOccurred())
		Expect(tcaps).To(Equal(allcaps))
		Expect(euid).To(Equal(os.Geteuid()))
	})

	It("rule #2, but not #3: process inside target's user namespace hierarchy, but different owner", func() {
		if os.Geteuid() == 0 {
			Skip("only non-root")
		}
		// need to fake this in case we're not ro(o)t.
		fakeinit := &model.Process{PID: 1}
		fakeinit.Namespaces[model.UserNS] =
			allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.UserNS]
		tcaps, _, err := caps(fakeinit, allns.Namespaces[model.NetNS][tnsid])
		Expect(err).ToNot(HaveOccurred())
		Expect(tcaps).To(Equal(effcaps))
	})

})

var _ = Describe("branches and forks", func() {

	It("creates process branch", func() {
		_, err := processbranch(&model.Process{PID: -1}, 0)
		Expect(err).To(HaveOccurred())

		euid := os.Getegid()
		procbr, err := processbranch(allns.Processes[tpid], euid)
		Expect(err).ToNot(HaveOccurred())
		l := last(procbr)
		Expect(l).To(BeAssignableToTypeOf(&processnode{}))
		Expect(l.(*processnode).proc.PID).To(Equal(tpid))
		Expect(l.(*processnode).euid).To(Equal(euid))
	})

	It("creates target branch for user namespace", func() {
		tbr := targetbranch(allns.Namespaces[model.UserNS][tuserid], effcaps)
		Expect(tbr).ToNot(BeNil())
		l := last(tbr)
		Expect(l).To(BeAssignableToTypeOf(&nsnode{}))
		Expect(l.(*nsnode).istarget).To(BeTrue())
		Expect(l.(*nsnode).targetcaps).To(Equal(effcaps))
		Expect(l.(*nsnode).ns.ID()).To(Equal(tuserid))
	})

	It("creates target branch for non-user namespace", func() {
		tbr := targetbranch(allns.Namespaces[model.NetNS][tnsid], incapable)
		Expect(tbr).ToNot(BeNil())
		l := last(tbr)
		Expect(l).To(BeAssignableToTypeOf(&nsnode{}))
		Expect(l.(*nsnode).istarget).To(BeTrue())
		Expect(l.(*nsnode).ns.ID()).To(Equal(tnsid))
	})

	Describe("combines branches", func() {

		It("process outside target hierarchy", func() {
			procbr, err := processbranch(allns.Processes[model.PIDType(os.Getpid())], 0)
			Expect(err).ToNot(HaveOccurred())
			tbr := targetbranch(allns.Namespaces[model.NetNS][tnsid], incapable)
			root := combine(procbr, tbr)
			Expect(root).NotTo(BeNil())
			Expect(root.Children()).To(HaveLen(2))
			Expect(root.Children()[0]).To(BeAssignableToTypeOf(&processnode{}))
			Expect(root.Children()[1]).To(BeAssignableToTypeOf(&nsnode{}))
			Expect(root.Children()[1].(*nsnode).ns.ID()).To(Equal(tuserid))
		})

		It("process inside target hierarchy", func() {
			procbr, err := processbranch(allns.Processes[tpid].Parent, 0)
			Expect(err).ToNot(HaveOccurred())
			tbr := targetbranch(allns.Namespaces[model.NetNS][tnsid], incapable)
			root := combine(procbr, tbr)
			Expect(root).NotTo(BeNil())
			Expect(root.Children()).To(HaveLen(1))
			userchild := root.Children()[0]

			Expect(userchild).To(BeAssignableToTypeOf(&nsnode{}))
			Expect(userchild.(*nsnode).ns.ID()).To(Equal(tuserid))
			Expect(userchild.Children()).To(HaveLen(2))
			Expect(userchild.Children()[0]).To(BeAssignableToTypeOf(&processnode{}))
			Expect(userchild.Children()[1]).To(BeAssignableToTypeOf(&nsnode{}))
			Expect(userchild.Children()[1].(*nsnode).ns.ID()).To(Equal(tnsid))
		})

		It("process outside target hierarchy in another child user namespace", func() {
			procbr, err := processbranch(allns.Processes[procpid], 0)
			Expect(err).ToNot(HaveOccurred())
			tbr := targetbranch(allns.Namespaces[model.NetNS][tnsid], incapable)
			root := combine(procbr, tbr)
			Expect(root).NotTo(BeNil())
			Expect(root.Children()).To(HaveLen(2))

			procuserchild := root.Children()[0]
			Expect(procuserchild).To(BeAssignableToTypeOf(&nsnode{}))
			Expect(procuserchild.Children()).To(HaveLen(1))
			Expect(procuserchild.Children()[0]).To(BeAssignableToTypeOf(&processnode{}))

			userchild := root.Children()[1]
			Expect(userchild).To(BeAssignableToTypeOf(&nsnode{}))
			Expect(userchild.Children()[0].(*nsnode).ns.ID()).To(Equal(tnsid))
		})

		It("process below target namespace", func() {
			procbr, err := processbranch(allns.Processes[tpid], 0)
			Expect(err).ToNot(HaveOccurred())
			tbr := targetbranch(allns.Processes[model.PIDType(os.Getpid())].Namespaces[model.UserNS], incapable)
			root := combine(procbr, tbr)
			Expect(root).NotTo(BeNil())
			Expect(root.(*nsnode).istarget).To(BeTrue())
			Expect(root.Children()).To(HaveLen(1))

			userchild := root.Children()[0]
			Expect(userchild).To(BeAssignableToTypeOf(&nsnode{}))

			Expect(userchild.Children()).To(HaveLen(1))
			procchild := userchild.Children()[0]
			Expect(procchild).To(BeAssignableToTypeOf(&processnode{}))
		})

	})

})
