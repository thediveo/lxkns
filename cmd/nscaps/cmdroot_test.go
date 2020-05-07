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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns"
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

var _ = Describe("branches and forks", func() {

	It("creates process branch", func() {
		procbr := processbranch(allns.Processes[tpid])
		l := last(procbr)
		Expect(l).To(BeAssignableToTypeOf(&processnode{}))
		Expect(l.(*processnode).proc.PID).To(Equal(tpid))
	})

	It("creates target branch for user namespace", func() {
		tbr := targetbranch(allns.Namespaces[lxkns.UserNS][tuserid])
		Expect(tbr).ToNot(BeNil())
		l := last(tbr)
		Expect(l).To(BeAssignableToTypeOf(&nsnode{}))
		Expect(l.(*nsnode).istarget).To(BeTrue())
		Expect(l.(*nsnode).ns.ID()).To(Equal(tuserid))
	})

	It("creates target branch for non-user namespace", func() {
		tbr := targetbranch(allns.Namespaces[lxkns.NetNS][tnsid])
		Expect(tbr).ToNot(BeNil())
		l := last(tbr)
		Expect(l).To(BeAssignableToTypeOf(&nsnode{}))
		Expect(l.(*nsnode).istarget).To(BeTrue())
		Expect(l.(*nsnode).ns.ID()).To(Equal(tnsid))
	})

	Describe("combines branches", func() {

		It("process outside target hierarchy", func() {
			procbr := processbranch(allns.Processes[lxkns.PIDType(os.Getpid())])
			tbr := targetbranch(allns.Namespaces[lxkns.NetNS][tnsid])
			root := fork(procbr, tbr)
			Expect(root).NotTo(BeNil())
			Expect(root.Children()).To(HaveLen(2))
			Expect(root.Children()[0]).To(BeAssignableToTypeOf(&processnode{}))
			Expect(root.Children()[1]).To(BeAssignableToTypeOf(&nsnode{}))
			Expect(root.Children()[1].(*nsnode).ns.ID()).To(Equal(tuserid))
		})

		It("process inside target hierarchy", func() {
			procbr := processbranch(allns.Processes[tpid].Parent)
			tbr := targetbranch(allns.Namespaces[lxkns.NetNS][tnsid])
			root := fork(procbr, tbr)
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
			procbr := processbranch(allns.Processes[procpid])
			tbr := targetbranch(allns.Namespaces[lxkns.NetNS][tnsid])
			root := fork(procbr, tbr)
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

	})

})
