// Copyright 2023 Harald Albrecht.
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

import type { Meta, StoryObj } from '@storybook/react'

import { ContainerTypeIcon } from './ContainerTypeIcon'
import { Container } from 'models/lxkns'

const unknownCntr = {flavor: 'unknown'} as Container

const meta: Meta<typeof ContainerTypeIcon> = {
    title: 'Container/ContainerTypeIcon',
    component: ContainerTypeIcon,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ContainerTypeIcon>

export const UnknownContainerTypeAndFlavor: Story = {
    args: {
        container: unknownCntr,
    },
}

export const DockerContainer: Story = {
    args: {
        container: {
            ...unknownCntr,
            flavor: 'docker.com'
        },
    },
}

export const DockerPlugin: Story = {
    args: {
        container: {
            ...unknownCntr,
            flavor: 'plugin.docker.com'
        },
    },
}

export const ContainerdContainer: Story = {
    args: {
        container: {
            ...unknownCntr,
            flavor: 'containerd.io'
        },
    },
}

export const IndustrialEdgeRuntimeContainer: Story = {
    args: {
        container: {
            ...unknownCntr,
            flavor: 'com.siemens.industrialedge.runtime'
        },
    },
}

export const IndustrialEdgeAppContainer: Story = {
    args: {
        container: {
            ...unknownCntr,
            flavor: 'com.siemens.industrialedge.app'
        },
    },
}

export const CRIContainer: Story = {
    name: 'CRI',
    args: {
        container: {
            ...unknownCntr,
            flavor: 'k8s.io/cri-api'
        },
    },
}
