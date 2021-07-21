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

import { Container, Engine, Group } from './container'
import { Namespace, NamespaceType, Process, Discovery, EngineMap, GroupMap } from './model'
import { insertCommonChildPrefixMountPaths, MountGroupMap, MountPath, MountPoint } from './mount'

// There are things in *type*script that really give me the creeps, not least
// being able to *omit* things from types. On the other hand, it's exactly what
// we need here when doing bad things, anyway... The reason is that we transform
// id-based references from the JSON into proper object references so we're
// later able to quickly chase around the information model.
interface NamespaceSetJson { [key: string]: Namespace | number }
interface NamespaceJson extends Omit<Namespace,
    'ealdorman' | 'leaders' | 'namespaces' | 'owner' | 'parent'
> {
    parent: Namespace | number
    ealdorman: Process | number
    leaders: (Process | number)[]
    owner: Namespace | number
    namespaces: NamespaceSetJson
}

interface ContainerJson extends Omit<Container,
    'engine' | 'groups'
> {
    engine: Engine | number
    groups: Group[] | number[]
}

/**
 * Post-processes a discovery response from the lxkns discovery service,
 * resolving namespace and process (cross) references into ordinary object
 * references which can be directly used. Also resolve the references between
 * containers, container engines, groups, and processes, where available.
 *
 * @param discoverydata JSON discovery response in form of plain JS objects.
 * **IMPORTANT:** the discovery data gets modified in place, so there's no copy
 * being made by fromjson(). If necessary, the caller is responsible for a deep
 * clone!
 */
export const fromjson = (discoverydata: any): Discovery => {
    const discovery = discoverydata as Discovery
    // Process all (hierarchical) namespaces in a first round to initialize
    // their hierarchical references to empty array references.
    Object.values(discovery.namespaces).forEach(ns => {
        switch (ns.type) {
            case NamespaceType.user:
                ns.tenants = []
            // falls through
            case NamespaceType.pid:
                ns.children = []
        }
    })

    // With the initialization done, now resolve various references from their
    // transfer representation into object references.
    Object.values(discovery.namespaces).forEach(ns => {
        // Replace leader PIDs with leader process object references ... if
        // there is a list of leader PIDs; otherwise, set an empty array.
        let leaders: Process[] = [];
        (ns as NamespaceJson).leaders
            && (ns as NamespaceJson).leaders.forEach((leader: number) => {
                if (leader in discovery.processes) {
                    leaders.push(discovery.processes[leader])
                }
            });
        ns.leaders = leaders;

        // Resolve ealdorman, if present; otherwise set it to null instead of
        // undefined.
        ns.ealdorman = ((ns as NamespaceJson).ealdorman &&
            discovery.processes[(ns as NamespaceJson).ealdorman as number]) || null

        // resolve namespace hierarchy references, if present.
        switch (ns.type) {
            case NamespaceType.user:
            case NamespaceType.pid:
                ns.parent = ((ns as NamespaceJson).parent &&
                    discovery.namespaces[(ns as NamespaceJson).parent as number]) || null
                ns.parent && ns.parent.children.push(ns) // ...billions of gothers crying
        }

        // resolve ownership, if applicable.
        if ((ns as NamespaceJson).owner) {
            ns.owner = discovery.namespaces[(ns as NamespaceJson).owner as number];
            ns.owner.tenants.push(ns)
        }
    });

    // Process all, erm, processes and convert and initialize reference fields
    // correctly.
    Object.values(discovery.processes).forEach(proc => {
        proc.parent = null
        proc.children = []
    });

    // Now with initial null references in places, resolve references,
    // wherever possible.
    Object.values(discovery.processes).forEach(proc => {
        // Resolve the parent-child relationships.
        if (proc.ppid.toString() in discovery.processes) {
            proc.parent = discovery.processes[proc.ppid]
            proc.parent.children.push(proc)
        }

        // Resolve the attached namespaces relationships.
        for (const [type, nsref] of Object.entries(proc.namespaces)) {
            proc.namespaces[type] = discovery.namespaces[nsref]
        }
    });

    // Try to figure out which namespaces are the initial namespaces...
    if (discovery.processes[1] && discovery.processes[2]) {
        // At least someone has put some effort into fooling us...
        Object.values(discovery.processes[1].namespaces).forEach(
            (ns: Namespace) => {
                if (ns.nsid in discovery.namespaces) {
                    discovery.namespaces[ns.nsid].initial = true
                }
            }
        )
    }

    // Now go for the discovered containers, container engines, and container
    // groups. First reset the back references from engines and groups to the
    // containers, as we will rebuild them as port of transforming the
    // references in the containers from ids into proper object references.
    if (discovery.containers) {
        discovery.engines = discovery['container-engines'] as EngineMap
        discovery['container-engines'] = undefined
        Object.entries(discovery.engines).forEach(([eid, engine]) => {
            engine.containers = []
        })
        discovery.groups = discovery['container-groups'] as GroupMap
        Object.entries(discovery.groups).forEach(([gid, group]) => {
            group.containers = []
        })
        discovery['container-groups'] = undefined
        Object.entries(discovery.containers).forEach(([pid, container]) => {
            // transform engine reference
            const engine = discovery.engines[(container as ContainerJson).engine as number]
            container.engine = engine
            engine.containers.push(container)
            // transform group references
            const groups: Group[] = []
            ;((container as ContainerJson).groups as number[]).forEach(gid => {
                const group = discovery.groups[gid]
                groups.push(group)
                group.containers.push(container)
            })
            container.groups = groups
            // link with process (if possible), keyed by PID
            const proc = discovery.processes[pid]
            if (proc) {
                proc.container = container
                container.process = proc
            }
        })
    }

    // Resolve the references in the hierarchy of mount paths and also resolve
    // the references in the hierarchy of mount points. For mount points,
    // references are allowed to cross mount namespaces.
    const mountpointidmap: { [mountpointid: string]: MountPoint } = {}
    const mountgroups: MountGroupMap = {}
    Object.entries(discovery.mounts).forEach(([mntnsid, mountpathmap]) => {
        const mountns = discovery.namespaces[mntnsid]
        // In order to later resolve the hierarchical mount point references
        // within a single mount namespace we need to first build a map from
        // mount path identifiers to mount path objects.
        const mountpathidmap: { [mountpathid: string]: MountPath } = {}
        Object.values(mountpathmap).forEach(mountpath => {
            mountpath.path = mountpath.mounts[0].mountpoint // convenience
            mountpathidmap[mountpath.pathid.toString()] = mountpath
            mountpath.children = []
            // While we're at it, let's also map the mount point IDs to their
            // respective mount point objects; please note that mount point IDs
            // are system-wide and thus across mount namespaces. Oh, and let's
            // "fix" the mount paths when they contain escaped characters, such
            // as \040 for space.
            mountpath.mounts.forEach(mountpoint => {
                mountpoint.mountnamespace = mountns
                mountpointidmap[mountpoint.mountid.toString()] = mountpoint
                mountpoint.children = []
            })
        })
        // With the map build we can now resolve the mount path hierarchy
        // references.
        Object.values(mountpathmap).forEach(mountpath => {
            const parent = mountpathidmap[mountpath.parentid.toString()]
            if (parent) {
                mountpath.parent = parent
                parent.children.push(mountpath)
            }
        })
    })

    // With the map from mount point IDs to mount point objects finally complete
    // we can now resolve the hierarchical references. And while we're at it, we
    // also build our map of mount point propagation groups.
    Object.values(discovery.mounts).forEach(mountpathmap => {
        Object.values(mountpathmap).forEach(mountpath => {
            mountpath.mounts.forEach(mountpoint => {
                // set up the parent-child relation object references.
                const parent = mountpointidmap[mountpoint.parentid.toString()]
                if (parent) {
                    mountpoint.parent = parent
                    parent.children.push(mountpoint)
                }
                // create the peer group if necessary and then set up the peer
                // group membership.
                const peergroupid = mountpoint.tags['shared']
                if (peergroupid) {
                    var peergroup = mountgroups[peergroupid]
                    if (!peergroup) {
                        peergroup = { id: parseInt(peergroupid), members: [] }
                        mountgroups[peergroupid] = peergroup
                    }
                    peergroup.members.push(mountpoint)
                    mountpoint.peergroup = peergroup
                }
                // create the peer group with our master(s) if necessary and
                // then set up our slave membership in the peer group. This
                // means that the peer group will contain both our masters
                // (which are peers to themselves) as well as their slaves.
                const mastergroupid = mountpoint.tags['master']
                if (mastergroupid) {
                    var mastergroup = mountgroups[mastergroupid]
                    if (!mastergroup) {
                        mastergroup = { id: parseInt(mastergroupid), members: [] }
                        mountgroups[mastergroupid] = mastergroup
                    }
                    mastergroup.members.push(mountpoint)
                    mountpoint.mastergroup = mastergroup
                }
            })
        })
    })

    // Insert mount path nodes for common subpath prefixes...
    Object.values(discovery.mounts).forEach(mountpathmap => {
        insertCommonChildPrefixMountPaths(mountpathmap['/'])
    })

    // Finally reference the corresponding map of mount points from its mount
    // namespace object.
    Object.entries(discovery.mounts).forEach(([mntnsid, mountpathmap]) => {
        discovery.namespaces[mntnsid].mountpaths = mountpathmap
    })

    // A small step for me, a huge misstep for type safety...
    return discovery
}

export default fromjson
