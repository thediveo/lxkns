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
import { CarCruiseControl, Database, Lan, Laptop, Memory, PhoneInTalk, Timer } from 'mdi-material-ui'

import { NamespaceType } from "models/lxkns"
import { SvgIconProps } from '@material-ui/core'


// Maps Linux-kernel namespace types to icons and tooltip information.
export interface NamespaceTypeInfo {
    tooltip: string
    icon: React.ComponentType<SvgIconProps> // https://stackoverflow.com/a/52559982
}

// Maps namespace types to icons and suitable tooltip texts.
export const namespaceTypeInfo: { [key in NamespaceType]: NamespaceTypeInfo } = {
    [NamespaceType.cgroup]: { tooltip: "control group", icon: CarCruiseControl },
    [NamespaceType.ipc]: { tooltip: "inter-process", icon: PhoneInTalk },
    [NamespaceType.mnt]: { tooltip: "mount", icon: Database },
    [NamespaceType.net]: { tooltip: "network", icon: Lan },
    [NamespaceType.pid]: { tooltip: "process identifier", icon: Memory },
    [NamespaceType.user]: { tooltip: "user", icon: Person },
    [NamespaceType.uts]: { tooltip: "*nix time sharing system", icon: Laptop },
    [NamespaceType.time]: { tooltip: "monotonous timers", icon: Timer },
}
