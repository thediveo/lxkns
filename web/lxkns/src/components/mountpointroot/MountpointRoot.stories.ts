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

import type { Meta, StoryObj } from "@storybook/react-vite";

import { MountpointRoot } from "./MountpointRoot";
import { type Namespace, type NamespaceMap, NamespaceType } from "models/lxkns";

const meta: Meta<typeof MountpointRoot> = {
  title: "Mount/MountpointRoot",
  component: MountpointRoot,
  argTypes: {
    namespaces: { control: false },
  },
  tags: ["autodocs"],
};

export default meta;

type Story = StoryObj<typeof MountpointRoot>;

const namespaces: NamespaceMap = {
  "12345678": {
    nsid: 12345678,
    type: NamespaceType.net,
  } as Namespace,
};

export const Basic: Story = {
  args: {
    root: "net:[12345678]",
    namespaces: namespaces,
  },
};
