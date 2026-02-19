// Copyright 2022 Harald Albrecht.
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

import clsx from 'clsx'

import { styled, Tooltip } from '@mui/material'

import ThreadIcon from 'icons/Thread'
import type { Task } from 'models/lxkns'
import ContainerInfo from 'components/containerinfo'
import CgroupInfo from 'components/cgroupinfo'
import ProcessName from 'components/processname/ProcessName'
import SchedulerInfo from 'components/schedinfo'


const taskInfoClass = "short-taskinfo"

const TaskInformation = styled('span')(({ theme }) => ({
    fontWeight: theme.typography.fontWeightLight,
    display: 'inline-block',
    whiteSpace: 'nowrap',
    '& > .MuiSvgIcon-root': {
        marginRight: '0.15em',
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.2ex',
        color: theme.palette.task,
    },
    '& .init1': {
        color: theme.palette.init1,
    },
    [`&.${taskInfoClass},&.${taskInfoClass} *`]: {
        color: theme.palette.text.disabled,
    }
}))

const ContainerInformation = styled(ContainerInfo)(() => ({
    marginRight: '0.5em',
}))

const TaskName = styled('span')(({ theme }) => ({
    fontStyle: 'italic',
    color: theme.palette.task,
    '&::before': {
        content: '"«"',
        fontStyle: 'normal',
    },
    '&::after': {
        content: '"»"',
        fontStyle: 'normal',
        paddingLeft: '0.1em', // avoid italics overlapping with guillemet
    },
}))

/**
 * The `TaskInfo` component expects only a single property: the process to
 * render information about.
 */
export interface TaskInfoProps {
    /** information about a discovered Linux task (thread). */
    task: Task
    /** 
     * render only task name with TID and process name with PID, but nothing
     * else, and no cgroup and container information. The short format will also
     * use muted gray instead of different colors that otherwise differentiate
     * the individual information about the thread, its process, container, and
     * cgroup.
     */
    short?: boolean
    /** optional CSS class name(s). */
    className?: string
}

/** 
 * `TaskInfo` renders only certain information about a single Linux task to make
 * it easily identifyable:
 *
 * - **TID** (in square brackets instead of the usual round brackets for process
 *   PIDs) and **task/thread name**.
 * 
 * - when associated with a container: the **container name** and "**group**";
 *   this group can be a Compose project, a Kubernetes namespace, ... This
 *   information is hidden when `short=true`.
 * 
 * - the **PID and name of the process** to which the task belongs to.
 * 
 * - the **scheduling policy as well as priority/niceness** unless the short
 *   form is requested.
 * 
 * - the **cgroup path** unless the path empty (which means "we don't known").
 *   This information is hidden when `short=true`.
 * 
 * - a **pause indication** when process is freezing or has been frozen. This
 *   information is hidden when `short=true`.
 *
 * On purpose, this component doesn't render more comprehensive information
 * (such as parent and children, et cetera), as it is to be used in concise
 * contexts, such as a single process tree node.
 * 
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
export const TaskInfo = ({ task, short, className }: TaskInfoProps) => {

    const process = task && task.process

    return !!task && (
        <TaskInformation className={clsx(className, short && taskInfoClass)}>
            <ThreadIcon fontSize="inherit" />
            <Tooltip title="task"><>
                <TaskName>{task.name}</TaskName>
                &nbsp;[{task.tid}] of &nbsp;
            </></Tooltip>
            {process.container && <ContainerInformation container={process.container} />}
            <Tooltip title="task"><>
                <ProcessName>{process ? process.name : ''}</ProcessName>
                &nbsp;<span>({process.pid})</span>
            </></Tooltip>
            {!short && <> <SchedulerInfo process={task} /></>}
            {!short && task.cpucgroup && task.cpucgroup !== "/" && !process.container 
                && <CgroupInfo busybody={task} />}
        </TaskInformation>
    )
}

export default TaskInfo
