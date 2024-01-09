// Copyright 2021 Harald Albrecht.
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

import { styled, Theme, Tooltip } from '@mui/material'
import DoubleArrowIcon from '@mui/icons-material/DoubleArrow'
import { NamespaceIcon } from 'icons/Namespace'
import { GhostIcon } from 'icons/Ghost'
import { AngryghostIcon } from 'icons/Angryghost'

import { Namespace, ProcessMap } from 'models/lxkns'
import { keyframes } from '@mui/system'


const NamespaceReference = styled('span')(({ theme }) => ({
    display: 'inline-block',
    whiteSpace: 'nowrap',
    fontStyle: 'italic',
    fontWeight: theme.typography.fontWeightLight,
    color: theme.palette.nsref,
    '& .MuiSvgIcon-root': {
        marginRight: '0.15em',
        verticalAlign: 'middle',
    },
}))

// Define "Blinky"'s moves when the mouse pointer hovers over it...
const blinkyMoves = (theme: Theme) =>
    [
        { transform: 'translate(0, 0)' },
        { transform: 'translate(0, -.75ex) rotate3d(1, 0, 0, 30deg)' },
        { transform: 'translate(-.75ex, -.75ex) skew(10deg)' },
        { transform: 'translate(0, -.75ex) skew(-10deg)' },
        { transform: 'translate(.75ex, -.75ex) skew(-10deg)' },
        { transform: 'translate(.75ex, 0) rotate3d(1, 0, 0, 30deg)' },
        { transform: 'translate(.75ex, .75ex) rotate3d(1, 0, 0, 30deg)' },
        { transform: 'translate(0, .75ex) skew(10deg)' },
        { transform: 'translate(0, 0ex) rotate3d(1, 0, 0, 30deg)' },
        { transform: 'translate(-.75ex, 0ex) skew(10deg)' },
        { transform: 'translate(-.75ex, .75ex) rotate3d(1, 0, 0, 30deg)' },
        { transform: 'translate(0, .75ex) skew(-10deg)' },
        { transform: 'translate(.75ex, .75ex) skew(-10deg)' },
        { transform: 'transform: translate(0, .75ex) skew(10deg)' },
        { transform: 'transform: translate(0, 0) rotate3d(1, 0, 0, 30deg)' },
    ].reduce((keyframes, keyframe, keyframeno, keyframeslist) => ({
        ...keyframes,
        [`${Math.round(keyframeno * 1000 / (keyframeslist.length - 1)) / 10}%`]: {
            ...keyframe,
            color: !((keyframeno / 3) & 1) ? '#2121de' : theme.palette.nsref,
        },
    }), {})

const Blinky = styled(NamespaceReference)(({ theme }) => ({
    position: 'relative',
    verticalAlign: 'middle',
    width: '1.1em',
    height: '100%',
    animationPlayState: 'paused',
    '& .normal': {
        position: 'absolute',
        top: '.2ex',
        left: 0,
    },
    '& .angry': {
        position: 'absolute',
        top: '.2ex',
        left: 0,
        opacity: 0,
    },
    '&:hover': {
        animationPlayState: 'running',
        animation: `${keyframes(blinkyMoves(theme))} 2s ease infinite`,
    },
    '&:hover .normal': {
        opacity: 0,
    },
    '&:hover .angry': {
        opacity: 1,
    },
}))

const PathsSeparator = styled(DoubleArrowIcon)(() => ({
    paddingLeft: '0.2em',
}))

/**
 * Given a file system path returns the corresponding process name if
 * applicable, otherwise returns undefined.
 *
 * @param path file system path
 */
const ProcessNameOfProcPath = (path: string, processes?: ProcessMap) => {
    if (!processes || !path.startsWith('/proc/')) return
    const fields = path.split('/')
    if (fields.length < 3) return
    const process = processes[fields[2]]
    if (!process) return
    return process.name
}

// Returns the name of the process referenced by a procfs-based path, if
// available. Otherwise, returns undefined. 
const FancyProcessNameOfProcPath = (path: string, processes?: ProcessMap) => {
    const processName = ProcessNameOfProcPath(path, processes)
    return processName ? ` [${processName}]` : undefined
}

export interface NamespaceRefProps {
    /** namespace object, with type and reference information. */
    namespace: Namespace
    /** 
     * information about all processes; while optional, passing this process
     * information allows rendering the name of the process belonging to a
     * particular procfs-based path in addition to just the plain path.
     */
    processes?: ProcessMap,
    /** optional CSS class name(s) for the namespace reference component. */
    className?: string
}

/**
 * Renders information about how a particular Linux-kernel namespace can be
 * referenced by a filesystem path – when there is no process (and thus not any
 * process) using it. This component differentiates between
 * 1. a reference to the open file descriptor of some process (path is in the
 *    form of `/proc/$PID/fd/$FD`), 
 * 2. all other *non-empty* paths (considered to be bind-mounted namespaces), 
 * 3. and no reference – *empty* path – at all.
 *
 * **Notes:**
 *
 * This component is not intended to render the process reference
 * `/proc/$PID/ns/$NSTYPE` of a namespace; it will not crash, but it will
 * declare it to be bind-mounted.
 *
 * Hierarchical namespaces – that is, PID or user namespaces – can actually be
 * "unreferenced" in some situations, when they're "inside" the hierarchy, but
 * neither a leaf nor root namespace. In this case, this component will not
 * render any reference path, but a ghost icon instead. In such cases they can
 * only be referenced as the parent of another (PID or user) namespace, as the
 * Linux kernel only has an [`NS_GET_PARENT`
 * `ioctl()`](https://man7.org/linux/man-pages/man2/ioctl_ns.2.html) for finding
 * the parent of a PID or user namespace. In lxkns, such "unreferenced"
 * namespaces thus don't have their own explicit reference, so no path.
 */
export const NamespaceRef = ({ namespace, processes, className }: NamespaceRefProps) => {

    const isInProcfs = namespace.reference
        && namespace.reference.length === 1
        && namespace.reference[0].startsWith('/proc/')

    const isProcfdPath = isInProcfs && namespace.reference[0].includes('/fd/')

    // render a UI representation of the namespace reference element(s);
    // remember, there might be multiple namespace reference elements chained
    // together, such as first some mount namespace reference that must be
    // entered for the next element to become accessible, followed by the
    // "ultimate" namespace reference.
    const ref = (namespace.reference && namespace.reference[0] === '/proc/1/ns/mnt'
        ? namespace.reference.slice(1) : (namespace.reference || []))
        .map((refelement, idx) => [
            idx > 0 ? <PathsSeparator key={`pathssep-${idx}`} fontSize="inherit" /> : undefined,
            <span key={idx} className="bindmount">&ldquo;{refelement}&rdquo;</span>,
            FancyProcessNameOfProcPath(refelement, processes),
        ]).flat()

    return (
        ((!namespace.reference || !namespace.reference.length) &&
            <Tooltip title={`intermediate hidden ${namespace.type} namespace`}>
                <Blinky className={className}>
                    &nbsp;
                    <GhostIcon className="normal" fontSize="inherit" />
                    <AngryghostIcon className="angry" fontSize="inherit" />
                </Blinky>
            </Tooltip>
        ) || (isProcfdPath &&
            <Tooltip title={`${namespace.type} namespace kept alive only by file descriptor`}>
                <NamespaceReference className={className}>
                    <NamespaceIcon fontSize="inherit" />
                    <span className="file-descriptor">{ref}</span>
                </NamespaceReference>
            </Tooltip>
        ) || (isInProcfs && !isProcfdPath &&
            <Tooltip title={`${namespace.type} namespace kept alive by process/task`}>
                <NamespaceReference className={className}>
                    <NamespaceIcon fontSize="inherit" />
                    <span className="process-task-reference">{ref}</span>
                </NamespaceReference>
            </Tooltip>
        ) || (
            <Tooltip title={`bind-mounted ${namespace.type} namespace`}>
                <NamespaceReference className={className}>
                    <NamespaceIcon fontSize="inherit" />
                    <span className="file-descriptor">{ref}</span>
                </NamespaceReference>
            </Tooltip>
        )
    )
}
