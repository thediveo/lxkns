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
 */
export interface MountPath {
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
 * Dictionary of mount tags with values (or "" values).
 */
export interface MountTags {
    [tag: string]: string
}

