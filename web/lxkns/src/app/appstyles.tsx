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

import { createMuiTheme } from '@material-ui/core'
import { fade } from '@material-ui/core/styles/colorManipulator'
import { cloneDeep, merge as mergeDeep } from 'lodash'


// We augment the existing Material-UI theme with new elements for uniform color
// styling of lxkns UI elements beyond the predefined Material UI elements. This
// avoids scattering and potentially duplicating the same color configurations
// all over the various lxkns-specific UI elements.
//
// See also:
// https://medium.com/javascript-in-plain-english/extend-material-ui-theme-in-typescript-a462e207131f
declare module '@material-ui/core/styles/createPalette' {
    interface Palette {
    }
    // allow configuration using `createMuiTheme`
    interface PaletteOptions {
    }
}

// FIXME: remove and refactor CSS into components.
const globalTheme = createMuiTheme()


// The (basic) light theme parts specific to lxkns.
export const lxknsLightTheme = {
    overrides: {
        MuiCssBaseline: {
            '@global': {
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
                        color: globalTheme.palette.text.disabled,
                    },
                    '& .controlledprocess .controlledprocess': {
                        marginLeft: '1.1em',
                    },
                },
                // style to active route and hover for the drawer items to follow the theme.
                '.MuiListItem-root.Mui-selected, .MuiListItem-root.Mui-selected:hover': {
                    backgroundColor: [fade(globalTheme.palette.primary.dark, 0.1), '!important'],
                },
                '.MuiListItem-root:hover': {
                    backgroundColor: [fade(globalTheme.palette.primary.dark, 0.05), '!important'],
                },
                '.MuiAvatar-colorDefault': {
                    backgroundColor: [fade(globalTheme.palette.primary.light, 1), '!important'],
                }
            },
        },
    },
    palette: {
    },
}

// The dark theme, based on the light theme.
export const lxknsDarkTheme = mergeDeep(
    cloneDeep(lxknsLightTheme),
    {
        palette: {
        },
    }
)
