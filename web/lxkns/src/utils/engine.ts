// Copyright 2024 Harald Albrecht.
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

import type { Engine } from "models/lxkns"

const coll = new Intl.Collator(undefined, {
    numeric: true,
})

const engineTypeNames: { [key: string]: string } = {
    '': 'unknown',
    'containerd.io': 'containerd',
    'k8s.io/cri-api': 'CRI API',
    'cri-o.io': 'CRI-O',
    'docker.com': 'Docker',
    'podman.io': 'Podman',
}

export const engineTypeName = (type: string) => {
    const typename = engineTypeNames[type]
    return typename != '' ? typename : engineTypeNames['']
}

export const compareEngines = (eng1: Engine, eng2: Engine) => {
    const beforeAfter = coll.compare(engineTypeName(eng1.type), engineTypeName(eng2.type))
    if (beforeAfter != 0) {
        return beforeAfter
    }
    const beforeAfterPID = eng1.pid - eng2.pid
    if (beforeAfterPID != 0) {
        return beforeAfterPID
    }
    return coll.compare(eng1.id, eng2.id)
}
