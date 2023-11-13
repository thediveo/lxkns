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

import { TaskInfo } from './TaskInfo'
import { Container, Engine, NamespaceSet, Process, Task } from 'models/lxkns'

const meta: Meta<typeof TaskInfo> = {
    title: 'Universal/TaskInfo',
    component: TaskInfo,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof TaskInfo>

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
} as Process

const task: Task = {
    tid: 42,
    name: 'foobartask',
    process: process,
    starttime: 123,
    cpucgroup: "/fridge",
    fridgecgroup: "/fridge",
    fridgefrozen: true,
    namespaces: {} as NamespaceSet,
}

export const Basic: Story = {
    args: {
        task: task,
    },
}

export const Short: Story = {
    args: {
        task: task,
        short: true,
    },
}

export const InContainer: Story = {
    args: {
        task: {
            ...task,
            process: {
                ...process,
                container: container,
            },
        },
    },
}
