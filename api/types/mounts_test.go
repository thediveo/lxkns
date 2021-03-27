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
	"github.com/thediveo/lxkns"
	"github.com/thediveo/lxkns/mounts"
	"github.com/thediveo/lxkns/species"
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
		jtext, err := json.Marshal(MountPathMap(mountpathmap))
		Expect(err).NotTo(HaveOccurred())
		var m M
		Expect(json.Unmarshal([]byte(jtext), &m)).NotTo(HaveOccurred())

		rootid := m.get(`$["/"].pathid`).(float64)
		Expect(rootid).NotTo(BeZero())
		Expect(m.get(`$["/"].parentid`)).To(BeZero())

		aid := m.get(`$["/a"].pathid`).(float64)
		Expect(aid).NotTo(BeZero())
		Expect(m.get(`$["/a"].parentid`)).To(Equal(rootid))

		bid := m.get(`$["/b"].pathid`).(float64)
		Expect(bid).NotTo(BeZero())
		Expect(m.get(`$["/b"].parentid`)).To(Equal(rootid))
	})

	It("unmarshals its own marshalling", func() {
		mpm := MountPathMap(mountpathmap)
		jtext, err := json.Marshal(mpm)
		Expect(err).NotTo(HaveOccurred())
		var m MountPathMap
		Expect(json.Unmarshal(jtext, &m)).NotTo(HaveOccurred())
		for path, mountpath := range m {
			// Checks mount path hierarchy...
			if mpm[path].Parent != nil {
				Expect(mountpath.Parent.Path()).To(Equal(mpm[path].Parent.Path()))
			} else {
				Expect(mountpath.Parent).To(BeNil())
			}
			children := []string{}
			for _, child := range mountpath.Children {
				children = append(children, child.Path())
			}
			ochildren := []string{}
			for _, ochild := range mpm[path].Children {
				ochildren = append(ochildren, ochild.Path())
			}
			Expect(children).To(ConsistOf(ochildren))
			// Checks mount point hierarchy...
			for _, mount := range mountpath.Mounts {
				var omount *mounts.MountPoint
				for _, om := range mpm[path].Mounts {
					if om.MountID == mount.MountID {
						omount = om
						break
					}
				}
				Expect(omount).NotTo(BeNil())
				if omount.Parent != nil {
					Expect(mount.Parent.MountID).To(Equal(omount.Parent.MountID))
				} else {
					Expect(mount.Parent).To(BeNil())
				}
				children := []int{}
				for _, child := range mount.Children {
					children = append(children, child.MountID)
				}
				ochildren := []int{}
				for _, ochild := range omount.Children {
					ochildren = append(ochildren, ochild.MountID)
				}
				Expect(children).To(ConsistOf(ochildren))
			}
		}
	})

	It("marshals mount path maps from multiple mount namespaces", func() {
		allm := lxkns.NamespacedMountPathMap{
			species.NamespaceIDfromInode(123): mounts.MountPathMap(mountpathmap),
		}
		jtext, err := json.Marshal(NamespacedMountMap(allm))
		Expect(err).NotTo(HaveOccurred())
		var m M
		Expect(json.Unmarshal([]byte(jtext), &m)).NotTo(HaveOccurred())
		Expect(m.get(`$["123"]`)).NotTo(BeNil())
		Expect(m.get(`$["123"]["/"]`)).NotTo(BeNil())
	})

	It("unmarshals its own JSON", func() {
		allm := lxkns.NamespacedMountPathMap{
			species.NamespaceIDfromInode(123): mounts.MountPathMap(mountpathmap),
		}
		jtext, err := json.Marshal(NamespacedMountMap(allm))
		Expect(err).NotTo(HaveOccurred())
		var m NamespacedMountMap
		Expect(json.Unmarshal(jtext, &m)).To(Succeed())
		Expect(m).To(HaveKey(species.NamespaceIDfromInode(123)))
		Expect(m[species.NamespaceIDfromInode(123)]).To(HaveKey("/a"))
		Expect(m[species.NamespaceIDfromInode(123)]["/a"].Mounts[0].MountID).To(Equal(3))
	})

})
