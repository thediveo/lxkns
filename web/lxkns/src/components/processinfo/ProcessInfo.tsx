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

import React from 'react'

import clsx from 'clsx'

import { styled, Tooltip } from '@mui/material'

import ProcessIcon from 'icons/Process'
import Init1Icon from 'icons/Init1'
import { Process } from 'models/lxkns'
import ContainerInfo from 'components/containerinfo/ContainerInfo'
import CgroupInfo from 'components/cgroupinfo/CgroupInfo'


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
    }
}))

const ContainerInformation = styled(ContainerInfo)(() => ({
    marginRight: '0.5em',
}))

export const ProcessName = styled('span')(({ theme }) => ({
    fontStyle: 'italic',
    color: theme.palette.process,
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
 * The `ProcessInfo` component expects only a single property: the process to
 * render information about.
 */
export interface ProcessInfoProps {
    /** information about a discovered Linux OS process. */
    process: Process
    /** render only process name with PID and nothing else */
    short?: boolean
    /** optional CSS class name(s). */
    className?: string
}

/** 
 * The `ProcessInfo` component renders only (almost) minimal information about a
 * single Linux OS process to make it easily identifyable:
 *
 * - if associated with a container: container information (name, group).
 *
 * - name of the process, which is has been either set by the process itself, or
 *   has been derived from the process' command line. Please note that this
 *   component only renders the `name` field, so this has to be set.
 * - PID.
 * - cgroup path, if path is not empty.
 * - pause indication if process is freezing or has been frozen.
 *
 * On purpose, this component doesn't render more comprehensive information
 * (such as parent and children, et cetera), as it is to be used in concise
 * contexts, such as a single process tree node.
 *
 * Also in this spirit, this component doesn't render the cgroup-related
 * information in case the process belongs to a container.
 */
export const ProcessInfo = ({ process, short, className }: ProcessInfoProps) => {
    return !!process && (
        <ProcessInformation className={clsx(className, short && piShort)}>
            {process.container && <ContainerInformation container={process.container} />}
            <Tooltip title="process"><>
                {process.pid === 1 ? <Init1Icon className="init1" fontSize="inherit" /> : <ProcessIcon fontSize="inherit" />}
                <ProcessName>{process.name}</ProcessName>
                &nbsp;<span>({process.pid})</span>
            </></Tooltip>
            {!short && process.cpucgroup && process.cpucgroup !== "/" && !process.container 
                && <CgroupInfo busybody={process} />}
        </ProcessInformation>
    )
}

export default ProcessInfo
