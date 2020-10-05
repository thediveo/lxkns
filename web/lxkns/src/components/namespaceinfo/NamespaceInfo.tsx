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
import Tooltip from '@material-ui/core/Tooltip'

import Ghost from 'mdi-material-ui/Ghost'
import Database from 'mdi-material-ui/Database'
import CarCruiseControl from 'mdi-material-ui/CarCruiseControl'
import Lan from 'mdi-material-ui/Lan'
import Laptop from 'mdi-material-ui/Laptop'
import FileLinkOutline from 'mdi-material-ui/FileLinkOutline'

import { ProcessInfo } from 'components/processinfo'
import { Namespace, NamespaceType } from 'models/lxkns'

import { makeStyles } from '@material-ui/core'

// Component styling...
const useStyles = makeStyles({
    namespace: {
    },
    namespacePath: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        fontStyle: 'italic',
        color: '#bf8d19',
        '& .MuiSvgIcon-root': {
            marginRight: '0.15em',
            verticalAlign: 'middle',
        },
    },
    namespacePill: {
        minWidth: '11.5em',
        verticalAlign: 'middle',

        display: 'inline-flex',
        justifyContent: 'space-between',
        alignItems: 'center',

        marginTop: '0.2ex',
        marginBottom: '0.2ex',
        marginRight: '0.5em',
        paddingLeft: '0.2em',
        paddingRight: '0.2em',
        paddingTop: '0.2ex',
        borderRadius: '0.2em',

        // ...and now for the namespace-type specific styling.
        '&$cgroup': {
            backgroundColor: '#fce1e1',
        },
        '&$ipc': {
            backgroundColor: '#f5ffcc',
        },
        '&$mnt': {
            backgroundColor: '#e4f2f5',
        },
        '&$net': {
            backgroundColor: '#e0ffe0',
        },
        '&$pid': {
            backgroundColor: '#daddf2',
        },
        '&$time': {
            backgroundColor: 'mediumaquamarine',
        },
        '&$user': {
            width: '9.5em',
            textAlign: 'center',
            backgroundColor: '#e9e8e8',
            fontWeight: 'bold',
        },
        '&$uts': {
            backgroundColor: '#fff2d9',
        }
    },
    userchildrenInfo: {
        display: 'inline-block',
        whiteSpace: 'nowrap',
        marginRight: '0.5em',
    },
    // The following is required so we can reference and thus combine
    // selectors for namespace type-specific styling of the "pill".
    cgroup: {},
    ipc: {},
    mnt: {},
    net: {},
    pid: {},
    user: {},
    uts: {},
    time: {}
})

// Maps Linux-kernel namespace types to icons, including tooltips.
interface NamespaceIcon {
    tooltip: string
    icon: JSX.Element
}

const namespaceTypeIcons: { [key: string]: NamespaceIcon } = {
    "cgroup": { tooltip: "control group", icon: <CarCruiseControl fontSize="inherit" /> },
    "ipc": { tooltip: "inter-process", icon: <PhoneInTalkIcon fontSize="inherit" /> },
    "mnt": { tooltip: "mount", icon: <Database fontSize="inherit" /> },
    "net": { tooltip: "network", icon: <Lan fontSize="inherit" /> },
    "pid": { tooltip: "process identifier", icon: <MemoryIcon fontSize="inherit" /> },
    "user": { tooltip: "user", icon: <PersonIcon fontSize="inherit" /> },
    "uts": { tooltip: "*nix time sharing system", icon: <Laptop fontSize="inherit" /> },
    "time": { tooltip: "monotonous timers", icon: <TimerIcon fontSize="inherit" /> },
}

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
const NamespaceInfo = ({ namespace, noprocess }: NamespaceInfoProps) => {
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
        <NamespacePath namespace={namespace} />

    // For user namespaces also prepare ownership information.
    const ownerinfo = namespace.type === NamespaceType.user &&
        <span className="owner">
            owned by UID {namespace['user-id']} {namespace['user-name'] && ('"' + namespace['user-name'] + '"')}
        </span>

    const children = namespace.type === NamespaceType.user &&
        <span className={classes.userchildrenInfo}>
            [<SubdirectoryArrowRightIcon fontSize="inherit" />
            {countNamespaceWithChildren(-1, namespace)}]
        </span>

    return (
        <span className={`${classes.namespace} ${namespace.type}`}>
            <NamespacePill namespace={namespace} />
            {children}
            {procinfo || pathinfo} {ownerinfo}
        </span>
    )
}

export default NamespaceInfo;

// reduce function returning the sum of children and grand-children plus this
// namespace itself.
const countNamespaceWithChildren = (acc: number, ns: Namespace) =>
    acc + ns.children.reduce(countNamespaceWithChildren, 1)


export interface NamespaceProps {
    /** namespace with type and identifier. */
    namespace: Namespace
}

/**
 * Component `NamespacePill` renders a namespace "pill" consisting of just the
 * namespace's type and identifier, in the typical "nstype:[nsid]" textual
 * notation. Yet it gets some simple graphical adornments; in particular, an
 * icon matching the type of namespace.
 */
export const NamespacePill = ({ namespace }: NamespaceProps) => {
    const classes = useStyles()

    return (
        // Ouch ... Tooltip won't display its tooltip on a <> child, but
        // instead we have to use a <span> to make it work as expected...
        <Tooltip title={`${namespaceTypeIcons[namespace.type].tooltip} namespace`}>
            <span className={`${classes.namespacePill} ${classes[namespace.type]}`}>
                {namespaceTypeIcons[namespace.type].icon}
                {namespace.type}:[{namespace.nsid}]
            </span>
        </Tooltip>
    )
}

/**
 * 
 */
const NamespacePath = ({ namespace }: NamespaceProps) => {
    const classes = useStyles()

    const procfdPath = namespace.reference &&
        namespace.reference.startsWith('/proc/') &&
        namespace.reference.includes('/fd/')

    return (
        (!namespace.reference &&
            <Tooltip title={"intermediate hidden " + namespace.type + " namespace"}>
                <span className={classes.namespacePath}>
                    <Ghost fontSize="inherit" />
                </span>
            </Tooltip>
        ) || (procfdPath && 
            <Tooltip title="kept alive by file descriptor">
                <span className={classes.namespacePath}>
                    <FileLinkOutline fontSize="inherit" />
                    <span className="bindmount">"{namespace.reference}"</span>
                </span>
            </Tooltip>
        ) || (
            <Tooltip title="bind mount">
                <span className={classes.namespacePath}>
                    <FileLinkOutline fontSize="inherit" />
                    <span className="bindmount">"{namespace.reference}"</span>
                </span>
            </Tooltip>
        )
    )
}
