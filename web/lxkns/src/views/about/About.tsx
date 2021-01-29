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

import React from 'react'

import { Box } from '@material-ui/core'
import { MuiMarkdown } from 'components/muimarkdown'
import { SmartA } from 'components/smarta'

/* eslint import/no-webpack-loader-syntax: off */
import AboutMDX from "!babel-loader!mdx-loader!./About.mdx"


export const About = () => {

    return (
        <Box m={2} flex={1} overflow="auto">
            <MuiMarkdown mdx={AboutMDX} shortcodes={{a:SmartA}} />
        </Box>
    )
}