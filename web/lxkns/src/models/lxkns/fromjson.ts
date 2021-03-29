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

import { Namespace, NamespaceType, Process, Discovery } from './model'
import { MountPath, MountPoint } from './mount'

// There are things in *type*script that really give me the creeps, not least
// being able to *omit* things from types. On the other hand, it's exactly
// what we need here when doing bad things, anyway...
interface NamespaceSetJson { [key: string]: Namespace | number }
interface NamespaceJson extends Omit<Namespace, 'ealdorman' | 'leaders' | 'namespaces' | 'owner'> {
    ealdorman: Process | number
    leaders: (Process | number)[]
    owner: Namespace | number
    namespaces: NamespaceSetJson
}

/**
 * Post-processes a discovery response from the lxkns discovery service,
 * resolving namespace and process (cross) references into ordinary object
 * references which can be directly used.
 *
 * @param discoverydata JSON discovery response in form of plain JS objects.
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
        (ns as NamespaceJson).leaders && (ns as NamespaceJson).leaders.forEach(leader => {
            if (leader.toString() in discovery.processes) {
                leaders.push(discovery.processes[leader.toString()]);
            }
        });
        ns.leaders = leaders;

        // Resolve ealdorman, if present; otherwise set it to null instead of
        // undefined.
        ns.ealdorman = ((ns as NamespaceJson).ealdorman &&
            discovery.processes[(ns as NamespaceJson).ealdorman.toString()]) || null;

        // resolve namespace hierarchy references, if present.
        switch (ns.type) {
            case NamespaceType.user:
            case NamespaceType.pid:
                ns.parent = ((ns as NamespaceJson).parent &&
                    discovery.namespaces[(ns as NamespaceJson).parent.toString()]) || null;
                ns.parent && ns.parent.children.push(ns); // ...billions of gothers crying
        }

        // resolve ownership, if applicable.
        if ((ns as NamespaceJson).owner) {
            ns.owner = discovery.namespaces[(ns as NamespaceJson).owner.toString()];
            ns.owner.tenants.push(ns);
        }
    });

    // Process all, erm, processes and convert and initialize reference fields
    // correctly.
    Object.values(discovery.processes).forEach(proc => {
        proc.parent = null;
        proc.children = [];
    });

    // Now with initial null references in places, resolve references,
    // wherever possible.
    Object.values(discovery.processes).forEach(proc => {
        // Resolve the parent-child relationships.
        if (proc.ppid.toString() in discovery.processes) {
            proc.parent = discovery.processes[proc.ppid.toString()];
            proc.parent.children.push(proc);
        }

        // Resolve the attached namespaces relationships.
        for (const [type, nsref] of Object.entries(proc.namespaces)) {
            proc.namespaces[type] = discovery.namespaces[nsref.toString()]
        }
    });

    // Try to figure out which namespaces are the initial namespaces...
    if (discovery.processes[1] && discovery.processes[2]) {
        // At least someone has put some effort into fooling us...
        Object.values(discovery.processes[1].namespaces).forEach(
            // Make thousands of gophers cry in syntactic agony...
            ns => (discovery.namespaces[ns.nsid.toString()].initial = true)
        )
    }

    // Resolve the references in the hierarchy of mount paths and also resolve
    // the references in the hierarchy of mount points. For mount points,
    // references are allowed to cross mount namespaces.
    const mountpointidmap: {[mountpointid: string]: MountPoint} = {}
    Object.values(discovery.mounts).forEach(mountpathmap => {
        // In order to later resolve the hierarchical mount point references
        // within a single mount namespace we need to first build a map from
        // mount path identifiers to mount path objects.
        const mountpathidmap: {[mountpathid: string]: MountPath} = {}
        Object.values(mountpathmap).forEach(mountpath => {
            mountpathidmap[mountpath.pathid.toString()] = mountpath
            mountpath.children = []
            // While we're at it, let's also map the mount point IDs to their
            // respective mount point objects; please note that mount point IDs
            // are system-wide and thus across mount namespaces.
            mountpath.mounts.forEach(mountpoint => {
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
    // we can now resolve the hierarchical references.
    Object.values(discovery.mounts).forEach(mountpathmap => {
        Object.values(mountpathmap).forEach(mountpath => {
            mountpath.mounts.forEach(mountpoint => {
                const parent = mountpointidmap[mountpoint.parentid.toString()]
                if (parent) {
                    mountpoint.parent = parent
                    parent.children.push(mountpoint)
                }
            })
        })
    })

    // A small step for me, a huge misstep for type safety...
    return discovery
}

export default fromjson
