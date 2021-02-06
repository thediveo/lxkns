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

import { Discovery, Namespace, NamespaceSet, NamespaceType, Process } from "models/lxkns"

const initialNs = Object.values(NamespaceType)
    .reduce((o, nstype, idx) => ({
        ...o,
        [nstype]: {
            nsid: 4026531835 + idx,
            type: nstype,
            reference: `/proc/1/ns/${nstype}`,
            children: [],
            parent: null,
            initial: true,
        } as Namespace
    }), {}) as NamespaceSet

initialNs.user = {
    ...initialNs.user,
    "user-id": 0,
    "user-name": 'root',
    tenants: Object.values(initialNs)
        .filter((ns: Namespace) => ns.type !== NamespaceType.user),
}

Object.values(initialNs).forEach((ns: Namespace) => {
    if (ns.type !== NamespaceType.user) {
        ns.owner = initialNs.user
    }
})


const initProc: Process = {
    pid: 1,
    ppid: 0,
    name: 'süsdemdee',
    cmdline: '/sbin/init',
    cgroup: '',
    starttime: 12,
    children: [],
    parent: null,
    namespaces: initialNs,
}

Object.values(initialNs).forEach((ns: Namespace) => {
    ns.ealdorman = initProc
    ns.leaders = [initProc]
})

const lxknsNs = {
    ...initialNs,
    ...[NamespaceType.ipc, NamespaceType.mnt, NamespaceType.net, NamespaceType.uts]
        .map((nstype, idx) => ({
            nsid: 4026531987 + idx,
            type: nstype,
            reference: `/proc/123456/ns/${nstype}`,
            children: [],
            parent: null,
            initial: false,
            owner: initialNs.user,
        } as Namespace))
        .reduce((o, ns) => ({
            ...o,
            [ns.type]: ns,
        }), {}),
}

const lxknsProc: Process = {
    pid: 123456,
    ppid: 1,
    name: 'lskns',
    cmdline: '/world/domination',
    cgroup: '/whale/3d054ed83e84450dcaeb3b823bc1bba910ea3d821cb0660521466b6c9b5e9ccd',
    starttime: 777666,
    children: [],
    parent: initProc,
    namespaces: lxknsNs,
}

initProc.children.push(lxknsProc)

Object.values(lxknsNs)
    .filter((ns: Namespace) => !ns.initial)
    .forEach((ns: Namespace) => {
        ns.ealdorman = lxknsProc
        ns.leaders = [lxknsProc]
        initialNs.user.tenants.push(ns)
    })

const chronicumMountNs = {
    nsid: 4026532123,
    type: NamespaceType.mnt,
    reference: `/run/snackd/ns/calcium.mnt`,
    children: [],
    ealdorman: null,
    leaders: [],
    owner: initialNs.user,
} as Namespace

initialNs.user.tenants.push(chronicumMountNs)

const looserProc: Process = {
    pid: 666,
    ppid: 1,
    name: 'looser',
    cmdline: '/bin/looser --world-furcapination',
    cgroup: '',
    starttime: 777666,
    children: [],
    parent: initProc,
    namespaces: initialNs,
}

initProc.children.push(looserProc)

const looserUserNs = {
    nsid: 4026532666,
    type: NamespaceType.user,
    reference: `/proc/666/ns/user`,
    parent: initialNs.user,
    children: [],
    "user-id": 65534,
    "user-name": 'looser',
    ealdorman: looserProc,
    leaders: [looserProc],
    tenants: [],
} as Namespace

initialNs.user.children.push(looserUserNs)
looserProc.namespaces.user = looserUserNs

export const discovery: Discovery = {
    namespaces: Object.values(initialNs)
        .concat(Object.values(lxknsNs))
        .concat(chronicumMountNs)
        .reduce((nsmap, ns: Namespace) => ({
            ...nsmap,
            [ns.nsid.toString()]: ns,
        }), {}),
    processes: [initProc, lxknsProc]
        .reduce((procmap, proc) => ({
            ...procmap,
            [proc.pid.toString()]: proc,
        }), {})
}
