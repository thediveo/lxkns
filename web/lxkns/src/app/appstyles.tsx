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

import { amber, lightBlue, blue, blueGrey, brown, green, grey, indigo, lime, pink, purple, red, teal, yellow } from '@mui/material/colors'
import { cloneDeep, merge as mergeDeep } from 'lodash'

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
        nsref: string, // filesystem reference of a namespace
        container: string, // container information
        process: string, // process information (name&PID)
        cgroup: string, // process cgroup path
        ownername: string, // owner user name
        ownerroot: string, // owner user root
        fstype: string, // filesystem type
        init1: string, // PID1 icon
    }
    // allow configuration using `createTheme`
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
        cgroup?: string,
        ownername?: string,
        ownerroot?: string,
        fstype?: string,
        init1?: string,
    }
}

// The (basic) light theme parts specific to lxkns.
export const lxknsLightTheme = {
    components: {
        MuiCssBaseline: {
            styleOverrides: {
                // Please note: now automatic translation into class names is
                // off, so don't forget the prefix dots on CSS class names.
                '.namespacetree': {
                    '& .MuiTreeItem-group': {
                        marginLeft: '2em',
                    },
                    '& .namespace .controlledprocess': {
                        marginLeft: '2em',
                    },
                    '& .namespace .controlledprocess .MuiTreeItem-content::before': {
                        content: '"····"',
                        marginRight: '0.35em',
                        color: grey[500],
                    },
                    '& .controlledprocess .controlledprocess': {
                        marginLeft: '1.1em',
                    },
                },
            },
        },
    },
    palette: {
        namespace: {
            cgroup: red[50],
            ipc: lime[50],
            mnt: blue[50],
            net: green[50],
            pid: purple[50],
            user: blueGrey[50],
            uts: brown[50],
            time: amber[50],
        },
        nsref: yellow[800],
        container: lightBlue[700],
        process: teal[700],
        cgroup: grey[600],
        ownername: lime[800],
        ownerroot: pink[700],
        fstype: grey[600],
        init1: amber[500],
    },
}

// The dark theme, based on the light theme.
export const lxknsDarkTheme = mergeDeep(
    cloneDeep(lxknsLightTheme),
    {
        components: {
            MuiCssBaseline: {
                styleOverrides: {
                    '.namespacetree': {
                        '& .namespace .controlledprocess .MuiTreeItem-content::before': {
                            color: grey[600],
                        },
                    },
                },
            },
        },
        palette: {
            primary: {
                main: indigo[400],
                light: indigo[300],
                dark: indigo[700],
            },
            namespace: {
                cgroup: red[900],
                ipc: lime[900],
                mnt: blue[900],
                net: green[900],
                pid: purple[900],
                user: blueGrey[700],
                uts: brown[700],
                time: amber[900],
            },
            process: teal[300],
            container: lightBlue[300],
            cgroup: grey[500],
            ownername: lime[500],
            ownerroot: pink[500],
            fstype: grey[500],
        },
    }
)
