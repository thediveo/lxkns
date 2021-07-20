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

import { Namespace, Process } from './model'

/**
 * Returns a number indicating whether a first namespace comes before a second
 * namespace (<0), is the same (0), or comes after (>0). Namespaces are ordered
 * simply by their namespace identifiers, which are inode numbers. Initial
 * namespaces are always ordered before non-initial namespaces.
 *
 * @param ns1 one namespace.
 * @param ns2 another namespace.
 */
export const compareNamespaceById = (ns1: Namespace, ns2: Namespace) => {
    if (ns1.initial !== ns2.initial) {
        return ns1.initial ? -1 : 1
    }
    return ns1.nsid - ns2.nsid
}

/**
 * Returns a number indicating whether a first namespace comes before a second
 * namespace (<0), is the same (0), or comes after (>0). Namespaces are
 * ordered first based on their types, and then by their identifiers in case
 * of equal types.
 *
 * @param ns1 one namespace.
 * @param ns2 another namespace.
 */
export const compareNamespaceByTypeId = (ns1: Namespace, ns2: Namespace) => {
    const beforeAfter = ns1.type.localeCompare(ns2.type)
    return beforeAfter !== 0 ? beforeAfter : compareNamespaceById(ns1, ns2)
}

/**
 * Returns a number indicating whether a first namespace comes before a second
 * namespace (<0), is the same (0), or comes after (>0). Namespaces are
 * ordered first based on their reference path, then by type, and finally by
 * their identifiers.
 *
 * @param ns1 one namespace.
 * @param ns2 another namespace
 */
export const compareNamespaceByRefTypeId = (ns1: Namespace, ns2: Namespace) => {
    const beforeAfter = ns1.reference.localeCompare(ns2.reference)
    return beforeAfter !== 0 ? beforeAfter : compareNamespaceByTypeId
}

/**
 * Returns a number indicating whether a first process comes before a second
 * process (<0), is the same (0), or comes after (>0). Processes are ordered by
 * their names, taking the current locale settings into account. The only
 * exception is the initial process which always comes first. If the names of
 * two processes are the same ("bash", ...), then the processes identifiers
 * (PIDs) are compared instead to break the tie.
 *
 * @param proc1 one process.
 * @param proc2 another process.
 */
export const compareProcessByNameId = (proc1: Process, proc2: Process) => {
    if (proc1.pid === 1 || proc2.pid === 1) {
        return proc1.pid === 1 ? -1 : 1
    }
    const name1 = proc1.container ? proc1.container.name : proc1.name
    const name2 = proc2.container ? proc2.container.name : proc2.name
    return name1.localeCompare(name2) || proc1.pid - proc2.pid
}
