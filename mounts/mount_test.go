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

package mounts

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/thediveo/go-mntinfo"
)

var sillyMounts = []mntinfo.Mountinfo{
	{MountPoint: "/sys", MountID: 3, ParentID: 2},
	{MountPoint: "/", MountID: 2, ParentID: 1},
	{MountPoint: "/sys/kernel/debug", MountID: 4, ParentID: 3},
	// mount over /sys, hiding previous /sys...
	{MountPoint: "/sys", MountID: 666, ParentID: 3},
}

var _ = Describe("MountPath", func() {

	It("builds a tree", func() {
		root := mountPathTree(sillyMounts)
		Expect(root).NotTo(BeNil())
		Expect(root.Path()).To(Equal("/"))
		Expect(root.Parent).To(BeNil())
		Expect(root.Children).To(HaveLen(1)) // only "/sys" (ID:3)
		Expect(root.Mounts).To(HaveLen(1))

		sys := root.Children[0]
		Expect(sys.Path()).To(Equal("/sys"))
		Expect(sys.Parent).To(Equal(root))
		Expect(sys.Children).To(HaveLen(1))
		Expect(sys.Mounts).To(HaveLen(2)) // ID:3 and ID:666

		syskerneldebug := sys.Children[0]
		Expect(syskerneldebug.Path()).To(Equal("/sys/kernel/debug"))
		Expect(syskerneldebug.Parent).To(Equal(sys))

		rootmount := root.Mounts[0]
		Expect(rootmount.Parent).To(BeNil())
		Expect(rootmount.Children).To(HaveLen(1))

		sysmount := rootmount.Children[0]
		Expect(sysmount.MountID).To(Equal(3))
		Expect(sysmount.Children).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Mountinfo": MatchFields(IgnoreExtras, Fields{
					"MountID": Equal(4),
				}),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Mountinfo": MatchFields(IgnoreExtras, Fields{
					"MountID": Equal(666),
				}),
			})),
		))
	})

})
