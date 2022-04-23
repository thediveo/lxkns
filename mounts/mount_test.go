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
	"time"

	"github.com/thediveo/go-mntinfo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("MountPath", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithPolling(100 * time.Millisecond).ShouldNot(HaveLeaked())
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("builds a tree", func() {
		mp := NewMountPathMap([]mntinfo.Mountinfo{
			{MountPoint: "/a", MountID: 3, ParentID: 2},
			{MountPoint: "/", MountID: 2, ParentID: 1},
			{MountPoint: "/a/b/c", MountID: 4, ParentID: 3},
			{MountPoint: "/a", MountID: 666, ParentID: 3},
		})
		root := mp["/"]

		Expect(root).NotTo(BeNil())
		Expect(root.Path()).To(Equal("/"))
		Expect(root.Parent).To(BeNil())
		Expect(root.Children).To(HaveLen(1)) // only "/a" (ID:3)
		Expect(root.Mounts).To(HaveLen(1))

		sys := root.Children[0]
		Expect(sys.Path()).To(Equal("/a"))
		Expect(sys.Parent).To(Equal(root))
		Expect(sys.Children).To(HaveLen(1))
		Expect(sys.Mounts).To(HaveLen(2)) // ID:3 and ID:666

		syskerneldebug := sys.Children[0]
		Expect(syskerneldebug.Path()).To(Equal("/a/b/c"))
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

	It("detects overmounts", func() {
		mp := NewMountPathMap([]mntinfo.Mountinfo{
			{MountPoint: "/", MountID: 2, ParentID: 1},
			{MountPoint: "/a", MountID: 3, ParentID: 2},
			{MountPoint: "/a/b", MountID: 30, ParentID: 3},
			{MountPoint: "/a", MountID: 4, ParentID: 3},
		})

		Expect(mp["/a/b"].Mounts).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Mountinfo": MatchFields(IgnoreExtras, Fields{
					"MountID": Equal(30),
				}),
				"Hidden": BeTrue(),
			})),
		))
		Expect(mp["/a"].Mounts).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Mountinfo": MatchFields(IgnoreExtras, Fields{
					"MountID": Equal(4),
				}),
				"Hidden": BeFalse(),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Mountinfo": MatchFields(IgnoreExtras, Fields{
					"MountID": Equal(3),
				}),
				"Hidden": BeTrue(),
			})),
		))

		Expect(mp["/a"].VisibleMount().MountID).To(Equal(4))
		Expect(mp["/a/b"].VisibleMount()).To(BeNil())
	})

	It("finds prefix overmounts including in-place overmounts", func() {
		mp := NewMountPathMap([]mntinfo.Mountinfo{
			{MountPoint: "/", MountID: 2, ParentID: 1},
			{MountPoint: "/a/b", MountID: 3, ParentID: 2},
			{MountPoint: "/a/b/c", MountID: 4, ParentID: 3},
			{MountPoint: "/a/b/c", MountID: 40, ParentID: 4},
			{MountPoint: "/a/b/c/e/f", MountID: 41, ParentID: 40},
			{MountPoint: "/a", MountID: 5, ParentID: 2},
			{MountPoint: "/a/b/c/e/f", MountID: 50, ParentID: 5},
		})

		Expect(mp["/a/b/c/e/f"].Mounts).To(HaveLen(2))

		Expect(mp["/a"].Mounts[0].Hidden).To(BeFalse())
		Expect(mp["/a/b"].Mounts[0].Hidden).To(BeTrue())
		Expect(mp["/a/b/c"].Mounts[0].Hidden).To(BeTrue())
		Expect(mp["/a/b/c"].Mounts[1].Hidden).To(BeTrue())
		Expect(mp["/a/b/c/e/f"].Mounts).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Mountinfo": MatchFields(IgnoreExtras, Fields{
					"MountID": Equal(41),
				}),
				"Hidden": BeTrue(),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Mountinfo": MatchFields(IgnoreExtras, Fields{
					"MountID": Equal(50),
				}),
				"Hidden": BeFalse(),
			})),
		))
	})

})
