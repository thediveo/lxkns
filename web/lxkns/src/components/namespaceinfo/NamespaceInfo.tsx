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

import { Person } from '@material-ui/icons'

import Rude from 'icons/Root'

import { ProcessInfo } from 'components/processinfo'
import { Namespace, NamespaceType } from 'models/lxkns'

import { makeStyles } from '@material-ui/core'
import { NamespaceRef } from 'components/namespaceref'
import { NamespaceBadge } from 'components/namespacebadge'
import clsx from 'clsx'
import ChildrenIcon from 'icons/Children'


// Component styling...
const useStyles = makeStyles((theme) => ({
    namespace: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        verticalAlign: 'middle',
    },
    shared: {
        color: `${theme.palette.text.disabled} !important`,
    },
    procInfo: {
        marginLeft: '0.5em',
    },
    pathInfo: {
        marginLeft: '0.5em',
    },
    userchildrenInfo: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        marginLeft: '0.5em',
        '& .MuiSvgIcon-root': {
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.1ex',
        },
    },
    ownerInfo: {
        '& .MuiSvgIcon-root': {
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.1ex',
        },
    },
    ownerName: {
        color: theme.palette.ownername,
        '&.root': { color: theme.palette.ownerroot },
        '&::before': { content: '"«"' },
        '&::after': { content: '"»"' },
    },
}))

// Reduce function returning the (recursive) sum of children and grand-children
// plus this namespace itself.
const countNamespaceWithChildren = (sum: number, ns: Namespace) =>
    sum + ns.children.reduce(countNamespaceWithChildren, 1)


export interface NamespaceInfoProps {
    /** namespace with type, identifier and initial namespace indication. */
    namespace: Namespace,
    /** suppress rendering leader process information.  */
    noprocess?: boolean,
    /** is this a namespace shared with other leader processes? */
    shared?: boolean,
    /** optional CSS class name(s). */
    className?: string,
}

/**
 * Component `Namespace` renders information about a particular namespace. The
 * type and ID get rendered, as well as the most senior process with its name,
 * or alternatively a bind-mounted or fd reference.
 *
 * Please note: this component never renders any child namespaces (even if the
 * given namespace is either a PID or user namespace).
 */
export const NamespaceInfo = ({
    namespace, noprocess, shared, className
}: NamespaceInfoProps) => {

    const classes = useStyles()

    // If there is a leader process joined to this namespace, then prepare some
    // process information to be rendered alongside with the namespace type and
    // ID. Unless the process information is to be suppressed.
    const procinfo = !noprocess && namespace.ealdorman &&
        <ProcessInfo process={namespace.ealdorman} className={classes.procInfo} />

    // If there isn't any process attached to this namespace, prepare
    // information about bind mounts and fd references, if possible. This also
    // covers "hidden" (PID, user) namespaces which are somewhere in the
    // hierarchy without any other references to them anymore beyond the
    // parent-child references.
    const pathinfo = !namespace.ealdorman &&
        <NamespaceRef namespace={namespace} className={classes.pathInfo} />

    // For user namespaces also prepare ownership information: the user name as
    // well as the UID of the Linux user "owning" the user namespace.
    const ownerinfo = namespace.type === NamespaceType.user &&
        'user-id' in namespace &&
        <span className={classes.ownerInfo}>
            owned by {
                namespace['user-id'] ? <Person fontSize="inherit" /> : <Rude fontSize="inherit" />
            } UID {namespace['user-id']}
            {namespace['user-name'] && <>
                {' '}
                <span className={clsx(classes.ownerName, namespace['user-name'] === 'root' && 'root')}>
                    {namespace['user-name']}
                </span>
            </>}
        </span>

    // For PID and user namespaces determine the total number of children and
    // grandchildren.
    const childrenCount = [NamespaceType.pid, NamespaceType.user].includes(namespace.type) && !shared &&
        namespace.children.length > 0 &&
        <span className={classes.userchildrenInfo}>
            [<ChildrenIcon fontSize="inherit" />&#8239;{countNamespaceWithChildren(-1, namespace)}]
        </span>

    return (
        <span className={clsx(classes.namespace, namespace.type, shared && classes.shared, className)}>
            <NamespaceBadge namespace={namespace} shared={shared} />
            {childrenCount}
            {procinfo || pathinfo} {ownerinfo}
        </span>
    )
}
