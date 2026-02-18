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

import clsx from 'clsx'

import { styled, Tooltip } from '@mui/material'

import ProcessIcon from 'icons/Process'
import Init1Icon from 'icons/Init1'
import type { Process } from 'models/lxkns'
import ContainerInfo from 'components/containerinfo/ContainerInfo'
import CgroupInfo from 'components/cgroupinfo/CgroupInfo'
import CPUList from 'components/cpulist/CPUList'
import SchedulerInfo from 'components/schedinfo/SchedulerInfo'
import ProcessName from 'components/processname/ProcessName'
import TuxIcon from 'icons/Tux'

const piShort = "short-processinfo"

const ProcessInformation = styled('span')(({ theme }) => ({
    fontWeight: theme.typography.fontWeightLight,
    display: 'inline-block',
    whiteSpace: 'nowrap',
    '& .MuiSvgIcon-root': {
        marginRight: '0.15em',
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.2ex',
        color: theme.palette.process,
    },
    '& .init1': {
        color: theme.palette.init1,
    },
    [`&.${piShort} *`]: {
        color: theme.palette.text.disabled,
    },
    '& .cpulist': {
        marginLeft: '0.4em',
    }
}))

const ContainerInformation = styled(ContainerInfo)(() => ({
    marginRight: '0.5em',
}))

/**
 * The `ProcessInfo` component expects only a single property: the process to
 * render information about.
 */
export interface ProcessInfoProps {
    /** information about a discovered Linux OS process. */
    process: Process
    /** render only process name with PID and nothing else */
    short?: boolean
    /** hide CPU affinity detail */
    hideAffinity?: boolean
    /** optional CSS class name(s). */
    className?: string
}

/** 
 * `ProcessInfo` renders certain information about a single Linux OS process to
 * make it easily identifyable (see
 * [TaskInfo](?path=/docs/process-taskinfo--docs) for displaying OS task-related
 * details):
 *
 * - when associated with a container: the **container name** and "**group**";
 *   this group can be a Compose project, a Kubernetes namespace, ... This
 *   information is hidden when `short=true`.
 *
 * - the **name of the process**, which is has been either set by the process
 *   itself, or has been derived from the process' command line. Please note
 *   that this component only renders the `process.name` field, so this has to
 *   be set.
 * 
 * - the **PID**.
 * 
 * - the **CPU affinity**, unless `hideAffinity=true` or `short=true`.
 * 
 * - the **cgroup path** unless the path empty (which means "we don't known").
 *   This information is hidden when `short=true`.
 * 
 * - a **pause indication** when process is freezing or has been frozen. This
 *   information is hidden when `short=true`.
 *
 * On purpose, this component doesn't render even more comprehensive information
 * (such as parent and children, et cetera), as it is to be used in concise
 * contexts, such as a single process tree node.
 *
 * Also in this spirit, this component doesn't render the cgroup-related
 * information in case the process belongs to a container.
 * 
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
export const ProcessInfo = ({ process, short, hideAffinity, className }: ProcessInfoProps) => {

    return !!process && (
        <ProcessInformation className={clsx(className, short && piShort)}>
            {process.container && <ContainerInformation container={process.container} />}
            <Tooltip title="process"><>
                {process.pid === 1 
                    ? <Init1Icon className="init1" fontSize="inherit" /> 
                    : process.pid === 2 || process.parent?.pid === 2 ? <TuxIcon fontSize="inherit" /> : <ProcessIcon fontSize="inherit" />}
                <ProcessName>{process.name}</ProcessName>
                &nbsp;<span>({process.pid})</span>
            </></Tooltip>
            {!(short || hideAffinity) && <CPUList cpus={process.affinity} noWrap showIcon tooltip="CPU affinity list" />}
            {!short && <SchedulerInfo process={process} />}
            {!short && process.cpucgroup && process.cpucgroup !== "/" && !process.container
                && <CgroupInfo busybody={process} />}
        </ProcessInformation>
    )
}

export default ProcessInfo
