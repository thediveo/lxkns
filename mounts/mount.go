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
	"path"
	"sort"
	"strings"

	"github.com/thediveo/go-mntinfo"
)

// For details about the mount point information provided by Linux via
// /proc/$PID/mountinfo, please also refer to:
// https://man7.org/linux/man-pages/man5/procfs.5.html

// MountPathMap maps mount paths to (potentially multiple) mount points. If
// there are multiple mount points with the same mount path, then only at most
// one of them will be visible, unless this mount path is completely hidden by
// an overmount with a shorter VFS prefix and no child mounts with the this
// prefix path.
type MountPathMap map[string]*MountPath

// MountPath represents a path name to a “place” in the VFS where there are
// one or even multiple mount points. MountPaths form a (“sparse”) hierarchy
// of mount paths with references to their nearest parent mount path (that is,
// longest path prefix) as well as child mount paths.
//
// For instance, given the mount paths "/a" and "/a/b/c" the model only stored
// "/a" and "/a/b/c", but there is no mount path "/a/b".
//
// Use [MountPath.Path] to get the mount path of a MountPath object.
type MountPath struct {
	Mounts   []*MountPoint `json:"mounts"` // one or several mount points at this same (VFS) path.
	Parent   *MountPath    `json:"-"`      // Parent mount path, except for root mount path.
	Children []*MountPath  `json:"-"`      // Children mount paths.
}

// MountPoint contains information about a single mount point with additional
// object references to parent and child mount points.
//
// MountPoint objects form a hierarchical tree that is separate from the mount
// path tree. The parent/child references base on the mount and parent IDs (from
// /proc/$PID/mountinfo), instead of mount paths.
type MountPoint struct {
	mntinfo.Mountinfo               // mount (point) information.
	Hidden            bool          `json:"hidden"` // mount point hidden or "overmounted".
	Parent            *MountPoint   `json:"-"`      // parent mount point, if its ID could be resolved.
	Children          []*MountPoint `json:"-"`      // child mount points, derived from mount and parent IDs.
}

// Path returns the path name of a [MountPath] object.
func (p MountPath) Path() string {
	// As all mount points of this MountPath share the same path name, we don't
	// store the path explicitly but instead simply take it from the first
	// MountInfo object.
	return p.Mounts[0].MountPoint
}

// VisibleMount returns the (only) visible mount point at this mount path, or
// nil if all the mount points at this mount path are hidden by overmounts.
func (p MountPath) VisibleMount() *MountPoint {
	for _, mountpoint := range p.Mounts {
		if !mountpoint.Hidden {
			return mountpoint
		}
	}
	return nil
}

// NewMountPathMap takes a list of mounts (mount information) and builds a map
// based on the mount paths, taking into account that due to overmounting the
// same mount path may map onto multiple mount points (but with only one being
// not hidden).
//
// Additionally, the mount information accessible by the map is arranged into
// two separate trees: one tree for the hierarchy of mount paths and a separate
// tree for the mount point hierarchy.
func NewMountPathMap(mounts []mntinfo.Mountinfo) (mountpathmap MountPathMap) {
	// Bail out immediately when there are no mounts to process. This may happen
	// when reading the mount point information for a particular mount namespace
	// failed.
	if len(mounts) == 0 {
		return
	}
	// First gather all (unique) mount paths from the mounts and reference the
	// corresponding mounts from them. At the same time gather all mount point
	// IDs, so we can later resolve them into object references.
	mountpathmap = MountPathMap{}
	mountidmap := map[int]*MountPoint{}
	for _, mount := range mounts {
		mpoint, ok := mountpathmap[mount.MountPoint]
		if !ok {
			mpoint = &MountPath{}
			mountpathmap[mount.MountPoint] = mpoint
		}
		mnt := &MountPoint{Mountinfo: mount}
		mpoint.Mounts = append(mpoint.Mounts, mnt)
		mountidmap[mnt.MountID] = mnt
	}
	// Build the tree based on the mount paths seen: for each mount path go up
	// the path hierarchy until we find the nearest parent mount path. Then
	// create bidirectional references between the particular mount path and its
	// nearest parent mount path.
	for mountpath, mp := range mountpathmap {
		// Wire up child mount paths with their parent mount paths and vice
		// versa, but always based on the child mount path.
		for mountpath != "/" {
			mountpath = path.Dir(mountpath)
			if parentmpoint, ok := mountpathmap[mountpath]; ok {
				mp.Parent = parentmpoint
				parentmpoint.Children = append(parentmpoint.Children, mp)
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
	// To slightly optimize later operations, we now sort the list of child
	// mount points of each mount point by the length of their paths.
	for _, mount := range mountidmap {
		sort.SliceStable(mount.Children, func(i, j int) bool {
			return len(mount.Children[i].MountPoint) < len(mount.Children[j].MountPoint)
		})
	}
	// Now check mount points for being hidden either by in-place overmounts or
	// hiding mounts higher up the mount path hierarchy.
	mountpathmap["/"].Mounts[0].determineVisibility()
	// Finally: done.
	return
}

// determineVisbility determines whether a mount point is visible. It
// recursively checks child mount points.
func (p *MountPoint) determineVisibility() {
	childcount := len(p.Children)
	if childcount == 0 {
		return
	}
	// Have we been "overmounted" in place so that there is a child mount point
	// with the same mount path as ours? Please note that since the child mount
	// points are sorted by their path length, we know that if there is an
	// overmount at all, it has to be the first child mount point. And there can
	// only be at most one.
	if p.Children[0].MountPoint == p.MountPoint {
		p.Hidden = true
		// Transitively, all other child mount points of ours – except for the
		// overmount – will also be hidden.
		for idx := 1; idx < childcount; idx++ {
			p.Children[idx].hideMountpointSubtree()
		}
		p.Children[0].determineVisibility()
		return
	}
	// No overmount, so we now check if the child mounts of ours are hiding
	// other of our child mounts. Again, we rely on the children sorted by the
	// length of their mount paths to half the combinations to search.
	for childidx, childp := range p.Children {
		prefix := childp.MountPoint + "/" // actually, the path, but 'tis is Linux terminology.
		for idx := childidx + 1; idx < childcount; idx++ {
			if strings.HasPrefix(p.Children[idx].MountPoint, prefix) {
				p.Children[idx].hideMountpointSubtree()
			}
		}
	}
	// For the "surviving" children, the ones not already known to being hidden,
	// recursively check their visibility.
	for _, childp := range p.Children {
		if !childp.Hidden {
			childp.determineVisibility()
		}
	}
}

// hideMountpointSubtree recursively marks a particular mount point as hidden,
// as well as all its child mount points. Just to drive this point home: this
// works on the mount point subtree, not the mount path subtree.
func (p *MountPoint) hideMountpointSubtree() {
	p.Hidden = true
	for _, childmount := range p.Children {
		if !childmount.Hidden {
			childmount.hideMountpointSubtree()
		}
	}
}
