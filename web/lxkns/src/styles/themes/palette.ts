// Copyright 2020, 2025 Harald Albrecht.
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

import '@mui/material/styles'

// We augment the existing Material-UI theme with new elements for uniform color
// styling of lxkns UI elements beyond the predefined Material UI elements. This
// avoids scattering and potentially duplicating the same color configurations
// all over the various lxkns-specific UI elements.
//
// See also:
// https://medium.com/javascript-in-plain-english/extend-material-ui-theme-in-typescript-a462e207131f
declare module '@mui/material/styles' {

    interface Palette {
        // namespace badge background colors
        namespace: {
            cgroup: string,
            ipc: string,
            mnt: string,
            net: string,
            pid: string,
            user: string,
            uts: string,
            time: string,
        },

        nsref: string,          // filesystem reference of a namespace color
        container: string,      // container information color
        process: string,        // process information (name&PID) color
        task: string            // task information color
        cgroup: string,         // process cgroup path color
        ownername: string,      // owner user name color
        ownerroot: string,      // owner user root color
        fstype: string,         // filesystem type color
        init1: string,          // PID1 icon color

        freezer: {
            run: string         // color for run icon.
            frozen: string      // color for pause icon.
        }

        cpulist: string         // CPU (affinity) list color
        nice: string            // nice nice value color
        notnice: string         // not-nice value color
        prio: string            // non-0/non-1 prio value color
        relaxedsched: string    // scheduler NORMAL/BATCH/IDLE color
        stressedsched: string   // scheduler FIFO/RR/DEADLINE color
    }

    // augment custom configuration elements when using `createTheme`
    interface PaletteOptions {
        namespace?: {
            cgroup?: string,
            ipc?: string,
            mnt?: string,
            net?: string,
            pid?: string,
            user?: string,
            uts?: string,
            time?: string,
        },

        nsref?: string,
        container?: string,
        process?: string,
        task?: string,
        cgroup?: string,
        ownername?: string,
        ownerroot?: string,
        fstype?: string,
        init1?: string,

        freezer?: {
            run?: string,
            frozen?: string,
        },

        cpulist?: string
        nice?: string
        notnice?: string
        prio?: string
        relaxedsched?: string
        stressedsched?: string
    }
}
