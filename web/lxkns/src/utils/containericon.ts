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

import ContainerIcon from 'icons/containers/Container'
import DockerIcon from 'icons/containers/Docker'
import ContainerdIcon from 'icons/containers/Containerd'

import IERuntimeIcon from 'icons/containers/IERuntime'
import IEAppIcon from 'icons/containers/IEApp'

//import PodIcon from 'icons/containers/Pod'
//import K8sPodIcon from 'icons/containers/K8sPod'
import { Container } from 'models/lxkns'
//import ComposerProjectIcon from 'icons/containers/ComposerProject'

const ContainerTypeIcons = {
    'unknowntype': ContainerIcon,
    'docker.com': DockerIcon,
    'containerd.io': ContainerdIcon,
    'com.siemens.industrialedge.runtime': IERuntimeIcon,
    'com.siemens.industrialedge.app': IEAppIcon,
}

/*
const PodTypeIcons = {
    'io.kubernetes.pod': K8sPodIcon,
}
*/

/**
 * Returns a container type icon (constructor) based on the type and flavor of
 * the specified container.
 *
 * @param container container object.
 */
 export const ContainerTypeIcon = (container: Container) => {
    // Now try to find a suitable container-flavor icon, or fall back to our
    // generic one.
    return ContainerTypeIcons[container.flavor] || ContainerIcon
}
