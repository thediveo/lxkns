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
            /* fall through */
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

        // resolve namspace hierarchy references, if present.
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

    // Process all, erm, processes and add object references for the hierarchy,
    // making navigation quick and easy.
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

    // A small step for me, a huge misstep for type safety...
    return discovery
}

export default fromjson
