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

// First create the "default" theme, so we can reuse some of its definitions in
// our additional styles. See also: https://stackoverflow.com/a/62453393
const globalTheme = createMuiTheme()

const lxknsTheme = createMuiTheme({
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
                }
            }
        }
    },
}, globalTheme)

export default lxknsTheme
