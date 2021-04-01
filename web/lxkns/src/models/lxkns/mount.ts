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

/**
 * map of mount namespace identifiers (inode numbers only) mapping to the
 * individual per-mount namespace mount path maps.
 */
export interface NamespacedMountPathMaps {
    [mntnsid: string]: MountPathMap
}

/**
 * map of mount paths mapping to one or more mount points each.
 */
export interface MountPathMap {
    [path: string]: MountPath
}

/**
 * Mount path with corresponding mount point(s); the same mount path can be used
 * by multiple mount points, but in the end only at most one of these mount
 * points can be visible. However, no mount point(s) might be visible at all if
 * this mount path has been completely overmounted.
 *
 * Important note: in contrast to the Linux kernel's mount point(!) hierarchy,
 * our mount path(!) hierarchy references an in-place overmounted mount point
 * with its overmounting (child) mount point into the mounts list of the same
 * mount path.
 */
export interface MountPath {
    /** (calculated in client) */
    path: string
    /** mount namespace-local ID of this mount path. (lxkns supplied) */
    pathid: number
    /** ID of parent mount path. (lxkns supplied) */
    parentid: number
    /** parent mount path. (calculated in client) */
    parent: MountPath
    /** child mount paths. (calculated in client) */
    children: MountPath[]
    /** mount points at this mount path */
    mounts: MountPoint[]
}

/**
 * Linux-kernel supplied information about an individual mount point.
 * Additionally features lxkns-determined visibility and mount point object
 * hierarchy references.
 */
export interface MountPoint {
    /** visibility of this mount point in the VFS. (lxkns supplied) */
    hidden: boolean
    /** unambiguous ID of this mount; IDs can be reused after unmounting. */
    mountid: number
    /** ID of parent mount. */
    parentid: number
    /** parent mount point object, if any. (calculated in client) */
    parent: MountPoint
    /** 
     * child mount points, including an optional in-place overmount.
     * (calculated in client) 
     */
    children: MountPoint[]
    /** major device number */
    major: number
    /** minor device number */
    minor: number
    /** 
     * the pathname of the directory in the filesystem which forms the root of
     * this mount.
     */
    root: string
    /** 
     * pathname of the mount point in the VFS, relative to the process's root
     * directory.
     */
    mountpoint: string
    /** per-mount options, see mount(2). */
    mountoptions: string[]
    /** tags with optional values, such as for mount propagation. */
    tags: MountTags
    /** filesystem type in the form "type[.subtype]" */
    fstype: string
    /** filesystem-specific information or "none". */
    source: string
    /** per-superblock options, see mount(2). */
    superoptions: string
}

/**
 * Dictionary of mount tags with values (where the tag values can be empty).
 */
export interface MountTags {
    [tag: string]: string
}

/**
 * Insert "fake" common mount path nodes into the hierarchy of mount points
 * whenever mount points have child mount paths sharing common prefix segments
 * and then reparent the affected child mount path nodes accordingly. The newly
 * inserted "fake" common mount path nodes lack any mount points as well as
 * mount path IDs (which are fake themselves anyway in the sense that they are
 * lxkns service-generated and not understood by the Linux kernel).
 *
 * @param mountpath mount path object to start from.
 */
export const insertCommonChildPrefixMountPaths = (mountpath: MountPath) => {
    // First, scan the mount paths of the child mount points to see if they
    // share common "starter" directories and map such starter directories to
    // their corresponding mount paths.
    const starters: { [starter: string]: MountPath[] } = {}
    const baseskip = mountpath.path !== "/" ? mountpath.path.length + 1 : 1
    mountpath.children.forEach(childmountpath => {
        const starter = starterDir(childmountpath.path.substr(baseskip))
        if (starter) {
            if (starter in starters) {
                starters[starter].push(childmountpath)
            } else {
                starters[starter] = [childmountpath]
            }
        }
    })
    // Never "reparent" a single child mount path, but only when there are
    // several ones sharing the same "starter" (prefix sub-) directory.
    const base = mountpath.path !== "/" ? mountpath.path + "/" : "/"
    Object.entries(starters).forEach(([starter, childmountpaths]) => {
        if (childmountpaths.length > 1) {
            // replace all children with same starter dir with a single common
            // child mount path node and then put all the original children
            // below the new intermediate node, the "new parent".
            const newparent = {
                path: base + starter,
                children: childmountpaths,
                mounts: [],
            } as MountPath
            childmountpaths.forEach(childmountpath => {
                childmountpath.parent = newparent
                const idx = mountpath.children.indexOf(childmountpath)
                mountpath.children.splice(idx, 1)
            })
            mountpath.children.push(newparent)
            // Recursively insert more common starter mount path nodes if
            // necessary.
            insertCommonChildPrefixMountPaths(newparent)
            // Squash consecutive fake single child path nodes into a single
            // node (that is, reparent again, but this time in the opposite
            // direction). We simply make use of recursion here in that the the
            // child path node will already have been checked by the time we end
            // up here, so we just need to drop the single child node if
            // necessary, taking over all its children, et voila!
            const singlechild = newparent.children.length === 1 ? newparent.children[0] : null
            if (singlechild && singlechild.mounts.length === 0) {
                newparent.path = singlechild.path
                newparent.children = singlechild.children
                newparent.children.forEach(child => child.parent = newparent)
            }
        } else {
            // Recursively check this child mount point for common paths of its
            // child mount paths.
            insertCommonChildPrefixMountPaths(childmountpaths[0])
        }
    })
}

/**
 * Returns only the first directory from a file path, without any
 * subdirectories.
 *
 * @param path file path that can be either absolute "/foo/bar" or relative
 *   "foo/bar" path.
 * @returns first directory element, or "" in case the path given is the root
 *   path "/".
 */
export const starterDir = (path: string) => {
    const start = path[0] === '/' ? 1 : 0
    const afterFirstDir = path.indexOf('/', start)
    return afterFirstDir > 0 ? path.substr(start, afterFirstDir - start) : path.substr(start)
}

/**
 * Unescapes a mount path as used in /proc/[PID]/mountinfo so it can be used for
 * unencumbered display.
 *
 * @param path mount path string with might contain Linux-kernel (octal) escapes
 *   in order to not break /proc/[PID]/mountinfo.
 * @returns mount path with octal escapes replaced by their corresponding ASCII
 *   codes.
 */
export const unescapeMountPath = (path: string) =>
    path.replace(/\\([0-1][0-7][0-7])/g,
        (match, octal) => String.fromCharCode(parseInt(octal, 8)))

/**
 * Returns a number indicating whether a first mount path comes before a second
 * mount point (<0), is the same (0), or comes after (>0). Order is defined on
 * the lexicographic order of the mount paths.
 * 
 * Note: use this compare function on the children of a mount path.
 * 
 * @param mp1 one mount path object
 * @param mp2 another mount path object
 */
export const compareMountPaths = (mp1: MountPath, mp2: MountPath) =>
    mp1.path.localeCompare(mp2.path)

export const compareMounts = (mp1: MountPoint, mp2: MountPoint) => {
    if (mp1.hidden !== mp2.hidden) {
        return mp1.hidden ? -1 : 1
    }
    return mp1.mountpoint.localeCompare(mp2.mountpoint)
}
