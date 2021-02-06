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

import Tooltip from '@material-ui/core/Tooltip'

import RunFast from 'mdi-material-ui/RunFast'
import CarCruiseControl from 'mdi-material-ui/CarCruiseControl'

import { Process } from 'models/lxkns'

import { makeStyles } from '@material-ui/core'
import clsx from 'clsx'


const useStyles = makeStyles((theme) => ({
    // The whole component as such...
    processInfo: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        '& .MuiSvgIcon-root': {
            marginRight: '0.15em',
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.1ex',
            color: theme.palette.process,
        },
    },
    processName: {
        fontStyle: 'italic',
        color: theme.palette.process,
        '&::before': {
            content: '"«"',
            fontStyle: 'normal',
        },
        '&::after': {
            content: '"»"',
            fontStyle: 'normal',
        },
    },
    cgroupInfo: {
        marginLeft: '0.5em',
        '& .MuiSvgIcon-root': {
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.1ex',
            color: theme.palette.cgroup,
        },
    },
    cgroupPath: {
        color: theme.palette.cgroup,
        '&::before': { content: '"«"' },
        '&::after': { content: '"»"' },
    }
}))

/**
 * The `ProcessInfo` component expects only a single property: the process to
 * render information about.
 */
export interface ProcessInfoProps {
    /** information about a discovered Linux OS process. */
    process: Process
    /** optional CSS class name(s). */
    className?: string
}

/** 
 * The `ProcessInfo` component renders only minimal information about a single
 * Linux OS process to make it easily identifyable:
 *
 * - name of the process, which is has been either set by the process itself,
 *   or has been derived from the process' command line.
 * - PID.
 * - cgroup path, if path is not empty.
 *
 * On purpose, this component doesn't render more comprehensive information
 * (such as parent and children, et cetera), as it is to be used in concise
 * contexts, such as a single process tree node.
 */
export const ProcessInfo = ({ process, className }: ProcessInfoProps) => {
    const classes = useStyles()

    return !!process && (
        <span className={clsx(classes.processInfo, className)}>
            <Tooltip title="process"><>
                <RunFast fontSize="inherit" />
                <span className={classes.processName}>{process.name}</span>
                &nbsp;({process.pid})
            </></Tooltip>
            {process.cgroup && (
                <Tooltip title="control-group path" className="cgroupinfo">
                    <span className={clsx(classes.cgroupInfo, className)}>
                        <CarCruiseControl fontSize="inherit" />
                        <span className={classes.cgroupPath}>{process.cgroup}</span>
                    </span>
                </Tooltip>)}
        </span>
    )
}

export default ProcessInfo
