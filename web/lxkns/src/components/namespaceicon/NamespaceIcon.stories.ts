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

import { NamespaceIcon } from "./NamespaceIcon";
import { NamespaceType } from "models/lxkns";

const meta: Meta<typeof NamespaceIcon> = {
  title: "Namespace/NamespaceIcon",
  component: NamespaceIcon,
  tags: ["autodocs"],
};

export default meta;

type Story = StoryObj<typeof NamespaceIcon>;

export const Cgroup: Story = {
  name: "cgroup",
  args: {
    type: NamespaceType.cgroup,
  },
};

export const IPC: Story = {
  name: "ipc",
  args: {
    type: NamespaceType.ipc,
  },
};

export const MNT: Story = {
  name: "mnt",
  args: {
    type: NamespaceType.mnt,
  },
};

export const NET: Story = {
  name: "net",
  args: {
    type: NamespaceType.net,
  },
};

export const PID: Story = {
  name: "pid",
  args: {
    type: NamespaceType.pid,
  },
};

export const User: Story = {
  name: "user",
  args: {
    type: NamespaceType.user,
  },
};

export const UTS: Story = {
  name: "uts",
  args: {
    type: NamespaceType.uts,
  },
};

export const Time: Story = {
  name: "time",
  args: {
    type: NamespaceType.time,
  },
};

export const Invalid: Story = {
  args: {
    type: "foobar" as NamespaceType,
  },
};
