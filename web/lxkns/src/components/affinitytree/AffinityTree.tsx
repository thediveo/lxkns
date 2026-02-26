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
import { styled, Typography } from "@mui/material"
import type { TreeAPI } from "app/treeapi"
import { RichTreeView, TreeItemContent, TreeItemIconContainer, TreeItemProvider, TreeItemRoot, useTreeItem, useTreeItemModel, type TreeViewItemId, type UseTreeItemParameters } from "@mui/x-tree-view"
import ProcessInfo from "components/processinfo"
import TaskInfo from "components/taskinfo"
import React, { useEffect, useImperativeHandle, useMemo, useRef, useState } from "react"
import CPUList from "components/cpulist"
import PinnedIcon from "icons/Pinned"
import PinnedBelowIcon from "icons/PinnedBelow"
import { numCPUs, sameAffinity } from "models/lxkns/affinity"
import CPUIcon from "icons/CPU"
import { ChevronRight, ExpandMore } from "@mui/icons-material"

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
        opacity: '0.85 !important',
    },
}))


/**
 * Information about a task or process that either has been pinned (restricted
 * to a subset of CPUs) or is an anchestor process up to the particular root
 * PID1 or PID2.
 */
type Runner = {
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
    children?: Runner[]
}

// sameScheduling returns true if both processes or tasks are subject to the
// same scheduling policy as well as scheduling priority or niceness.
const sameScheduling = (bbA: Busybody, bbB: Busybody) => {
    return (bbA?.policy || 0) === (bbB?.policy || 0)
        && (bbA?.priority || 0) === (bbB?.priority || 0) // assuming 0 is fine here
        && (bbA?.nice || 0) === (bbB?.nice || 0)
}

/**
 * Mapping a PID or TID (same number space) to information about that task or
 * process on a particular logical CPU. While the name on purpose focuses on the
 * on-CPU/pinned tasks and processes, this map additionally contains information
 * about the process branches leading to these on-CPU tasks and threads.
 */
type SameCPURunners = { [key: number]: Runner }

/**
 * Mapping a logical CPU number to the map and hierarchy of processes/tasks on
 * that particular CPU.
 */
type RunnersByCPU = { [key: number]: SameCPURunners }

/**
 * mapRunners returns the mapping and hierarchy of processes and tasks on
 * particular CPUs.
 * 
 * @param processes map of all processes and tasks, with their affinities.
 * @param onlineCPUs list of online CPUs, as a series of ranges.
 * @returns map indexed by logical CPU number to the process and task hierarchy
 *   on this cpu.
 */
const mapRunners = (processes: ProcessMap, onlineCPUs: number[][] | null) => {
    const runnersByCPU: RunnersByCPU = {}
    // Pour over all processes with all their tasks and pick up the tasks that
    // are allowed to run on a particular logical CPU.
    Object.values(processes).forEach(proc => {
        Object.values(proc.tasks).forEach(task => {
            task?.affinity?.forEach(([cpuFrom, cpuTo]) => {
                for (let cpu = cpuFrom; cpu <= cpuTo; cpu++) {
                    // now that we're looking at a single logical CPU, ensure
                    // that we have a map for this particular logical CPU.
                    const runnersOnCPU = runnersByCPU[cpu] ??= {}

                    // So we've got a task but we don't want to totally visually
                    // clutter the tree with tasks of processes which all have
                    // the same properties regarding at least their CPU affinity
                    // and scheduling configuration.
                    //
                    // Thus we default to showing only the process of a task,
                    // UNLESS:
                    // - the task has different CPU affinities than its process;
                    // - or, the task has different scheduling parameters than
                    //   its process.
                    const busybody = task.tid === task.process.pid
                        || (sameAffinity(task.affinity, task.process.affinity)
                            && sameScheduling(task, task.process))
                        ? task.process : task
                    const xid = xidOf(busybody)
                    let runner = runnersOnCPU[xid]
                    if (!runner) {
                        runner = runnersOnCPU[xid] = {
                            onCPU: true,
                            pinned: !sameAffinity(busybody.affinity, onlineCPUs),
                            pinnedBelow: false,
                            busybody: busybody,
                        }
                    } else {
                        // We've already processed this task, but it might have
                        // been as part of a branch, so mark this as actually
                        // on-CPU and then call it a day.
                        runner.onCPU = true
                        break
                    }

                    // walk up the process tree to collect breadcrumbs. As we start
                    // back-tracking our beginning is either already a process or a
                    // task. In case of a task we proceed from the task's process, but
                    // in case of the single-task process we've already mapped the
                    // process, so we don't want to do this twice.
                    let proc: Process | null = isProcess(busybody) ? busybody.parent : busybody.process
                    const pinned = !sameAffinity(busybody.affinity, onlineCPUs)
                    while (proc) {
                        // have we reached a parent that we've already mapped?
                        let procRunner = runnersOnCPU[proc.pid]
                        if (procRunner) {
                            // don't forget to chain the crumbs! That is, the
                            // downward chain, a.k.a. path, so that we later can
                            // recursively render the tree view without further
                            // hunting.
                            if (!procRunner.children) procRunner.children = []
                            procRunner.children.push(runner)
                            if (pinned) {
                                // when we're dealing with a task/process that
                                // has been pinned to a single specific CPU then
                                // make sure to set indicators along the
                                // breadcrumb path, so that we later can render
                                // node adornments whenever we're on a branch
                                // with single CPU pinning somewhere below.
                                while (procRunner && !procRunner.pinnedBelow) {
                                    procRunner.pinnedBelow = true
                                    proc = proc?.parent || null
                                    procRunner = runnersOnCPU[proc?.pid || 0]
                                }
                            }
                            break
                        }
                        procRunner = {
                            onCPU: false,
                            pinned: !sameAffinity(proc.affinity, onlineCPUs),
                            pinnedBelow: pinned,
                            busybody: proc,
                            children: [runner],
                        }
                        runnersOnCPU[proc.pid] = procRunner
                        runner = procRunner
                        proc = proc.parent
                    }
                }
            })
        })
    })
    return runnersByCPU
}

/**
 * Return the PID/TID of a process or task; PIDs and TIDs are in the same PID
 * number space and always uniquely identify a process (that actually is a task
 * group leader) or a non-group leading task.
 * 
 * @param bb a Process or Task
 * @returns PID of Process or TID of Task
 */
const xidOf = (bb: Busybody) => isProcess(bb) ? bb.pid : bb.tid

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
 * @param xit PID or TID of a task/process
 * @param cpu logical CPU number
 * @returns stable item ID (string)
 */
const runnerOnCPUItemID = (xit: number, cpu: number) => `task${xit}-cpu${cpu}`

// cpuTID extracts the logical CPU number and optionally a runner's TID (PID)
// from the specified item ID.
const itemCPUTID = (id: string): { cpu?: number, tid?: number } => {
    if (id.startsWith('task')) {
        const cpuIdx = id.indexOf('-cpu')
        if (cpuIdx < 0) return {}
        return {
            cpu: Number(id.slice(cpuIdx + 4)),
            tid: Number(id.slice(4, cpuIdx))
        }
    } else if (id.startsWith('cpu')) {
        return {
            cpu: Number(id.slice(3)),
        }
    }
    return {}
}

// CPUItem represents a logcial CPU in the affinity tree.
type CPUItem = {
    cpu: number
}

// RunnerItem represents a process or task in the affinity tree.
type RunnerItem = {
    runner: Runner
}

// AffinityTreeItem is either a CPU or a task/process, and each item might have
// child items in the affinity tree.
type AffinityTreeItem = (CPUItem | RunnerItem) & {
    /** unique ID (used by the rendering tree view) */
    id: string

    children?: AffinityTreeItem[]
}

// isCPUItem is a type guard returning true if an AffinityTreeItem is a
// LogicalCPU.
const isCPUItem = (item: AffinityTreeItem): item is Extract<AffinityTreeItem, CPUItem> =>
    (item as CPUItem).cpu !== undefined

// isRunnerItem is a type guard returning true if an AffinityTreeItem is a
// Runner, representing a process or task.
const isRunnerItem = (item: AffinityTreeItem): item is Extract<AffinityTreeItem, RunnerItem> =>
    (item as RunnerItem).runner !== undefined

// runnersBranch returns a runner's affinity tree item for the specified
// PID/TID, where this tree item recursively contains child runner items as
// discovered.
const runnersBranch = (cpu: number, xid: number, runnersOnCPU: SameCPURunners): Exclude<AffinityTreeItem, CPUItem> | null => {
    const runner = runnersOnCPU[xid]
    if (!runner) {
        return null
    }
    const children = runner.children
        ?.map(subexec => runnersBranch(cpu, xidOf(subexec.busybody), runnersOnCPU))
        .filter(item => !!item)
        .sort((a, b) => compareRunners(a.runner, b.runner))
    return {
        id: runnerOnCPUItemID(xidOf(runner.busybody), cpu),
        runner: runner,
        children: children,
    }
}

// returns the top-level list of logical CPU tree items, with the CPU tree items
// containing the PID2 and PID1 hierarchies of processes and tasks related to
// that CPU.
const toplevelAffinityItems = (runnersByCPU: RunnersByCPU) => {
    return Object.entries(runnersByCPU)
        .map(([cpu, tasksOnCores]) => [Number(cpu), tasksOnCores] as [number, SameCPURunners])
        .sort(([cpuA], [cpuB]) => cpuA - cpuB)
        .map(([cpu, execsPerCore]) => {
            return {
                id: cpuItemID(cpu),
                cpu: cpu,
                children: [
                    runnersBranch(cpu, 2, execsPerCore),
                    runnersBranch(cpu, 1, execsPerCore),
                ].filter(item => !!item),
            } satisfies AffinityTreeItem
        })
}

const treeItemId = (item: AffinityTreeItem) => item.id
const treeItemLabel = (item: AffinityTreeItem) => item.id
const treeItemChildren = (item: AffinityTreeItem) => item.children

/**
 * Properties required to render information about a process or task in its
 * per-CPU hierarchy.
 */
interface RunnerInfoProps {
    /** the process or task; this gives details such as name and scheduling */
    busybody: Busybody
    /** is this runner actually on the CPU or a parent process in the hierarchy
     * and off-CPU? 
     */
    onCPU: boolean
    /** is this runner pinned to a subset of CPUs including this one the
     * hierarchy is for? 
     */
    pinned: boolean
    /** is there a runner lower down in a branch that is pinned?  */
    pinnedBelow: boolean
}

/**
 * Component `BusybodyInfo` renders information about a process or task, including
 * pinning and on-CPU adornments.
 */
const RunnerInfo = ({ busybody, onCPU, pinned, pinnedBelow }: RunnerInfoProps) => {
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

/**
 * `RunnerItem` renders a tree item representing a process or task (aka
 * "runner").
 */
const RunnerItem = ({ runner }: { runner: Runner }) => {
    const isPinned = runner.pinned
    return <div style={{ display: 'inline-block', whiteSpace: 'nowrap' }}>
        {isPinned && <><PinnedIcon fontSize="inherit" /> </>}
        <RunnerInfo busybody={runner.busybody} onCPU={runner.onCPU} pinned={isPinned} pinnedBelow={runner.pinnedBelow} />
    </div>
}

// Custom item's custom double click handler type ;) gets just the logical CPU
// number if its a CPU item, otherwise if its a runner item, the handler gets
// passed both cpu and runner's TID (which might be a PID, who knows).
type onItemDoubleClickHandler = (event: React.MouseEvent<HTMLDivElement>, cpu?: number, tid?: number) => void

// AffinityTreeItemProps defines the properties including a reference passed to
// our customized tree item renderer.
interface AffinityTreeItemProps extends Omit<UseTreeItemParameters, 'rootRef'>,
    Omit<React.HTMLAttributes<HTMLLIElement>, 'onFocus'> {
    onItemDoubleClick?: onItemDoubleClickHandler
}

// AffinityTreeItem returns a heavily customized rendition of an affinity tree
// item, that can be either a logical CPU or a process/task item.
const AffinityTreeItem = ({ ref, ...props }: AffinityTreeItemProps & { ref?: React.Ref<HTMLLIElement> }) => {
    const { id, itemId, label, disabled, children, onItemDoubleClick, ...other } = props
    const {
        getContextProviderProps,
        getRootProps,
        getContentProps,
        getIconContainerProps,
        status,
    } = useTreeItem({ id, itemId, children, label, disabled, rootRef: ref })
    const item = useTreeItemModel<AffinityTreeItem>(itemId)!

    return <TreeItemProvider {...getContextProviderProps()}>
        <TreeItemRoot {...getRootProps(other)}>
            <TreeItemContent {...getContentProps({
                onDoubleClick: (event: React.MouseEvent<HTMLDivElement>) => {
                    if (onItemDoubleClick) {
                        const { cpu, tid } = itemCPUTID(itemId)
                        onItemDoubleClick(event, cpu, tid)
                    }
                }
            })}>
                <TreeItemIconContainer {...getIconContainerProps()}>
                    {status.expandable && (status.expanded ? <ExpandMore /> : <ChevronRight />)}
                </TreeItemIconContainer>
                {
                    isCPUItem(item)
                    && <OnlineCore><CPUIcon fontSize="inherit" /> {item.cpu}</OnlineCore>
                    || (isRunnerItem(item) && <RunnerItem runner={item.runner} />)
                }
            </TreeItemContent>
            {status.expandable && status.expanded && <ul>{children}</ul>}
        </TreeItemRoot>
    </TreeItemProvider>
}

// Compare two runners A and B, returning <0 if A goes before B, >0 if A goes
// after B, otherwise 0; the order is determined as follows:
// - tasks go before processes.
// - more specific affinity (fewer CPUs) go before broader CPU love.
// - finally, order by names, then PIDs/TIDs.
const compareRunners = (a: Runner, b: Runner) => {
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

export interface AffinityTreeProps {
    /** tree API for expansion, collapsing */
    apiRef?: React.Ref<TreeAPI>
    /* lxkns discovery data */
    discovery: Discovery
}

/**
 * The `AffinityTree` component renders a tree depicting which tasks and
 * processes are runnable on which logical CPUs. The necessary information is
 * already part of the discovery information.
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
export const AffinityTree = ({ apiRef, discovery }: AffinityTreeProps) => {

    const pinnedExecutorsMemo = useMemo(
        () => mapRunners(discovery.processes, discovery.onlineCPUs),
        [discovery])

    // Tree node expansion is a component-local state. We need to also use a
    // reference to the really current expansion state as for yet unknown
    // reasons setExpanded() will pass stale state information to its reducer.  
    const [expanded, setExpanded] = useState<string[]>([])

    const currOnlineCPUs = useRef<string[]>([])

    useImperativeHandle(apiRef, () => ({
        expandAll() { /* not supported because it can literally render your browser unusable */ },
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

    // return all the ids of the pinned task/process as well as of all ancestors
    // on their path. All other paths are ignored.
    const runnerSubtreeIds = (cpu: number, runner: Runner): string[] => {
        const ids = [runnerOnCPUItemID(xidOf(runner.busybody), cpu)]
        if (!runner.children || runner.children.length === 0) {
            return ids
        }
        return ids.concat(runner.children
            .filter((subrunner) => subrunner.pinnedBelow || subrunner.pinned)
            .map(subrunner => runnerSubtreeIds(cpu, subrunner)).flat())
    }

    // double clicking on the contents of a collapsed task/process node expands
    // it, together with all child and grandchild nodes. Double clicking on an
    // expanded task/process node collapses it.
    const handleItemDoubleClick = (event: React.MouseEvent<HTMLDivElement>, cpu: number, tid: number) => {
        event.preventDefault() // ♫ I'm the Great Preventer ♪♬

        if (tid === undefined) {
            // expand a CPU item...
            const execIds = Object.values(pinnedExecutorsMemo[cpu]).
                map(exec => runnerOnCPUItemID(xidOf(exec.busybody), cpu))
            setExpanded(expanded.concat(execIds, cpuItemID(cpu)))
            return
        }

        // expand or collapse a runner item...
        const idx = expanded.indexOf(runnerOnCPUItemID(tid, cpu))
        if (idx >= 0) {
            setExpanded([...expanded.slice(0, idx), ...expanded.slice(idx + 1)])
            return
        }
        setExpanded(expanded.concat(runnerSubtreeIds(cpu, pinnedExecutorsMemo[cpu][tid])))
    }

    // prevent the browser from selecting the node contents upon double
    // clicking, but allow first click etc.; this still allows click-dragging to
    // select content.
    const handleMouseDown = (event: React.MouseEvent<HTMLUListElement>) => {
        if (event.detail === 2) {
            event.preventDefault() // ♫ I'm the Great Preventer ♪♬
        }
    }

    return (
        (toplevelAffinityItems.length &&
            <RichTreeView
                className="affinitytree"
                items={toplevelAffinityItems(pinnedExecutorsMemo)}

                getItemId={treeItemId}
                getItemLabel={treeItemLabel}
                getItemChildren={treeItemChildren}

                slots={{
                    item: AffinityTreeItem
                }}

                slotProps={{
                    item: {
                        onItemDoubleClick: handleItemDoubleClick,
                    } as Partial<AffinityTreeItemProps>,
                }}

                expandedItems={expanded}
                expansionTrigger="iconContainer"

                onExpandedItemsChange={handleToggle}
                onMouseDown={handleMouseDown}
            />
        ) || (
            <Typography variant="body1" color="textSecondary">
                nothing discovered yet, please refresh
            </Typography>
        )
    )
}

export default AffinityTree
