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

import HomeIcon from '@mui/icons-material/Home'

import { DrawerLinkItem } from './DrawerLinkItem'

const meta: Meta<typeof DrawerLinkItem> = {
    title: 'Universal/DrawerLinkItem',
    component: DrawerLinkItem,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof DrawerLinkItem>

export const Basic: Story = {
    args: {
        label: 'Foo',
        path: '/foo',
    },
}

export const Icon: Story = {
    args: {
        label: 'Home',
        icon: <HomeIcon/>,
        path: '/home',
    },
}

export const Avatar: Story = {
    args: {
        label: 'Home',
        icon: <HomeIcon/>,
        path: '/home',
        avatar: true,
    },
}
