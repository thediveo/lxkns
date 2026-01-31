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

import type { Parameters, Preview } from '@storybook/react-vite'

import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'

import { MemoryRouter } from "react-router-dom"
import { lxknsDarkTheme, lxknsLightTheme } from 'styles/themes'
import { StyledEngineProvider, ThemeProvider, createTheme } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'

// For themes:Themes exported constant, see:
// https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/theming/create.ts#L20
// and Themes interface type:
// https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/theming/create.ts#L7
import { themes } from 'storybook/theming'

const lightTheme = createTheme(
    { palette: { mode: 'light' } }, lxknsLightTheme)
const darkTheme = createTheme(
    { palette: { mode: 'dark' } }, lxknsDarkTheme)

export const parameters: Parameters = {
    docs: {
        theme: themes.normal, // use same theme as the surrounding parts
    },
}

const preview: Preview = {
    decorators: [
        (Story, context) => {
            // support theme changes; see also:
            // https://storybook.js.org/docs/essentials/backgrounds#configuration
            // and
            // https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/backgrounds/types.ts#L21
            //
            // However, the backgrounds.value is undocumented :(
            const isDark = (context.globals.backgrounds?.value || themes.normal.base) === 'dark'
            const theme = isDark ? darkTheme : lightTheme

            // support Story-individual MemoryRouter properties...
            const routerProps = context.parameters.routerProps || 
                { initialEntries: ["/"] }

            return (
                <StyledEngineProvider injectFirst>
                    <ThemeProvider theme={theme} >
                        <CssBaseline enableColorScheme />
                        <MemoryRouter {...routerProps}>
                            <div style={{background: theme.palette.background.paper}}>
                                <Story />
                            </div>
                        </MemoryRouter>
                    </ThemeProvider>
                </StyledEngineProvider>
            )
        },
    ],
}

export default preview
