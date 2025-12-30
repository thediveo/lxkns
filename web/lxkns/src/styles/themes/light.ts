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

import { amber, lightBlue, blue, blueGrey, brown, green, grey, lime, pink, purple, red, teal, yellow, lightGreen, orange } from '@mui/material/colors'

// The (basic) light theme parts specific to lxkns; note that this is
// incomplete, so you normally want to use it with createTheme with the options
// object being at least "{ palette: { mode: 'light' } }".
export const lxknsLightTheme = {
    components: {
        MuiSelect: {
            defaultProps: {
                variant: 'standard', // MUI v4 default.
            },
        },
        MuiCssBaseline: {
            styleOverrides: {
                // Please note: now automatic translation into class names is
                // off, so don't forget the prefix dots on CSS class names.
                '.namespacetree': {
                    '& .MuiTreeItem-root': {
                        marginLeft: '2em',
                    },
                    '& .namespace .controlledprocess, & .namespace .controlledtask': {
                        marginLeft: '2em',
                    },
                    '& .namespace .controlledprocess .MuiTreeItem-content::before, & .namespace .controlledtask .MuiTreeItem-content::before': {
                        content: '"····"',
                        marginRight: '0.35em',
                        color: grey[500],
                    },
                    '& .controlledprocess .controlledprocess': {
                        marginLeft: '3em',
                    },
                    '& .controlledprocess .controlledprocess .MuiTreeItem-content::before': {
                        content: '""',
                    },
                },
                '.containertree': {
                    '& .MuiTreeItem-root': {
                        marginLeft: '24px',
                    },
                },
            },
        },
        MuiTreeItem: {
            styleOverrides: {
                content: {
                    padding: '0',
                },
            },
        },
    },
    palette: {
        background: {
            default: '#fafafa', // restore v4 palette
            paper: '#fff',
        },
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
        task: lime[800],
        cgroup: grey[600],
        ownername: lime[800],
        ownerroot: pink[700],
        fstype: grey[600],
        init1: amber[500],
        freezer: {
            run: green[500],
            froozen: red[900],
        },
        cpulist: grey[600],
        nice: lightGreen[700],
        notnice: orange[900],
        prio: red[400],
        relaxedsched: lightGreen[400],
        stressedsched: red[400],
    },
}
