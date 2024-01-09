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

import React, { useMemo } from 'react'

import { Action } from 'app/treeaction'
import { Discovery } from 'models/lxkns'
import { Typography } from '@mui/material'
import { TreeItem, TreeView } from '@mui/x-tree-view'

export interface ContainerTreeProps {
    /** lxkns discovery data */
    discovery: Discovery
    /** tree action */
    action: Action
}

export const ContainerTree = ({ discovery, /*action*/ }: ContainerTreeProps) => {
    const engineItemsMemo = useMemo(() => (
        Object.values(discovery.engines || {})
            .map(engine => {
                const keyid = engine.pid.toString() + engine.api
                return <TreeItem
                    key={keyid}
                    nodeId={keyid}
                    label={engine.type}
                ></TreeItem>
            })
    ), [discovery])

    console.log(discovery)
    return (
        (engineItemsMemo.length &&
            <TreeView
                className="containertree"
            >{engineItemsMemo}</TreeView>
        ) || (Object.keys(discovery.engines).length &&
            <Typography variant="body1" color="textSecondary">
                this Linux system doesn&apos;t have any container engines with workloads
            </Typography>
        ) || (
            <Typography variant="body1" color="textSecondary">
                nothing discovered yet, please refresh
            </Typography>
        )
    )
}

export default ContainerTree
