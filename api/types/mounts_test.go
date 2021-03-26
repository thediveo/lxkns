// Copyright 2021 Harald Albrecht.
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
	"encoding/json"

	"github.com/PaesslerAG/jsonpath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thediveo/go-mntinfo"
	"github.com/thediveo/lxkns/mounts"
)

var mountpathmap = mounts.NewMountPathMap([]mntinfo.Mountinfo{
	{MountPoint: "/", MountID: 2, ParentID: 1},
	{MountPoint: "/a", MountID: 3, ParentID: 2},
	{MountPoint: "/b", MountID: 4, ParentID: 2},
})

type M map[string]interface{}

func (m M) get(path string) interface{} {
	val, err := jsonpath.Get(path, map[string]interface{}(m))
	Expect(err).NotTo(HaveOccurred(), "no %q in %v", path, m)
	return val
}

var _ = Describe("NamespacedMountMap JSON", func() {

	It("marshals mount paths from a single namespace", func() {
		j, err := json.Marshal(MountPathMap(mountpathmap))
		Expect(err).NotTo(HaveOccurred())
		var m M
		Expect(json.Unmarshal([]byte(j), &m)).NotTo(HaveOccurred())

		rootid := m.get(`$["/"].pathid`).(float64)
		Expect(rootid).NotTo(BeZero())
		Expect(m.get(`$["/"].parentid`)).To(BeZero())

		aid := m.get(`$["/a"].pathid`).(float64)
		Expect(aid).NotTo(BeZero())
		Expect(m.get(`$["/a"].parentid`)).To(Equal(rootid))

		bid := m.get(`$["/b"].pathid`).(float64)
		Expect(bid).NotTo(BeZero())
		Expect(m.get(`$["/b"].parentid`)).To(Equal(rootid))

		Expect(m.get(`$["/"].childrenids`)).To(ConsistOf(bid, aid))
	})

})
