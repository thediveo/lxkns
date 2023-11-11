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
import { Task } from 'models/lxkns'
import ContainerInfo from 'components/containerinfo'
import { ProcessName } from 'components/processinfo'
import CgroupInfo from 'components/cgroupinfo'


const tiShort = "short-taskinfo"

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
    [`&.${tiShort},&.${tiShort} *`]: {
        color: theme.palette.text.disabled,
    }
}))

const ContainerInformation = styled(ContainerInfo)(({ theme }) => ({
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
 * The `TaskInfo` component renders only (almost) minimal information about a
 * single Linux task to make it easily identifyable:
 *
 * - TID and thread name.
 * - if associated with a container: container information (name, group).
 * - process name and PID, which is has been either set by the process itself,
 *   or has been derived from the process' command line.
 * - cgroup path, if path is not empty.
 * - pause indication if process is freezing or has been frozen.
 *
 * On purpose, this component doesn't render more comprehensive information
 * (such as parent and children, et cetera), as it is to be used in concise
 * contexts, such as a single process tree node.
 */
export const TaskInfo = ({ task, short, className }: TaskInfoProps) => {

    const process = task && task.process

    return !!task && (
        <TaskInformation className={clsx(className, short && tiShort)}>
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
            {!short && task.cpucgroup && task.cpucgroup !== "/" && !process.container 
                && <CgroupInfo busybody={task} />}
        </TaskInformation>
    )
}

export default TaskInfo
