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

import React from 'react';
import classNames from 'classnames';
import DirectionsRunIcon from '@material-ui/icons/DirectionsRun';
import LinkIcon from '@material-ui/icons/Link';
import PersonIcon from '@material-ui/icons/Person';
import StorageIcon from '@material-ui/icons/Storage';
import SettingsEthernetIcon from '@material-ui/icons/SettingsEthernet';
import PhoneInTalkIcon from '@material-ui/icons/PhoneInTalk';
import SpeedIcon from '@material-ui/icons/Speed';
import TimerIcon from '@material-ui/icons/Timer';
import DnsIcon from '@material-ui/icons/Dns';
import TextureIcon from '@material-ui/icons/Texture';
import MemoryIcon from '@material-ui/icons/Memory';
import SubdirectoryArrowRightIcon from '@material-ui/icons/SubdirectoryArrowRight';
import Tooltip from '@material-ui/core/Tooltip';

// Maps Linux-kernel namespace types to icons, including tooltips. 
const namespaceTypeIcons = {
    "cgroup": <Tooltip title="control group namespace"><SpeedIcon fontSize="inherit" /></Tooltip>,
    "ipc": <Tooltip title="inter-process communication namespace"><PhoneInTalkIcon fontSize="inherit" /></Tooltip>,
    "mnt": <Tooltip title="mount namespace"><StorageIcon fontSize="inherit" /></Tooltip>,
    "net": <Tooltip title="network namespace"><SettingsEthernetIcon fontSize="inherit" /></Tooltip>,
    "pid": <Tooltip title="process identifier namespace"><MemoryIcon fontSize="inherit" /></Tooltip>,
    "user": <Tooltip title="user namespace"><PersonIcon fontSize="inherit" /></Tooltip>,
    "uts": <Tooltip title="*nix time sharing namespace"><DnsIcon fontSize="inherit" /></Tooltip>,
    "time": <Tooltip title="monotonous timers namespace"><TimerIcon fontSize="inherit" /></Tooltip>
};

// Component Namespace renders information about a particular namespace, passed
// in as a namespace object; type and ID get rendered, as well as the most
// senior process with its name, or a bind-mounted reference. This component
// never renders any child namespaces (of PID and user namespaces).
const Namespace = ({namespace}) => {
    // Prepare information about the control group of the leader process (if
    // there is any joined to this namespace), which is useful in identifying
    // processes with generic names. 
    const cgroup = namespace.ealdorman && namespace.ealdorman.cgroup &&
        <span className="cgroupinfo">
            <SpeedIcon fontSize="inherit"/>&nbsp;
            <span>"<span className="cgroupname">{namespace.ealdorman.cgroup}</span>"</span>
        </span>;
    // If there is a leader process joined to this namespace, then prepare some
    // process information to be rendered alongside with the namespace type and
    // ID.
    const process = (namespace.ealdorman &&
        <Tooltip title="process"><span className="processinfo">
            <DirectionsRunIcon fontSize="inherit" />
            <span className="processname">"{namespace.ealdorman.name}"</span> ({namespace.ealdorman.pid})
            {cgroup}
        </span></Tooltip>) || (namespace.reference &&
            <Tooltip title="bind mount"><span className="bindmount">
                <LinkIcon fontSize="inherit" />
                <span className="bindmount">"{namespace.reference}"</span>
            </span></Tooltip>) ||
        <Tooltip title={"intermediate hidden " + namespace.type + " namespace"}>
            <TextureIcon fontSize="inherit" />
        </Tooltip>;

    const owner = namespace.type === 'user' &&
        <span className="owner">
            owned by UID {namespace['user-id']} {namespace['user-name'] && ('"' + namespace['user-name'] + '"')}
        </span>;

    const children = namespace.type === 'user' &&
        <span className="userchildren">
            (<SubdirectoryArrowRightIcon fontSize="inherit" />
            {countNamespaceWithChildren(-1, namespace)})
        </span>;

    return <span className={classNames('namespace', namespace.type)}>
        {namespaceTypeIcons[namespace.type]}&nbsp;
        <span className="pill">{namespace.type}:[{namespace.nsid}]</span>
        {children}
        {process} {owner}
    </span>;
};

export default Namespace;

// reduce function returning the sum of children and grand-children plus this
// namespace itself.
const countNamespaceWithChildren = (acc, ns, idx, arr) =>
    acc + ns.children.reduce(countNamespaceWithChildren, 1);