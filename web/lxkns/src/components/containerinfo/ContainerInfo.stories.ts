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

import { ContainerInfo } from './ContainerInfo'
import { Container, Engine, Group, Process } from 'models/lxkns'

const meta: Meta<typeof ContainerInfo> = {
    title: 'Container/ContainerInfo',
    component: ContainerInfo,
    argTypes: {
        container: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ContainerInfo>

const process: Process = {
    pid: 41,
    ppid: 1,
    name: 'foobar-process',
} as Process

const container: Container = {
    id: 'deadbeafco1dcafe',
    name: 'mouldy_moby',
    type: 'docker.com',
    flavor: 'docker.com',
    pid: 41,
    paused: false,
    labels: {},
    groups: [],
    engine: {} as Engine,
    process: process,
}

export const Running: Story = {
    args: {
        container: container,
    },
}

export const Paused: Story = {
    args: {
        container: {
            ...container,
            paused: true,
        },
    },
}

export const Grouped: Story = {
    args: {
        container: {
            ...container,
            groups: [{
                name: 'bleary_beathoven',
                type: 'com.docker.compose.project',
                flavor: 'com.docker.compose.project',
            } as Group],
        },
    },
}

export const ContainerInContainer: Story = {
    args: {
        container: {
            ...container,
            labels: {
                'turtlefinder/container/prefix': 'outer-container',
            }
        },
    },
}
