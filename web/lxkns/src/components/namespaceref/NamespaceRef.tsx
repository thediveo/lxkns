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

import clsx from 'clsx'
import { makeStyles, Tooltip } from '@material-ui/core'
import { FileLinkOutline, Ghost } from 'mdi-material-ui'

import { Namespace } from 'models/lxkns'


const useStyles = makeStyles((theme) => ({
    namespaceReference: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        fontStyle: 'italic',
        color: theme.palette.nsref,
        '& .MuiSvgIcon-root': {
            marginRight: '0.15em',
            verticalAlign: 'middle',
        },
    },
    intermediate: {
    },
    fdref: {
    },
    bindmounted: {
    },
}))


export interface NamespaceRefProps {
    /** namespace object, with type and reference information. */
    namespace: Namespace
    /** optional CSS class name(s) for the badge. */
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
export const NamespaceRef = ({ namespace, className }: NamespaceRefProps) => {
    const classes = useStyles()

    const isProcfdPath = namespace.reference &&
        namespace.reference.startsWith('/proc/') &&
        namespace.reference.includes('/fd/')

    return (
        (!namespace.reference &&
            <Tooltip title={`intermediate hidden ${namespace.type} namespace`}>
                <span className={clsx(classes.namespaceReference, classes.intermediate, className)}>
                    <Ghost fontSize="inherit" />
                </span>
            </Tooltip>
        ) || (isProcfdPath &&
            <Tooltip title={`${namespace.type} namespace kept alive only by file descriptor`}>
                <span className={clsx(classes.namespaceReference, classes.fdref, className)}>
                    <FileLinkOutline fontSize="inherit" />
                    <span className="bindmount">"{namespace.reference}"</span>
                </span>
            </Tooltip>
        ) || (
            <Tooltip title={`bind-mounted ${namespace.type} namespace`}>
                <span className={clsx(classes.namespaceReference, classes.bindmounted, className)}>
                    <FileLinkOutline fontSize="inherit" />
                    <span className="bindmount">"{namespace.reference}"</span>
                </span>
            </Tooltip>
        )
    )
}
