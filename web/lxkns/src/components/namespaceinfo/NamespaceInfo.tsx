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
import classNames from 'classnames'

import PersonIcon from '@material-ui/icons/Person'
import PhoneInTalkIcon from '@material-ui/icons/PhoneInTalk'
import TimerIcon from '@material-ui/icons/Timer'
import TextureIcon from '@material-ui/icons/Texture'
import MemoryIcon from '@material-ui/icons/Memory'
import SubdirectoryArrowRightIcon from '@material-ui/icons/SubdirectoryArrowRight'
import Tooltip from '@material-ui/core/Tooltip'

import Database from 'mdi-material-ui/Database'
import CarCruiseControl from 'mdi-material-ui/CarCruiseControl'
import Lan from 'mdi-material-ui/Lan'
import Laptop from 'mdi-material-ui/Laptop'
import FileLinkOutline from 'mdi-material-ui/FileLinkOutline'

import { ProcessInfo } from 'components/processinfo'
import { Namespace } from 'models/lxkns'

// Maps Linux-kernel namespace types to icons, including tooltips. 
const namespaceTypeIcons = {
    "cgroup": <Tooltip title="control group namespace"><CarCruiseControl fontSize="inherit" /></Tooltip>,
    "ipc": <Tooltip title="inter-process communication namespace"><PhoneInTalkIcon fontSize="inherit" /></Tooltip>,
    "mnt": <Tooltip title="mount namespace"><Database fontSize="inherit" /></Tooltip>,
    "net": <Tooltip title="network namespace"><Lan fontSize="inherit" /></Tooltip>,
    "pid": <Tooltip title="process identifier namespace"><MemoryIcon fontSize="inherit" /></Tooltip>,
    "user": <Tooltip title="user namespace"><PersonIcon fontSize="inherit" /></Tooltip>,
    "uts": <Tooltip title="*nix time sharing namespace"><Laptop fontSize="inherit" /></Tooltip>,
    "time": <Tooltip title="monotonous timers namespace"><TimerIcon fontSize="inherit" /></Tooltip>
}

export interface NamespaceInfoProps {
    namespace: Namespace,
    noprocess?: boolean,
}

// Component Namespace renders information about a particular namespace, passed
// in as a namespace object; type and ID get rendered, as well as the most
// senior process with its name, or a bind-mounted reference. This component
// never renders any child namespaces (of PID and user namespaces).
const NamespaceInfo = ({ namespace, noprocess }: NamespaceInfoProps) => {
    // If there is a leader process joined to this namespace, then prepare some
    // process information to be rendered alongside with the namespace type and
    // ID.
    const process =
        (namespace.ealdorman && <ProcessInfo process={noprocess ? null : namespace.ealdorman} />)
        || (namespace.reference &&
            <Tooltip title="bind mount"><span className="bindmount">
                <FileLinkOutline fontSize="inherit" />
                <span className="bindmount">"{namespace.reference}"</span>
            </span></Tooltip>) ||
        <Tooltip title={"intermediate hidden " + namespace.type + " namespace"}>
            <TextureIcon fontSize="inherit" />
        </Tooltip>

    const owner = namespace.type === 'user' &&
        <span className="owner">
            owned by UID {namespace['user-id']} {namespace['user-name'] && ('"' + namespace['user-name'] + '"')}
        </span>

    const children = namespace.type === 'user' &&
        <span className="userchildren">
            (<SubdirectoryArrowRightIcon fontSize="inherit" />
            {countNamespaceWithChildren(-1, namespace)})
        </span>

    return (
        <span className={classNames('namespace', namespace.type)}>
            <NamespacePill namespace={namespace} />
            {children}
            {process} {owner}
        </span>
    )
}

export default NamespaceInfo;

// reduce function returning the sum of children and grand-children plus this
// namespace itself.
const countNamespaceWithChildren = (acc: number, ns: Namespace) =>
    acc + ns.children.reduce(countNamespaceWithChildren, 1)


export interface NamespacePillProps { namespace: Namespace }

/**
 * Component `NamespacePill` renders a namespace "pill" consisting of the
 * namespace's type and identifier, in the typical "nstype:[nsid]" textual
 * notation, yet with some graphical adornments.
 */
export const NamespacePill = ({ namespace }: NamespacePillProps) => (
    <Tooltip title={`${namespace.type} namespace`}><>
        {namespaceTypeIcons[namespace.type]}&nbsp;
        <span className="pill">
            {namespace.type}:[{namespace.nsid}]
        </span>
    </></Tooltip>
)
