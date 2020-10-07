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

/**
 * Namespace type identifier strings: these are de-facto definitions from the
 * Linux kernel and used, for instance, when reading namespace links or (bind)
 * mount information related to namespaces.
 */
export enum NamespaceType {
    cgroup = 'cgroup',
    ipc = 'ipc',
    mnt = 'mnt',
    net = 'net',
    pid = 'pid',
    user = 'user',
    uts = 'uts',
    time = 'time'
}

export interface Namespace {
    /** identifier of this namespace (an inode number, without a device id) */
    nsid: number
    /** 
     * type of namespace (in form of the well-known type strings 'mnt',
     * 'cgroup', et cetera.)
     */
    type: NamespaceType
    /** file system path for referencing this namespace, if any. */
    reference: string
    /** for non-user namespaces the owning user namespace, otherwise null. */
    owner: Namespace
    /** 
     * UID of the user which once created the user namespace in which
     * this namespace then was created later. 
     */
    'user-id': number
    /** 
     * name of the user which once created the user namespace in which this
     * namespace then was created later.
     */
    'user-name': string
    /** the most senior process joined to this namespace, from the set of
     * leader processes. */
    ealdorman: Process | null
    /** list of top-most processes joined to this namespace. */
    leaders: Process[]
    /** user and pid namespaces only: the parent namespace, otherwise null */
    parent: Namespace | null
    /** user and pid namespaces only: the child namespaces, otherwise []. */
    children: Namespace[]
    /** for user namespaces the owned (possessed?!) non-user namespaces. */
    tenants: Namespace[]
}

export interface NamespaceSet {
    cgroup: Namespace
    ipc: Namespace
    mnt: Namespace
    net: Namespace
    pid: Namespace
    user: Namespace
    uts: Namespace
    time: Namespace | null
}

export interface NamespaceMap { [key: string]: Namespace }

export interface Process {
    pid: number
    ppid: number
    parent: Process | null
    children: Process[]
    name: string
    cmdline: string
    starttime: number
    cgroup: string
    namespaces: NamespaceSet
}

export interface ProcessMap { [key: string]: Process }

export interface Discovery {
    namespaces: NamespaceMap
    processes: ProcessMap
}
