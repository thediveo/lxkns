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

import PersonIcon from '@material-ui/icons/Person'
import PhoneInTalkIcon from '@material-ui/icons/PhoneInTalk'
import TimerIcon from '@material-ui/icons/Timer'
import MemoryIcon from '@material-ui/icons/Memory'
import SubdirectoryArrowRightIcon from '@material-ui/icons/SubdirectoryArrowRight'

import { ProcessInfo } from 'components/processinfo'
import { Namespace, NamespaceType } from 'models/lxkns'

import { makeStyles } from '@material-ui/core'
import { NamespaceRef } from 'components/namespaceref'
import { NamespaceBadge } from 'components/namespacebadge'


// Component styling...
const useStyles = makeStyles({
    namespace: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        verticalAlign: 'middle',
    },
    userchildrenInfo: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        marginRight: '0.5em',
    },
})

// reduce function returning the sum of children and grand-children plus this
// namespace itself.
const countNamespaceWithChildren = (sum: number, ns: Namespace) =>
    sum + ns.children.reduce(countNamespaceWithChildren, 1)

export interface NamespaceInfoProps {
    namespace: Namespace,
    noprocess?: boolean,
}

// Component `Namespace` renders information about a particular namespace. The
// type and ID get rendered, as well as the most senior process with its name,
// or alternatively a bind-mounted or fd reference.
//
// Please note: this component never renders any child namespaces (even if it
// is a PID and user namespace).
export const NamespaceInfo = ({ namespace, noprocess }: NamespaceInfoProps) => {

    const classes = useStyles()

    // If there is a leader process joined to this namespace, then prepare some
    // process information to be rendered alongside with the namespace type and
    // ID. Unless the process information is to be suppressed.
    const procinfo = !noprocess && namespace.ealdorman &&
        <ProcessInfo process={namespace.ealdorman} />

    // If there isn't any process attached to this namespace, prepare
    // information about bind mounts and fd references, if possible. This also
    // covers "hidden" (PID, user) namespaces which are somewhere in the
    // hierarchy without any other references to them anymore beyond the
    // parent-child references.
    const pathinfo = !namespace.ealdorman &&
        <NamespaceRef namespace={namespace} />

    // For user namespaces also prepare ownership information.
    const ownerinfo = namespace.type === NamespaceType.user &&
        'user-id' in namespace &&
        <span className="owner">
            owned by UID {namespace['user-id']} {namespace['user-name'] && ('"' + namespace['user-name'] + '"')}
        </span>

    const children = namespace.type === NamespaceType.user &&
        namespace.children.length > 0 &&
        <span className={classes.userchildrenInfo}>
            [<SubdirectoryArrowRightIcon fontSize="inherit" />
            {countNamespaceWithChildren(-1, namespace)}]
        </span>

    return (
        <span className={`${classes.namespace} ${namespace.type}`}>
            <NamespaceBadge namespace={namespace} />
            {children}
            {procinfo || pathinfo} {ownerinfo}
        </span>
    )
}
