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

import { Person } from '@mui/icons-material'

import Rude from 'icons/Root'

import { ProcessInfo } from 'components/processinfo'
import { compareBusybodies, type Namespace, NamespaceType, type ProcessMap } from 'models/lxkns'

import { NamespaceRef } from 'components/namespaceref'
import { NamespaceBadge } from 'components/namespacebadge'
import clsx from 'clsx'
import ChildrenIcon from 'icons/Children'
import { styled } from '@mui/material'
import TaskInfo from 'components/taskinfo/TaskInfo'


const namespaceShared = "shared-namespace"

const NamespaceInformation = styled('span')(({ theme }) => ({
    display: 'inline-block',
    whiteSpace: 'nowrap',
    verticalAlign: 'middle',

    [`&.${namespaceShared}`]: {
        color: `${theme.palette.text.disabled} !important`,
    },
}))

const PathInformation = styled(NamespaceRef)(() => ({
    marginLeft: '0.5em',
}))

const UserChildrenInfo = styled('span')(() => ({
    display: 'inline-block',
    whiteSpace: 'nowrap',
    marginLeft: '0.5em',
    '& .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
    },
}))

const OwnerInformation = styled('span')(({ theme }) => ({
    fontWeight: theme.typography.fontWeightLight,
    '& .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
    },
    '&.root .MuiSvgIcon-root': { color: theme.palette.ownerroot },
}))

const OwnerName = styled('span')(({ theme }) => ({
    color: theme.palette.ownername,
    '&.root': { color: theme.palette.ownerroot },
    '&::before': { content: '"«"' },
    '&::after': { content: '"»"' },
}))

const ProcessInformation = styled(ProcessInfo)(() => ({
    marginLeft: '0.5em',
}))

const TaskInformation = styled(TaskInfo)(() => ({
    marginLeft: '0.5em',
}))


// Reduce function returning the (recursive) sum of children and grand-children
// plus this namespace itself.
const countNamespaceWithChildren = (sum: number, ns: Namespace): number =>
    sum + ns.children.reduce(countNamespaceWithChildren, 1)


export interface NamespaceInfoProps {
    /** namespace with type, identifier and initial namespace indication. */
    namespace: Namespace,
    /** information about all processes (for some render support) */
    processes?: ProcessMap,
    /** suppress rendering leader process information.  */
    noprocess?: boolean,
    /** show short process information. */
    shortprocess?: boolean,
    /** is this a namespace shared with other leader processes? */
    shared?: boolean,
    /** optional CSS class name(s). */
    className?: string,
}

const namespacesWithChildren: (NamespaceType)[] = [NamespaceType.pid, NamespaceType.user]

/**
 * Component `Namespace` renders information about a particular namespace. The
 * type and ID get rendered, as well as the most senior process with its name,
 * or alternatively a bind-mounted or fd reference.
 *
 * Please note: this component never renders any child namespaces (even if the
 * given namespace is either a PID or user namespace).
 */
export const NamespaceInfo = ({
    namespace, processes, noprocess, shortprocess, shared, className
}: NamespaceInfoProps) => {
    if (!namespace) {
        return <></>
    }
    // If there is a leader process joined to this namespace, then prepare some
    // process information to be rendered alongside with the namespace type and
    // ID. Unless the process information is to be suppressed.
    const procinfo = namespace.ealdorman
        && (!noprocess || shortprocess)
        && <ProcessInformation
            process={namespace.ealdorman}
            short={shortprocess}
        />

    // Or is there one or more loose threads joined to this namespace? Then show
    // the oldest loose thread instead.
    const oldesttask = !procinfo && namespace.loosethreads?.slice().sort(compareBusybodies)[0]
    const taskinfo = !!oldesttask
        && (!noprocess || shortprocess)
        && <TaskInformation task={oldesttask} short={shortprocess} />

    // If there isn't any process attached to this namespace, prepare
    // information about bind mounts and fd references, if possible. This also
    // covers "hidden" (PID, user) namespaces which are somewhere in the
    // hierarchy without any other references to them anymore beyond the
    // parent-child references.
    const pathinfo = !noprocess && !procinfo && !taskinfo &&
        <PathInformation namespace={namespace} processes={processes} />

    // For user namespaces also prepare ownership information: the user name as
    // well as the UID of the Linux user "owning" the user namespace.
    const ownedbyroot = namespace['user-name'] === 'root' && 'root'
    const ownerinfo = namespace.type === NamespaceType.user &&
        'user-id' in namespace &&
        <OwnerInformation className={clsx(ownedbyroot)}>
            owned by {
                namespace['user-id'] ? <Person fontSize="inherit" /> : <Rude fontSize="inherit" />
            } UID {namespace['user-id']}
            {namespace['user-name'] && <>
                {' '}
                <OwnerName className={clsx(ownedbyroot)}>
                    {namespace['user-name']}
                </OwnerName>
            </>}
        </OwnerInformation>

    // For PID and user namespaces determine the total number of children and
    // grandchildren.
    const childrenCount = namespacesWithChildren.includes(namespace.type) && !shared &&
        namespace.children.length > 0 &&
        <UserChildrenInfo>
            [<ChildrenIcon fontSize="inherit" />&#8239;{countNamespaceWithChildren(-1, namespace)}]
        </UserChildrenInfo>

    return (
        <NamespaceInformation className={clsx(namespace.type, shared && namespaceShared, className)}>
            <NamespaceBadge namespace={namespace} shared={shared} />
            {childrenCount}
            {procinfo || taskinfo || pathinfo} {ownerinfo}
        </NamespaceInformation>
    )
}
