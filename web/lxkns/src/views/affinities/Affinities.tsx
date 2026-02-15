// Copyright 2026 Harald Albrecht.
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

import { compareBusybodies, isProcess, type Busybody, type Discovery, type Process, type ProcessMap } from "models/lxkns"
import { Box, styled, Typography } from "@mui/material"
import type { TreeAPI } from "app/treeapi"
import { SimpleTreeView, TreeItem, type TreeViewItemId } from "@mui/x-tree-view"
import CPUAffinityIcon from "icons/CPUAffinity"
import ProcessInfo from "components/processinfo"
import TaskInfo from "components/taskinfo"
import { useEffect, useImperativeHandle, useMemo, useRef, useState } from "react"
import PinnedIcon from "icons/Pinned"

/**
 * Information about a task or process that either has been pinned or is an
 * anchestor process up to the particular root PID1 or PID2.
 */
type Executor = {
    pinned: boolean
    busybody: Busybody
    children?: Executor[]
}

/**
 * Mapping a PID or TID (same number space) to information about that task or
 * process on a particular logical CPU. While the name on purpose focuses on the
 * pinned tasks and processes, this map additionally contains information about
 * the process branches leading to these pinned tasks and threads.
 */
type CorePinnedExecutors = { [key: number]: Executor }

type PerCoreExecutors = { [key: number]: CorePinnedExecutors }

const pinnedExecutors = (processes: ProcessMap) => {
    const perCoreExecs: PerCoreExecutors = {}

    // Pour over all processes with all their tasks and pick up the tasks that
    // are pinned to a single logical CPU only.
    Object.values(processes).forEach((proc) => {
        Object.values(proc.tasks).forEach(task => {

            // we want to work only on tasks with single logical CPU affinity;
            // all others we'll ignore...
            const affinities = task?.affinity
            if (affinities?.length != 1) { return }
            const cpurange = affinities[0]
            const cpu = cpurange[0]
            if (cpu != cpurange[1]) { return }

            // now we're looking at a single logical CPU, ensure that we have a
            // map for this particular logical CPU.
            let coreExecutors = perCoreExecs[cpu]
            if (!coreExecutors) {
                coreExecutors = perCoreExecs[cpu] = {} as CorePinnedExecutors
            }

            // this task is pinned, so we put it into our per-CPU map as pinned,
            // no way around. However, if this task is the sole task in its
            // process, we map just the process instead of the task.
            const busybody = (task.process.tasks.length == 1) ? task.process : task
            let lowerexec = {
                pinned: true,
                busybody: busybody,
            }
            coreExecutors[task.tid] = lowerexec

            // walk up the process tree to collect breadcrumbs. As we start
            // back-tracking our beginning is either already a process or a
            // task. In case of a task we proceed from the task's process, but
            // in case of the single-task process we've already mapped the
            // process, so we don't want to do this twice.
            let proc: Process | null = isProcess(busybody) ? busybody.parent : busybody.process
            while (proc) {
                // have we reached a parent that we've already mapped?
                let exec = coreExecutors[proc.pid]
                if (exec) {
                    // don't forget to chain the crumbs!
                    if (!exec.children) exec.children = []
                    exec.children.push(lowerexec)
                    break
                }
                exec = {
                    pinned: false,
                    busybody: proc,
                    children: [lowerexec],
                }
                coreExecutors[proc.pid] = exec
                lowerexec = exec
                proc = proc.parent
            }
        })
    })

    return perCoreExecs
}

const pidtid = (bb: Busybody) => isProcess(bb) ? bb.pid : bb.tid

const cpuItemID = (cpu: number) => `cpu${cpu}`

const executorItemID = (tid: number, cpu: number) => `task${tid}-cpu${cpu}`

const OnlineCore = styled('span')(() => ({
    '& > .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
    },
}))

// UnpinnedInfo overrides the foreground color for child elements with higher
// priority.
const UnpinnedInfo = styled('span')(({ theme }) => ({
    '& *': {
        color: theme.palette.text.disabled + "!important",
    },
}))

const Info = ({ busybody, pinned }: { busybody: Busybody, pinned: boolean }) => {
    const info = isProcess(busybody)
        ? <ProcessInfo process={busybody} hideAffinity={pinned} />
        : <TaskInfo task={busybody} />
    return pinned ? info : <UnpinnedInfo>{info}</UnpinnedInfo>
}

type execOnCoreHandler = (event: React.MouseEvent<HTMLDivElement>, cpu: number, tid: number) => void

// execsTreeOnCore recursively renders the process or task with the specified
// TID/PID as well as all its children leading up to pinned tasks.
const execsTreeOnCore = (cpu: number, ptid: number, execsPerCore: CorePinnedExecutors, onDoubleClick: execOnCoreHandler, onMouseDown: (event: React.MouseEvent<HTMLDivElement>) => void) => {
    const exec = execsPerCore[ptid]
    if (!exec) {
        return <></>
    }
    return <TreeItem
        key={ptid}
        itemId={executorItemID(ptid, cpu)}
        label={<Info busybody={exec.busybody} pinned={exec.pinned} />}
        slotProps={{
            content: {
                onDoubleClick: (event: React.MouseEvent<HTMLDivElement>) => onDoubleClick(event, cpu, ptid),
                onMouseDown: onMouseDown,
            }
        }}
    >{
            exec.children
                ?.sort((a, b) => {
                    const isProcessA = isProcess(a.busybody)
                    const isProcessB = isProcess(b.busybody)
                    if (isProcessA != isProcessB) {
                        return (+isProcessB) - (+isProcessA)
                    }

                    return compareBusybodies(a.busybody, b.busybody)
                })
                .map((exec) => execsTreeOnCore(cpu, pidtid(exec.busybody), execsPerCore, onDoubleClick, onMouseDown))
        }</TreeItem>
}

export interface AffinitiesProps {
    /** tree API for expansion, collapsing */
    apiRef?: React.Ref<TreeAPI>
    /* lxkns discovery data */
    discovery: Discovery
}

export const Affinities = ({ apiRef, discovery }: AffinitiesProps) => {

    const pinnedExecutorsMemo = useMemo(() => pinnedExecutors(discovery.processes), [discovery])

    // Tree node expansion is a component-local state. We need to also use a
    // reference to the really current expansion state as for yet unknown
    // reasons setExpanded() will pass stale state information to its reducer.  
    const [expanded, setExpanded] = useState<string[]>([])
    const currExpanded = useRef<string[]>([])

    const currOnlineCPUs = useRef<string[]>([])

    useEffect(() => { currExpanded.current = expanded }, [expanded])

    useImperativeHandle(apiRef, () => ({
        expandAll() {
            const cpus = Object.keys(pinnedExecutorsMemo)
                .map((cpu) => cpuItemID(Number(cpu)))
            const lowerexecs = Object.entries(pinnedExecutorsMemo)
                .map(([cpuno, coreExecutors]) => {
                    const cpu = Number(cpuno)
                    return Object.values(coreExecutors)
                        .filter(exec => exec.children && exec.children.length > 0)
                        .map(exec => executorItemID(pidtid(exec.busybody), cpu))
                })
                .flat()
            setExpanded(cpus.concat(lowerexecs))
        },
        collapseAll() {
            const cpus = Object.keys(pinnedExecutorsMemo).map((cpu) => cpuItemID(Number(cpu)))
            setExpanded(cpus)
        },
    }))

    // ensure that the CPU nodes are always expanded when the discovery results
    // are updated; but don't force CPU nodes open that have been collapsed.
    useEffect(() => {
        const cpus = Object.keys(pinnedExecutorsMemo)
        const newCpuIds = cpus
            .filter(cpu => !currOnlineCPUs.current?.includes(cpu))
            .map(cpu => cpuItemID(Number(cpu)))
        currOnlineCPUs.current = cpus
        setExpanded(expanded => newCpuIds.concat(expanded))
    }, [pinnedExecutorsMemo])

    // Whenever the user clicks on the expand/close icon next to a tree item,
    // update the tree's expand state accordingly. This allows us to
    // explicitly take back control (ha ... hah ... HAHAHAHA!!!) of the expansion
    // state of the tree.
    const handleToggle = (_event: React.SyntheticEvent | null, nodeIds: Array<TreeViewItemId>) => {
        setExpanded(nodeIds)
    }

    // double clicking on the contents of a CPU node expands it, together with
    // all its child and grandchild nodes.
    const handleCoreExpand = (event: React.MouseEvent<HTMLDivElement>, cpu: number) => {
        event.preventDefault()
        const execIds = Object.values(pinnedExecutorsMemo[cpu]).
            map(exec => executorItemID(pidtid(exec.busybody), cpu))
        setExpanded(expanded.concat(execIds, cpuItemID(cpu)))
    }

    // return all the ids of the pinned task/process as well as of all children.
    const execSubtreeIds = (cpu: number, exec: Executor): string[] => {
        const ids = [executorItemID(pidtid(exec.busybody), cpu)]
        if (!exec.children || exec.children.length === 0) {
            return ids
        }
        return ids.concat(exec.children.map(subexec => execSubtreeIds(cpu, subexec)).flat())
    }

    // double clicking on the contents of a collapsed task/process node expands
    // it, together with all child and grandchild nodes. Double clicking on an
    // expanded task/process node collapses it.
    const handleExecExpandCollapse = (event: React.MouseEvent<HTMLDivElement>, cpu: number, tid: number) => {
        event.preventDefault() // ♫ I'm the Great Preventer ♪♬
        const idx = expanded.indexOf(executorItemID(tid, cpu))
        if (idx >= 0) {
            setExpanded([...expanded.slice(0, idx), ...expanded.slice(idx + 1)])
            return
        }
        setExpanded(expanded.concat(execSubtreeIds(cpu, pinnedExecutorsMemo[cpu][tid])))
    }

    // prevent the browser from selecting the node contents upon double
    // clicking; this still allows click-dragging to select content.
    const handleMouseDown = (event: React.MouseEvent<HTMLDivElement>) => {
        if (event.detail === 2) {
            event.preventDefault()
        }
    }

    // render all CPU nodes, as well as their pinned task child and grandchild
    // nodes.
    const cpus = Object.entries(pinnedExecutorsMemo).
        map(([cpu, tasksOnCores]) => [Number(cpu), tasksOnCores] as [number, CorePinnedExecutors]).
        sort(([cpuA,], [cpuB]) => cpuA - cpuB).
        map(([cpu, execsPerCore]) => {
            const hasUserspacePinnings = !!execsPerCore[1]?.children
            return <TreeItem
                key={cpu}
                itemId={cpuItemID(cpu)}
                label={<OnlineCore>
                    <CPUAffinityIcon fontSize="inherit" /> {cpu}
                    {hasUserspacePinnings && <> <PinnedIcon fontSize="inherit" /></>}
                </OnlineCore>}
                slotProps={{
                    content: {
                        onDoubleClick: (event: React.MouseEvent<HTMLDivElement>) => handleCoreExpand(event, cpu),
                        onMouseDown: handleMouseDown,
                    }
                }}
            >{[2, 1].map((tid) => execsTreeOnCore(cpu, tid, execsPerCore, handleExecExpandCollapse, handleMouseDown))}</TreeItem>
        })

    return (
        <Box pl={1}>
            {(cpus.length &&
                <SimpleTreeView
                    className="affinitytree"
                    onExpandedItemsChange={handleToggle}
                    expandedItems={expanded}
                    expansionTrigger="iconContainer"
                >{cpus}</SimpleTreeView>
            ) || (
                    <Typography variant="body1" color="textSecondary">
                        nothing discovered yet, please refresh
                    </Typography>
                )}
        </Box>
    )
}