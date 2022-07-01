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

package species

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace IDs", func() {

	BeforeEach(func() {
		nsfsdev = undefined
	})

	It("handles errors when detecting nsfs Device ID", func() {
		Expect(testableNsfsDev("foobar")).To(BeZero())
	})

	It("returns NoneID when nsfs detection fails", func() {
		nsfsdev = 0
		id, _ := IDwithType("net:[12345678]")
		Expect(id).To(Equal(NoneID))
	})

	It("compares namespace IDs", func() {
		ns1 := NamespaceID{Dev: 42, Ino: 123}
		ns11 := NamespaceID{Dev: 42, Ino: 123}
		Expect(ns1).To(Equal(ns11))

		ns2 := NamespaceID{Dev: 666, Ino: 123}
		Expect(ns1).NotTo(Equal(ns2))
	})

	It("parses namespace textual representations", func() {
		id, t := IDwithType("net:[1]")
		Expect(t).To(Equal(CLONE_NEWNET))
		Expect(id).To(Equal(NamespaceIDfromInode(1)))
	})

	DescribeTable("rejects invalid textual representations",
		func(text string) {
			id, t := IDwithType(text)
			Expect(t).To(Equal(NaNS), "%s is not a namespace", text)
			Expect(id).To(Equal(NoneID), "%s is not a namespace", text)
		},
		Entry("empty textual representation", ""),
		Entry("missing the colon", "net"),
		Entry("missing the colon 2", "net[1]"),
		Entry("missing the opening bracket", "net:"),
		Entry("missing the closing bracket", "net:[1"),
		Entry("unknown namespace type", "foo:[1]"),
		Entry("invalid negative namespace ID", "net:[-1]"),
		Entry("not a namespace ID number", "net:[123abc]"),
		Entry("too short a namespace type", "n:[1]"),
	)

	It("stringifies", func() {
		Expect(NamespaceID{Dev: 42, Ino: 123}.String()).To(Equal("NamespaceID(42,123)"))
	})

	It("converts inode numbers into namespace IDs", func() {
		Expect(NamespaceIDfromInode(0)).To(Equal(NoneID))
		Expect(NamespaceIDfromInode(123).Ino).To(Equal(uint64(123)))
	})

})
