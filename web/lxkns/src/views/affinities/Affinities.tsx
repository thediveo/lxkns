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

import type { Discovery, Process, ProcessMap, Task } from "models/lxkns"
import { Box } from "@mui/material"
import type { TreeAPI } from "app/treeapi"
import { SimpleTreeView, TreeItem } from "@mui/x-tree-view"
import CPUAffinityIcon from "icons/CPUAffinity"
import TaskInfo from "components/taskinfo"

type PinnedTask = {
    pinned: boolean
    task: Task
}
type TaskCorePinnings = { [key: number]: PinnedTask }

const pinnedTasks = (processes: ProcessMap) => {
    const tasksOnCores: TaskCorePinnings[] = []

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
            let coreTasks = tasksOnCores[cpu]
            if (!coreTasks) {
                coreTasks = tasksOnCores[cpu] = {} as TaskCorePinnings
            }

            // this task is pinned, so we put it into our per-CPU map as pinned,
            // no way around.
            coreTasks[task.tid] = {pinned: true, task: task}

            // add the branch of processes (that is, task group leaders) that
            // lead to this task; we first check if our task is already
            // representing the process itself: in this case we skip this task
            // group leader and go for the parent process/task group leader.
            let proc: Process | null = task.process
            if (task.tid === proc.pid) {
                proc = proc?.parent
            }
            while (proc) {
                if (coreTasks[proc.pid]) break
                coreTasks[proc.pid] = {pinned: false, task: proc.tasks[0]}
            }
        })
    })

    return tasksOnCores
}

export interface AffinitiesProps {
    /** tree API for expansion, collapsing */
    apiRef?: React.Ref<TreeAPI>
    /* lxkns discovery data */
    discovery: Discovery
}

//@ts-expect-error develop
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export const Affinities = ({ apiRef, discovery }: AffinitiesProps) => {
    const tasksOnCores = pinnedTasks(discovery.processes)

    const cpus = tasksOnCores.map((tasksOnCore, cpu) => {
        return <TreeItem
            key={cpu}
            itemId={`cpu${cpu}`}
            label={<span><CPUAffinityIcon fontSize="inherit" />{cpu}</span>}
        >{
                Object.entries(tasksOnCore).map(([tid, pinnedTask]) => {
                    return <TreeItem
                        key={tid}
                        itemId={`cpu${cpu}task${tid}`}
                        label={<TaskInfo task={pinnedTask.task}/>}
                    ></TreeItem>
                })
            }</TreeItem>
    })

    return (
        <Box pl={1}>
            <SimpleTreeView
                className="affinitytree"
            >{cpus}</SimpleTreeView>
        </Box>
    )
}