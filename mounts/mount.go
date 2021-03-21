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
	"path/filepath"

	"github.com/thediveo/go-mntinfo"
)

// For details about the mount point information provided by Linux via
// /proc/$PID/mountinfo, please also refer to:
// https://man7.org/linux/man-pages/man5/procfs.5.html

// MountPath represents a path name with one or even several(!) mount points.
// Additionally, MountPaths form a ("sparse") hierarchy of mount paths with
// references to their nearest parent mount path as well as child mount paths.
type MountPath struct {
	Mounts   []*MountInfo // one or several mount points at this same path.
	Parent   *MountPath   // Parent mount path, except for root mount path.
	Children []*MountPath // Children mount paths.
}

// MountInfo contains information about a single mount point with additional
// object references to parent and child mount points. The parent/child
// references base on the mount and parent IDs from /proc/$PID/mountinfo.
type MountInfo struct {
	mntinfo.Mountinfo              // mount (point) information.
	Parent            *MountInfo   // parent mount point, if its ID could be resolved.
	Children          []*MountInfo // child mount points, derived from mount and parent IDs.
}

// Path returns the path name of a MountPath object.
func (mp MountPath) Path() string {
	// As all mount points of this MountPath share the same path name, we don't
	// store the path explicitly but instead simply take it from the first
	// MountInfo object.
	return mp.Mounts[0].MountPoint
}

// mountPathTree takes a list of mounts (mount information) and arranges them
// into a tree of MountPaths, returning the root MountPath object.
func mountPathTree(mounts []mntinfo.Mountinfo) (root *MountPath) {
	// First gather all (unique) mount paths from the mounts and reference the
	// corresponding mounts from them. At the same time gather all mount point
	// IDs, so we can later resolve them into object references.
	mountpathmap := map[string]*MountPath{}
	mountidmap := map[int]*MountInfo{}
	for _, mount := range mounts {
		mp, ok := mountpathmap[mount.MountPoint]
		if !ok {
			mp = &MountPath{}
			mountpathmap[mount.MountPoint] = mp
		}
		mnt := &MountInfo{Mountinfo: mount}
		mp.Mounts = append(mp.Mounts, mnt)
		mountidmap[mnt.MountID] = mnt
	}
	// Build the tree based on the mount paths seen: for each mount path go up
	// the path hierarchy until we find the nearest parent mount path. Then
	// create bidirectional references between the particular mount path and its
	// nearest parent mount path.
	for mountpath, mp := range mountpathmap {
		for mountpath != "/" {
			mountpath = filepath.Dir(mountpath)
			if parentmp, ok := mountpathmap[mountpath]; ok {
				mp.Parent = parentmp
				parentmp.Children = append(parentmp.Children, mp)
				break
			}
		}
	}
	// Build the tree of mount points, this time based on the mount IDs as
	// opposed to mount paths. Please note that usually the root of the root
	// will be outside of our mount information; this is a normal situation, so
	// we simply skip unknown parent mount IDs.
	for _, mount := range mountidmap {
		if parentmount, ok := mountidmap[mount.ParentID]; ok {
			mount.Parent = parentmount
			parentmount.Children = append(parentmount.Children, mount)
		}
	}
	// Done, return root of tree.
	return mountpathmap["/"]
}
