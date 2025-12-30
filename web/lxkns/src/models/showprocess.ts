// Copyright 2025 Harald Albrecht.
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

import type { Process } from "./lxkns"

/**
 * `showProcess` determines whether a given process should be rendered/shown,
 * based on the cpu cgroup path and the flag to show "system" processes at all.
 * A process is considered to be a system process when it has a PID other than 2
 * and either...
 * - is inside "system.slice", but not Docker-related.
 * - is inside "init.scope",
 * - is "user.slice" itself (so processes deeper down the user.slice will be
 *   shown).
 * 
 * `showProcess` thus helps reducing visual clutter/overload with too many
 * system-related details, unless opted in by the user.
 * 
 * @param process 
 * @param showSystemProcs @returns 
 */
export const showProcess = (process: Process, showSystemProcs: boolean) => {
    return showSystemProcs /* ...whatever */ ||
        (process.pid > 2 &&
            !(process.cpucgroup.startsWith('/system.slice/') &&
                !process.cpucgroup.startsWith('/system.slice/docker-')) &&
            !process.cpucgroup.startsWith('/init.scope/') &&
            process.cpucgroup !== '/user.slice' &&
            process.cpucgroup !== '/init' && process.cpucgroup !== '/init.slice')
    }
