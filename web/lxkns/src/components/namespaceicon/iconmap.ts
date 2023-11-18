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

import { Person, Timer } from '@mui/icons-material'

import CgroupNamespace from 'icons/namespaces/Cgroup'
import MountNamespace from 'icons/namespaces/Mount'
import NetworkNamespace from 'icons/namespaces/Network'
import IPCNamespace from 'icons/namespaces/IPC'
import PIDNamespace from 'icons/namespaces/PID'
import UTSNamespace from 'icons/namespaces/UTS'

import { NamespaceType } from 'models/lxkns'
import { SvgIconProps } from '@mui/material'

type SvgIconer = (props: SvgIconProps) => JSX.Element

// Maps Linux-kernel namespace types to icons and tooltip information.
export interface NamespaceTypeInfo {
    tooltip: string
    icon: SvgIconer
}

// Maps namespace types to icons and suitable tooltip texts.
export const namespaceTypeInfo: { [key in NamespaceType]: NamespaceTypeInfo } = {
    [NamespaceType.cgroup]: { tooltip: "control group", icon: CgroupNamespace },
    [NamespaceType.ipc]: { tooltip: "inter-process", icon: IPCNamespace },
    [NamespaceType.mnt]: { tooltip: "mount", icon: MountNamespace },
    [NamespaceType.net]: { tooltip: "network", icon: NetworkNamespace },
    [NamespaceType.pid]: { tooltip: "process identifier", icon: PIDNamespace },
    [NamespaceType.user]: { tooltip: "user", icon: Person as SvgIconer },
    [NamespaceType.uts]: { tooltip: "*nix time sharing system", icon: UTSNamespace },
    [NamespaceType.time]: { tooltip: "monotonous timers", icon: Timer as SvgIconer },
}
