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
import { Discovery, Namespace } from 'models/lxkns'
import { Typography } from '@mui/material'
import { TreeItem, TreeView } from '@mui/x-tree-view'
import { EngineInfo } from 'components/engineinfo'
import { compareEngines } from 'utils/engine'
import ProcessInfo from 'components/processinfo'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import { NamespaceInfo } from 'components/namespaceinfo'

const coll = new Intl.Collator(undefined, {
    numeric: true,
})

export interface ContainerTreeProps {
    /** lxkns discovery data */
    discovery: Discovery
    /** tree action */
    action: Action
}

export const ContainerTree = ({ discovery, /*action*/ }: ContainerTreeProps) => {
    const engineItemsMemo = useMemo(() => (
        Object.values(discovery.engines || {})
            .sort(compareEngines)
            .map(engine => {
                const keyid = `${engine.type}-${engine.pid}-${engine.api}`
                const cntrs = engine.containers
                    .sort((c1, c2) => coll.compare(c1.name, c2.name))
                    .map((cntr) => {
                        const proc = cntr.process
                        //const usernamespace = proc.namespaces['user']
                        return <TreeItem
                            className="containerprocess"
                            key={proc.pid}
                            nodeId={`${proc.pid}`}
                            label={<ProcessInfo process={proc} />}
                        >
                            {Object.entries(proc.namespaces)
                                .filter((entry): entry is [string, Namespace] => {
                                    const [, procns] = entry
                                    return !!procns
                                })
                                .sort(([, procns1], [, procns2]) => procns1.type.localeCompare(procns2.type))
                                .map(([nstype, procns]) => <TreeItem
                                    className="tenant"
                                    key={procns.nsid}
                                    nodeId={`${proc.pid}-${procns.nsid}`}
                                    label={
                                        <NamespaceInfo
                                            shared={procns === proc.parent?.namespaces[nstype]}
                                            noprocess={true}
                                            namespace={procns}
                                        />
                                    }
                                />)
                            }
                        </TreeItem>
                    })
                return <TreeItem
                    className="engine"
                    key={keyid}
                    nodeId={keyid}
                    label={<EngineInfo engine={engine} />}
                >{cntrs}</TreeItem>
            })
    ), [discovery])

    return (
        (engineItemsMemo.length &&
            <TreeView
                className="containertree"
                defaultCollapseIcon={<ExpandMoreIcon />}
                defaultExpandIcon={<ChevronRightIcon />}
            >{engineItemsMemo}</TreeView>
        ) || (discovery.namespaces.length && (!discovery.containers || !Object.keys(discovery.containers).length) &&
            <Typography variant="body1" color="textSecondary">
                this Linux system doesn&apos;t have any container workloads
            </Typography>
        ) || (
            <Typography variant="body1" color="textSecondary">
                nothing discovered yet, please refresh
            </Typography>
        )
    )
}

export default ContainerTree
