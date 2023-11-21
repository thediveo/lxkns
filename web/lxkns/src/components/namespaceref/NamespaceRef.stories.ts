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

import { NamespaceRef } from './NamespaceRef'
import { Namespace, NamespaceType } from 'models/lxkns'

const meta: Meta<typeof NamespaceRef> = {
    title: 'Namespace/NamespaceRef',
    component: NamespaceRef,
    argTypes: {
        namespace: { control: false },
        processes: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NamespaceRef>

const namespace: Namespace = {
    nsid: 12345678,
    type: NamespaceType.net,
} as Namespace

export const ProcfsReference: Story = {
    args: {
        namespace: {
            ...namespace,
            reference: ['/proc/123/ns/net'],
        } as Namespace,
    },
}

export const ProcfsFdReference: Story = {
    args: {
        namespace: {
            ...namespace,
            reference: ['/proc/123/fd/42'],
        } as Namespace,
    },
}

export const BindMountedReference: Story = {
    args: {
        namespace: {
            ...namespace,
            reference: ['/var/run/snaps/chaotium-latest'],
        } as Namespace,
    },
}

export const IntermediateHidden: Story = {
    args: {
        namespace: {
            ...namespace,
            type: NamespaceType.user,
        } as Namespace,
    },
}
