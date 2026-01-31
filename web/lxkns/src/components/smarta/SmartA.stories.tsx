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

import type { Meta, StoryObj } from "@storybook/react-vite"

import { SmartA } from "./SmartA"

const meta: Meta<typeof SmartA> = {
    title: "Universal/SmartA",
    component: SmartA,
    tags: ["autodocs"],
}

export default meta

type Story = StoryObj<typeof SmartA>

export const External: Story = {
    args: {
        href: "https://github.com/thediveo/lxkns",
        children: "@thediveo/lxkns",
    },
}

export const Internal: Story = {
    args: {
        href: "/internal",
        children: "an app-internal route",
    },
}
