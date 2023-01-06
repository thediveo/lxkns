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

import { Busybody, isTask, Namespace, Process, Task } from './model'

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
    const beforeAfter = ns1.reference.join(":").localeCompare(ns2.reference.join(":"))
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

export const compareBusybodies = (bb1: Busybody, bb2: Busybody) => {
    const id1 = isTask(bb1) ? bb1.tid : bb1.pid
    const id2 = isTask(bb2) ? bb2.tid : bb2.pid
    if (id1 === 1 || id2 === 1) {
        return id1 === 1 ? -1 : 1
    }
    const name1 = isTask(bb1) ? bb1.name : bb1.container ? bb1.container.name : bb1.name
    const name2 = isTask(bb2) ? bb2.name : bb2.container ? bb2.container.name : bb2.name
    return name1.localeCompare(name2) || id1 - id2
}

/**
 * Returns a number indicating whether a first task is older than a second task
 * (<0), the same age (0), or younger (>0).
 * 
 * @param task1 one task.
 * @param task2 another task.
 */
export const compareTaskByAge = (task1: Task, task2: Task) => {
    return task1.starttime - task2.starttime
}

/**
 * Returns a number indicating whether a first task comes before another task. A
 * task comes before another task if the name of the first task's name comes
 * before the second task's name. 
 * 
 * @param task1 one task.
 * @param task2  another task.
 */
export const compareTaskByNameId = (task1 : Task, task2: Task) => {
    return task1.name.localeCompare(task2.name) || task1.tid - task2.tid
}