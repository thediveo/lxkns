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

import { amber, lightBlue, blue, blueGrey, brown, green, grey, lime, pink, purple, red, teal, lightGreen, orange, indigo } from '@mui/material/colors'
import { cloneDeep, merge as mergeDeep } from 'lodash'

import { lxknsLightTheme } from './light'

// The dark theme, based on the light theme; note that this is incomplete, so
// you normally want to use it with createTheme with the options object being at
// least "{ palette: { mode: 'dark' } }".
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
            background: {
                default: '#303030', // restore v4 palette
                paper: '#424242',
            },
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
            task: lime[400],
            container: lightBlue[300],
            cgroup: grey[500],
            ownername: lime[500],
            ownerroot: pink[500],
            fstype: grey[500],
            freezer: {
                run: green[500],
                froozen: red[700],
            },
            cpulist: grey[500],
            nice: lightGreen[500],
            notnice: orange[500],
        },
    }
)
