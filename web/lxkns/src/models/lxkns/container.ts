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
import type { Process } from './model'

export interface LabelMap { [key: string]: string }

/**
 * An alive Container managed by a container Engine, might be member of one or
 * more groups (such as composer projects, pods, et cetera).
 */
export interface Container {
    id: string
    name: string
    type: string
    flavor: string
    pid: number
    paused: boolean
    labels: LabelMap
    groups: Group[]
    engine: Engine
    process: Process
}

export const containerGroup = (container: Container, typeorflavor: string): Group | undefined => {
    if (container) {
        return container.groups.find(
            group => (group.flavor === typeorflavor || group.type === typeorflavor))
    }
    return undefined
}

/**
 * A (container) Engine managaging a bunch of Containers.
 */
export interface Engine {
    id: string
    type: string
    api: string
    pid: number
    containers: Container[]
}

/**
 * A Group of containers, such as a composer project, pod, et cetera.
 */
export interface Group {
    name: string
    type: string
    flavor: string
    containers: Container[]
    labels: LabelMap
}
