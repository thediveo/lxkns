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

import { compareBusybodies, isProcess, isTask, type Busybody, type Discovery, type Process, type ProcessMap } from "models/lxkns"
import { Box, styled, Typography } from "@mui/material"
import type { TreeAPI } from "app/treeapi"
import { SimpleTreeView, TreeItem, type TreeViewItemId } from "@mui/x-tree-view"
import ProcessInfo from "components/processinfo"
import TaskInfo from "components/taskinfo"
import { useEffect, useImperativeHandle, useMemo, useRef, useState } from "react"
import CPUList from "components/cpulist"
import PinnedIcon from "icons/Pinned"
import PinnedBelowIcon from "icons/PinnedBelow"
import { numCPUs, sameAffinity } from "models/lxkns/affinity"
import CPUIcon from "icons/CPU"


const OnlineCore = styled('span')(() => ({
    '& > .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
    },
}))

// OffCPUInfo overrides the foreground color for child elements with higher
// priority.
const OffCPUInfo = styled('span')(({ theme }) => ({
    '& *': {
        color: theme.palette.text.disabled + "!important",
        opacity: '0.5 !important',
    },
}))

const UnpinnedInfo = styled('span')(() => ({
    '& *': {
        opacity: '0.75 !important',
    },
}))


/**
 * Information about a task or process that either has been pinned (restricted
 * to a subset of CPUs) or is an anchestor process up to the particular root
 * PID1 or PID2.
 */
type Executor = {
    /** true if on-CPU, false if anchestor on other CPU(s) */
    onCPU: boolean
    /** true if a subset of the online CPU(s) */
    pinned: boolean
    /** true if processes or tasks exists below that are using only a subset of CPUs */
    pinnedBelow: boolean
    /** process or task */
    busybody: Busybody
    /** further processes and tasks, might not be on-CPU if there's an process
     * or task further down the branch.
     */
    children?: Executor[]
}

/**
 * Mapping a PID or TID (same number space) to information about that task or
 * process on a particular logical CPU. While the name on purpose focuses on the
 * on-CPU/pinned tasks and processes, this map additionally contains information
 * about the process branches leading to these on-CPU tasks and threads.
 */
type PerCoreExecutors = { [key: number]: Executor }

/**
 * Mapping a logical CPU number to the PerCoreExecutors.
 */
type CoresWithExecutors = { [key: number]: PerCoreExecutors }

const pinnedExecutors = (processes: ProcessMap, onlineCPUs: number[][] | null) => {

    const coresWithExecutors: CoresWithExecutors = {}
    // Pour over all processes with all their tasks and pick up the tasks that
    // are allowed to run on a particular logical CPU.
    Object.values(processes).forEach((proc) => {
        Object.values(proc.tasks).forEach(task => {
            task?.affinity?.forEach(([from, to]) => {
                for (let cpu = from; cpu <= to; cpu++) {
                    // now that we're looking at a single logical CPU, ensure
                    // that we have a map for this particular logical CPU.
                    let coreExecutors = coresWithExecutors[cpu]
                    if (!coreExecutors) {
                        coreExecutors = coresWithExecutors[cpu] = {}
                    }

                    // this task is on this particular CPU (and might be as well
                    // on others). Now in the tree to be rendered we want to
                    // either want to add the process instead of its task group
                    // leader or if this task has the same CPU affinities as the
                    // task group leader.
                    const busybody = (task.tid === task.process.pid
                        || sameAffinity(task.affinity, task.process.affinity))
                        ? task.process : task
                    const xid = pidtid(busybody)
                    let lowerexec = coreExecutors[xid]
                    if (!lowerexec) {
                        lowerexec = coreExecutors[xid] = {
                            onCPU: true,
                            pinned: !sameAffinity(busybody.affinity, onlineCPUs),
                            pinnedBelow: false,
                            busybody: busybody,
                        }
                    } else {
                        // We've already processed this task, but it might have
                        // been as part of a branch, so mark this as actually
                        // on-CPU and then call it a day.
                        lowerexec.onCPU = true
                        break
                    }

                    const pinned = !sameAffinity(busybody.affinity, onlineCPUs)
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
                            // don't forget to chain the crumbs! That is, the
                            // downward chain, a.k.a. path, so that we later can
                            // recursively render the tree view without further
                            // hunting.
                            if (!exec.children) exec.children = []
                            exec.children.push(lowerexec)
                            if (pinned) {
                                // when we're dealing with a task/process that
                                // has been pinned to a single specific CPU then
                                // make sure to set indicators along the
                                // breadcrumb path, so that we later can render
                                // node adornments whenever we're on a branch
                                // with single CPU pinning somewhere below.
                                while (exec && !exec.pinnedBelow) {
                                    exec.pinnedBelow = true
                                    proc = proc?.parent || null
                                    exec = coreExecutors[proc?.pid || 0]
                                }
                            }
                            break
                        }
                        exec = {
                            onCPU: false,
                            pinned: !sameAffinity(proc.affinity, onlineCPUs),
                            pinnedBelow: pinned,
                            busybody: proc,
                            children: [lowerexec],
                        }
                        coreExecutors[proc.pid] = exec
                        lowerexec = exec
                        proc = proc.parent
                    }
                }
            })
        })
    })
    return coresWithExecutors
}

/**
 * Return the PID/TID of a process or task; PIDs and TIDs are in the same PID
 * number space and always uniquely identify a process (that actually is a task
 * group leader) or a non-group leading task.
 * 
 * @param bb a Process or Task
 * @returns PID of Process or TID of Task
 */
const pidtid = (bb: Busybody) => isProcess(bb) ? bb.pid : bb.tid

/**
 * Return a tree item ID for a CPU item; CPU item IDs are guaranteed to not
 * clash with process/task item IDs.
 * 
 * @param cpu logical CPU number
 * @returns stable item ID (string)
 */
const cpuItemID = (cpu: number) => `cpu${cpu}`

/**
 * Return a tree item ID for a task/process on a particular logical CPU;
 * task/process item IDs are guaranteed to not clash with CPU item IDs.
 * 
 * @param pidtid PID or TID of a task/process
 * @param cpu logical CPU number
 * @returns stable item ID (string)
 */
const executorItemID = (pidtid: number, cpu: number) => `task${pidtid}-cpu${cpu}`

interface InfoProps {
    busybody: Busybody
    onCPU: boolean
    pinned: boolean
    pinnedBelow: boolean
}

const Info = ({ busybody, onCPU, pinned, pinnedBelow }: InfoProps) => {
    const bbInfo = isProcess(busybody)
        ? <ProcessInfo process={busybody} hideAffinity={true} />
        : <TaskInfo task={busybody} />
    const pinnedBelowAdornment = pinnedBelow ? <><PinnedBelowIcon fontSize="inherit" /> </> : ''
    if (!onCPU) {
        return <OffCPUInfo>{pinnedBelowAdornment}{bbInfo}</OffCPUInfo>
    }
    const details = <><CPUList cpus={busybody.affinity} noWrap showIcon tooltip="CPU affinity list" /> {bbInfo}</>
    return pinned
        ? <>{pinnedBelowAdornment}{details}</>
        : <>{pinnedBelowAdornment}<UnpinnedInfo>{details}</UnpinnedInfo></>
}

// Compare two executors A and B, returning <0 if A goes before B, >0 if A goes
// after B, otherwise 0; the order is determined as follows:
// - tasks go before processes.
// - more specific affinity (fewer CPUs) go before broader CPU love.
// - finally, order by names, then PIDs/TIDs.
const compareExecutors = (a: Executor, b: Executor) => {
    const isTaskA = isTask(a.busybody)
    const isTaskB = isTask(b.busybody)
    if (isTaskA != isTaskB) {
        return (+isTaskB) - (+isTaskA)
    }
    const affinityA = numCPUs(a.busybody.affinity)
    const affinityB = numCPUs(b.busybody.affinity)
    if (affinityA != affinityB) {
        return affinityA - affinityB
    }
    return compareBusybodies(a.busybody, b.busybody)
}

type execOnCoreHandler = (event: React.MouseEvent<HTMLDivElement>, cpu: number, tid: number) => void

// execsTreeOnCore recursively renders the process or task with the specified
// TID/PID as well as all its children leading up to pinned tasks.
const execsTreeOnCore = (cpu: number, xid: number, execsPerCore: PerCoreExecutors, onDoubleClick: execOnCoreHandler, onMouseDown: (event: React.MouseEvent<HTMLDivElement>) => void) => {
    const exec = execsPerCore[xid]
    if (!exec) {
        return <></>
    }
    const isPinned = exec.pinned
    return <TreeItem
        key={xid}
        itemId={executorItemID(xid, cpu)}
        label={<div style={{ display: 'inline-block', whiteSpace: 'nowrap' }}>
            {isPinned && <><PinnedIcon fontSize="inherit" /> </>}
            <Info busybody={exec.busybody} onCPU={exec.onCPU} pinned={isPinned} pinnedBelow={exec.pinnedBelow} />
        </div>}
        slotProps={{
            content: {
                onDoubleClick: (event: React.MouseEvent<HTMLDivElement>) => onDoubleClick(event, cpu, xid),
                onMouseDown: onMouseDown,
            }
        }}
    >{exec.children
        ?.sort((a, b) => compareExecutors(a, b))
        .map((exec) => execsTreeOnCore(cpu, pidtid(exec.busybody), execsPerCore, onDoubleClick, onMouseDown))
        }
    </TreeItem>
}

export interface AffinitiesProps {
    /** tree API for expansion, collapsing */
    apiRef?: React.Ref<TreeAPI>
    /* lxkns discovery data */
    discovery: Discovery
}

/**
 * The `Affinities` component renders a tree depicting which tasks and processes
 * are runnable on which logical CPUs. The necessary information is already part
 * of the discovery information.
 * 
 * - the top-most level consists of the logical CPUs (numbers) for which we
 *   found tasks and processes.
 * 
 * - immediately below are PID1 and PID2 nodes, where PID2 contains kernel
 *   threads (which are actually seen as processes) and PID1 contains
 *   "user-space" tasks and processes other than kernel threads.
 * 
 * - tasks and processes on a particular logical CPU are always shown in the
 *   context of their ancestry processes; even if some ancestors aren't on this
 *   particular CPU. Such ancestors are rendered in the "disabled" muted color
 *   as to visually differentiate them from tasks and processes that are
 *   actually affine to the CPU where they are rendered.
 * 
 * - the task group leader represents the process itself: it is thus never
 *   rendered as a task node but always as a process node.
 * 
 * - non-group leader tasks are only rendered if their affinities differ from
 *   the affinities of their task group leader.
 * 
 * - a task or process that is affine to only a single logical CPU shows a pin
 *   as a visual clue.
 */
export const Affinities = ({ apiRef, discovery }: AffinitiesProps) => {

    const pinnedExecutorsMemo = useMemo(() => pinnedExecutors(discovery.processes, discovery.onlineCPUs), [discovery])

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

    // return all the ids of the pinned task/process as well as of all ancestors
    // on their path. All other paths are ignored.
    const execSubtreeIds = (cpu: number, exec: Executor): string[] => {
        const ids = [executorItemID(pidtid(exec.busybody), cpu)]
        if (!exec.children || exec.children.length === 0) {
            return ids
        }
        return ids.concat(exec.children
            .filter((subexec) => subexec.pinnedBelow || subexec.pinned)
            .map(subexec => execSubtreeIds(cpu, subexec)).flat())
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
            event.preventDefault() // ♫ I'm the Great Preventer ♪♬
        }
    }

    // render all CPU nodes, as well as their pinned task child and grandchild
    // nodes.
    const cpus = Object.entries(pinnedExecutorsMemo).
        map(([cpu, tasksOnCores]) => [Number(cpu), tasksOnCores] as [number, PerCoreExecutors]).
        sort(([cpuA], [cpuB]) => cpuA - cpuB).
        map(([cpu, execsPerCore]) => {
            return <TreeItem
                key={cpu}
                itemId={cpuItemID(cpu)}
                label={<OnlineCore><CPUIcon fontSize="inherit" /> {cpu}</OnlineCore>}
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