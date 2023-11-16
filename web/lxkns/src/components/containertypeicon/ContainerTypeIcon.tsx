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

import { SvgIconProps } from '@mui/material';


import ContainerIcon from 'icons/containers/Container'
import DockerIcon from 'icons/containers/Docker'
import ContainerdIcon from 'icons/containers/Containerd'

import IERuntimeIcon from 'icons/containers/IERuntime'
import IEAppIcon from 'icons/containers/IEApp'

import { Container } from 'models/lxkns'
import DockerManagedPluginIcon from 'icons/containers/DockerManagedPlugin'
import CRIIcon from 'icons/containers/CRI';

const containerTypeIcons: { [key: string]: (props: SvgIconProps) => JSX.Element } = {
    'docker.com': DockerIcon,
    'plugin.docker.com': DockerManagedPluginIcon,
    'containerd.io': ContainerdIcon,
    'com.siemens.industrialedge.runtime': IERuntimeIcon,
    'com.siemens.industrialedge.app': IEAppIcon,
    'k8s.io/cri-api': CRIIcon,
}

/**
 * The `ContainerTypeIconProps` component expects only a single property: the
 * container to render the corresponding icon.
 */
export interface ContainerTypeIconProps extends SvgIconProps {
    /** information about a discovered container. */
    container: Container
}

/**
 * Returns a container type (SVG) icon based on the type and flavor of the
 * specified container.
 *
 * @param container container object.
 */
export const ContainerTypeIcon = ({container, ...props}: ContainerTypeIconProps): JSX.Element => {
    // Now try to find a suitable container-flavor icon, or fall back to our
    // generic one.
    const Icon = containerTypeIcons[container.flavor]
    return (!!Icon && <Icon {...props} />) || <ContainerIcon {...props} />
}

export default ContainerTypeIcon
