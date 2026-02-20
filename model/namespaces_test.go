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

package model

import (
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/ginkgo/v2/dsl/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("model", func() {

	It("TypeIndex() fails for invalid kernel namespace type", func() {
		Expect(TypeIndex(species.CLONE_NEWCGROUP | species.CLONE_NEWNET)).To(
			Equal(NamespaceTypeIndex(-1)))
	})

	DescribeTable("returns type indices for namespace type names",
		func(name string, expected NamespaceTypeIndex, expectedok bool) {
			idx, ok := NamespaceTypeIndexByName(name)
			Expect(ok).To(Equal(expectedok))
			Expect(idx).To(Equal(idx))
		},
		Entry(nil, "foo", NamespaceTypeIndex(0), false),
		Entry(nil, "mnt", MountNS, true),
		Entry(nil, "cgroup", CgroupNS, true),
		Entry(nil, "uts", UTSNS, true),
		Entry(nil, "ipc", IPCNS, true),
		Entry(nil, "user", UserNS, true),
		Entry(nil, "pid", PIDNS, true),
		Entry(nil, "net", NetNS, true),
		Entry(nil, "time", TimeNS, true),
	)

	DescribeTable("returns namespace type names for type indices",
		func(idx NamespaceTypeIndex, expected string, expectedok bool) {
			tname, ok := NamespaceTypeNameByIndex(idx)
			Expect(ok).To(Equal(expectedok))
			Expect(tname).To(Equal(expected))
		},
		Entry(nil, NamespaceTypeIndex(-1), "", false),
		Entry(nil, MountNS, "mnt", true),
		Entry(nil, CgroupNS, "cgroup", true),
		Entry(nil, UTSNS, "uts", true),
		Entry(nil, IPCNS, "ipc", true),
		Entry(nil, UserNS, "user", true),
		Entry(nil, PIDNS, "pid", true),
		Entry(nil, NetNS, "net", true),
		Entry(nil, TimeNS, "time", true),
	)

})
