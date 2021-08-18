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

import { Pause, PlayArrow } from '@material-ui/icons'
import { makeStyles, Tooltip } from '@material-ui/core'

import CgroupNamespace from 'icons/namespaces/Cgroup'
import ProcessIcon from 'icons/Process'
import Init1Icon from 'icons/Init1'
import { Process } from 'models/lxkns'
import ContainerInfo from 'components/containerinfo/ContainerInfo'


const useStyles = makeStyles((theme) => ({
    // The whole component as such...
    processInfo: {
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
        '&$shortInfo *': {
            color: theme.palette.text.disabled,
        }
    },
    containerInfo: {
        marginRight: '0.5em',
    },
    shortInfo: {
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
            paddingLeft: '0.1em', // avoid italics overlapping with guillemet
        },
    },
    pid: {
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
    cgroupIcon: {
        '&.MuiSvgIcon-root + .MuiSvgIcon-root': {
            marginLeft: '-0.2em',
        }
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
 *   has been derived from the process' command line.
 * - PID.
 * - cgroup path, if path is not empty.
 * - pause indication if process is freezing or has been frozen.
 *
 * On purpose, this component doesn't render more comprehensive information
 * (such as parent and children, et cetera), as it is to be used in concise
 * contexts, such as a single process tree node.
 */
export const ProcessInfo = ({ process, short, className }: ProcessInfoProps) => {

    const classes = useStyles()

    const fridge = process.fridgefrozen ?
        <Pause fontSize="inherit" /> : <PlayArrow fontSize="inherit" />

    return !!process && (
        <span className={clsx(classes.processInfo, className, short && classes.shortInfo)}>
            {process.container && <ContainerInfo container={process.container} className={classes.containerInfo} />}
            <Tooltip title="process"><>
                {process.pid === 1 ? <Init1Icon className="init1" fontSize="inherit" /> : <ProcessIcon fontSize="inherit" />}
                <span className={classes.processName}>{process.name}</span>
                &nbsp;<span className={classes.pid}>({process.pid})</span>
            </></Tooltip>
            {!short && process.cpucgroup && process.cpucgroup !== "/" && !process.container && (
                <Tooltip title="control-group path" className="cgroupinfo">
                    <span className={clsx(classes.cgroupInfo, className)}>
                        <CgroupNamespace className={classes.cgroupIcon} fontSize="inherit" />
                        {fridge}
                        <span className={classes.cgroupPath}>{process.cpucgroup}</span>
                    </span>
                </Tooltip>)}
        </span>
    )
}

export default ProcessInfo
