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

import React, { useEffect, useMemo, useRef, useState } from 'react'

import { Action, COLLAPSEALL, EXPANDALL } from 'app/treeaction'
import { Discovery, Engine, Namespace } from 'models/lxkns'
import { Typography } from '@mui/material'
import { TreeItem, TreeView } from '@mui/x-tree-view'
import { EngineInfo } from 'components/engineinfo'
import { compareEngines } from 'utils/engine'
import ProcessInfo from 'components/processinfo'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import { NamespaceInfo } from 'components/namespaceinfo'
import { expandInitiallyAtom } from 'views/settings'
import { useAtom } from 'jotai'

const coll = new Intl.Collator(undefined, {
    numeric: true,
})

const enginekeyid = (engine: Engine) => `eng-${engine.type}-${engine.pid}`

export interface ContainerTreeProps {
    /** tree action */
    action: Action
    /** lxkns discovery data */
    discovery: Discovery
}

export const ContainerTree = ({ action, discovery }: ContainerTreeProps) => {

    const [expandInitially] = useAtom(expandInitiallyAtom)

    // Previous discovery information, if any.
    const previousDiscovery = useRef({ namespaces: {}, processes: {}, engines: {} } as Discovery)

    // Tree node expansion is a component-local state. We need to also use a
    // reference to the really current expansion state as for yet unknown
    // reasons setExpanded() will pass stale state information to its reducer.  
    const [expanded, setExpanded] = useState<string[]>([])
    const currExpanded = useRef<string[]>([])

    useEffect(() => { currExpanded.current = expanded }, [expanded])

    // Trigger an action when the action "state" changes; we are ignoring any
    // stuff appended to the commands, as we need to add noise to the commands
    // in order to make state changes trigger. Oh, well, bummer.
    useEffect(() => {
        const engines = discovery.engines || {}
        switch (action.action) {
            case EXPANDALL: {
                // expand all engines with their workload, as well as their
                // namespaces.
                const allengines = Object.values(engines)
                    .map(engine => enginekeyid(engine))
                const workloads = Object.values(engines)
                    .map(engine => engine.containers)
                    .flat()
                    .map(cntr => cntr.process.pid.toString())
                setExpanded(allengines.concat(workloads))
                break
            }
            case COLLAPSEALL: {
                const allengines = Object.values(engines)
                    .map(engine => enginekeyid(engine))
                setExpanded(allengines)
                break
            }
        }
    }, [action, discovery])

    useEffect(() => {
        const engines = discovery.engines || {}
        const previousEngines = previousDiscovery.current.engines || {}

        const expandEngineIds = Object.values(engines)
            .map(engine => enginekeyid(engine))
            .filter(id => !currExpanded.current.includes(id))

        const previousWorkloadIds = Object.values(previousEngines)
            .map(engine => engine.containers)
            .flat()
            .map(cntr => cntr.process.pid.toString())
        const expandWorkloadsIds = Object.values(engines)
            .map(engine => engine.containers)
            .flat()
            .map(cntr => cntr.process.pid.toString())
            .filter(id => expandInitially && !previousWorkloadIds.includes(id))

        setExpanded(currExpanded.current.concat(expandEngineIds, expandWorkloadsIds))
        previousDiscovery.current = discovery;
    }, [discovery, expandInitially])

    // Whenever the user clicks on the expand/close icon next to a tree item,
    // update the tree's expand state accordingly. This allows us to
    // explicitly take back control (ha ... hah ... HAHAHAHA!!!) of the expansion
    // state of the tree.
    const handleToggle = (event: React.SyntheticEvent, nodeIds: string[]) => {
        setExpanded(nodeIds)
    }

    const engineItemsMemo = useMemo(() => (
        Object.values(discovery.engines || {})
            .sort(compareEngines)
            .map(engine => {
                const keyid = enginekeyid(engine)
                const cntrs = engine.containers
                    .sort((c1, c2) => coll.compare(c1.name, c2.name))
                    .map((cntr) => {
                        const proc = cntr.process
                        const prockeyid = proc.pid.toString()
                        //const usernamespace = proc.namespaces['user']
                        return <TreeItem
                            className="containerprocess"
                            key={prockeyid}
                            nodeId={prockeyid}
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
                onNodeToggle={handleToggle}
                defaultCollapseIcon={<ExpandMoreIcon />}
                defaultExpandIcon={<ChevronRightIcon />}
                expanded={expanded}
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
