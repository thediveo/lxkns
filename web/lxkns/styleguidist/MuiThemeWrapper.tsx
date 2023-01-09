// Copyright 2020 by Harald Albrecht.
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

import React from 'react'
//import { BrowserRouter as Router } from 'react-router-dom'
import { MemoryRouter as Router } from 'react-router'

import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'


import { lxknsLightTheme } from 'app/appstyles'
import { createTheme, ScopedCssBaseline, ThemeProvider } from '@mui/material'


const lightTheme = createTheme(
    {
        components: {
            MuiSelect: {
                defaultProps: {
                    variant: 'standard', // MUI v4 default.
                },
            },
        },
        palette: {
            mode: 'light',
        },
    },
    lxknsLightTheme,
)


const MuiThemeWrapper = ({ children }) => (
    <ThemeProvider theme={lightTheme}>
        <ScopedCssBaseline>
            {children}
        </ScopedCssBaseline>
    </ThemeProvider>
)

export default MuiThemeWrapper
