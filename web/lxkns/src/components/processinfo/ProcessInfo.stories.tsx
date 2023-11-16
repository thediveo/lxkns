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

import { ProcessInfo } from './ProcessInfo'
import { Container, Engine, Group, Process } from 'models/lxkns'

const meta: Meta<typeof ProcessInfo> = {
    title: 'Process/ProcessInfo',
    component: ProcessInfo,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ProcessInfo>

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
    process: {} as Process,
}

const process: Process = {
    pid: 41,
    ppid: 1,
    name: 'foobar-process',
    starttime: 123,
    cpucgroup: "/fridge",
    fridgecgroup: "/fridge",
    fridgefrozen: false,
    container: container,
} as Process


export const Basic: Story = {
    args: {
        process: process,
    },
}

export const Short: Story = {
    args: {
        short: true,
        process: {
            ...process,
            container: null,
        },
    },
}

export const Frozen: Story = {
    args: {
        process: {
            ...process,
            container: null,
            fridgefrozen: true,
        },
    },
}

const composeProject: Group = {
    name: 'captn-ahab',
    type: 'com.docker.compose.project',
    flavor: 'com.docker.compose.project',
    containers: [],
    labels: {},
}

export const DockerComposeProject: Story = {
    args: {
        process: {
            ...process,
            container: {
                ...container,
                groups: [composeProject],
            },
        },
    },
}

export const IndustrialEdgeApp: Story = {
    args: {
        process: {
            ...process,
            container: {
                ...container,
                flavor: 'com.siemens.industrialedge.app',
                groups: [composeProject],
            },
        },
    },
}

export const KubernetesPod: Story = {
    args: {
        process: {
            ...process,
            container: {
                ...container,
                groups: [{
                    name: 'captn-ahab',
                    type: 'io.kubernetes.pod',
                    flavor: 'io.kubernetes.pod',
                } as Group],
            },
        },
    },
}
