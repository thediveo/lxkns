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

import { Discovery } from 'models/lxkns'
import { Action } from 'app/treeaction'
import { UserNamespaceTree } from 'components/usernamespacetree'
import { Box } from '@material-ui/core'


export interface AllNamespacesProps {
    /** lxkns discovery data */
    discovery: Discovery
    /** tree action */
    action: Action
}

export const AllNamespaces = ({ discovery, action }: AllNamespacesProps) => (
    <Box pl={1}>
        <UserNamespaceTree discovery={discovery} action={action} />
    </Box>
)
