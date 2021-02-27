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

import { Person, Timer } from '@material-ui/icons'
import { Lan, PhoneInTalk } from 'mdi-material-ui'

import CgroupNamespace from 'icons/namespaces/Cgroup'
import MountNamespace from 'icons/namespaces/Mount'
import PIDNamespace from 'icons/namespaces/PID'
import UTSNamespace from 'icons/namespaces/UTS'

import { NamespaceType } from 'models/lxkns'
import { SvgIconProps } from '@material-ui/core'


// Maps Linux-kernel namespace types to icons and tooltip information.
export interface NamespaceTypeInfo {
    tooltip: string
    icon: React.ComponentType<SvgIconProps> // https://stackoverflow.com/a/52559982
}

// Maps namespace types to icons and suitable tooltip texts.
export const namespaceTypeInfo: { [key in NamespaceType]: NamespaceTypeInfo } = {
    [NamespaceType.cgroup]: { tooltip: "control group", icon: CgroupNamespace },
    [NamespaceType.ipc]: { tooltip: "inter-process", icon: PhoneInTalk },
    [NamespaceType.mnt]: { tooltip: "mount", icon: MountNamespace },
    [NamespaceType.net]: { tooltip: "network", icon: Lan },
    [NamespaceType.pid]: { tooltip: "process identifier", icon: PIDNamespace },
    [NamespaceType.user]: { tooltip: "user", icon: Person },
    [NamespaceType.uts]: { tooltip: "*nix time sharing system", icon: UTSNamespace },
    [NamespaceType.time]: { tooltip: "monotonous timers", icon: Timer },
}
