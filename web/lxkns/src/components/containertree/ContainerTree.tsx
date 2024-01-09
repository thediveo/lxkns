// Copyright 2024 Harald Albrecht.
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

import { Action } from 'app/treeaction'
import { Discovery } from 'models/lxkns'
import { Typography } from '@mui/material'

export interface ContainerTreeProps {
    /** lxkns discovery data */
    discovery: Discovery
    /** tree action */
    action: Action
}

export const ContainerTree = ({ discovery, /*action*/ }: ContainerTreeProps) => {
    return (
        (discovery && "CONTAINERZ"
        ) || (
            <Typography variant="body1" color="textSecondary">
            nothing discovered yet, please refresh
        </Typography>
        )
    )
}

export default ContainerTree