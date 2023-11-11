// Copyright 2023 Harald Albrecht.
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

import React, { ReactNode } from 'react'
import { BrowserRouter } from 'react-router-dom'

import type { Preview } from '@storybook/react'

import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'

import { lxknsLightTheme } from 'app/appstyles'
import { createTheme, ScopedCssBaseline, StyledEngineProvider, ThemeProvider } from '@mui/material'


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


const preview: Preview = {
    decorators: [
        (Story, context) => (
            <BrowserRouter basename=''>
                <StyledEngineProvider injectFirst>
                    <ThemeProvider theme={lightTheme} >
                        <ScopedCssBaseline>
                            <Story />
                        </ScopedCssBaseline>
                    </ThemeProvider>
                </StyledEngineProvider>
            </BrowserRouter>
        ),
    ],
}

export default preview
