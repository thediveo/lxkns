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

import { ProcessInfo } from "./ProcessInfo";
import type { Container, Engine, Group, Process } from "models/lxkns";

const meta: Meta<typeof ProcessInfo> = {
    title: "Process/ProcessInfo",
    component: ProcessInfo,
    argTypes: {
        process: { control: false },
    },
    tags: ["autodocs"],
};

export default meta;

type Story = StoryObj<typeof ProcessInfo>;

const container: Container = {
    id: "deadbeafco1dcafe",
    name: "mouldy_moby",
    type: "docker.com",
    flavor: "docker.com",
    pid: 41,
    paused: false,
    labels: {},
    groups: [],
    engine: {} as Engine,
    process: {} as Process,
};

const process: Process = {
    pid: 41,
    ppid: 1,
    name: "foobar-process",
    starttime: 123,
    cpucgroup: "/fridge",
    fridgecgroup: "/fridge",
    fridgefrozen: false,
    container: container,
} as Process;

export const Default: Story = {
    args: {
        process: process,
    },
    parameters: {
        docs: {
            description: {
                story: 'Notice that when the process is running no running state indication is rendered in order to reduce clutter.',
            },
        },
    },
};

export const PID1_is_King: Story = {
    args: {
        process: {
            ...process,
            pid: 1,
            ppid: 0,
        },
    },
    parameters: {
        docs: {
            description: {
                story: 'Notice how PID1 is adorned with a crown.',
            },
        },
    },
};

export const Short: Story = {
    args: {
        short: true,
        process: {
            ...process,
            container: null,
        },
    },
    parameters: {
        docs: {
            description: {
                story: 'Just the process name and PID, no container or any other details.',
            },
        },
    },
};

export const Frozen: Story = {
    args: {
        process: {
            ...process,
            container: null,
            fridgefrozen: true,
        },
    },
    parameters: {
        docs: {
            description: {
                story: 'Notice that this time the cgroup details appear.',
            },
        },
    },
};

const composeProject: Group = {
    name: "captn-ahab",
    type: "com.docker.compose.project",
    flavor: "com.docker.compose.project",
    containers: [],
    labels: {},
};

export const ComposeProject: Story = {
    args: {
        process: {
            ...process,
            container: {
                ...container,
                groups: [composeProject],
            },
        },
    },
    parameters: {
        docs: {
            description: {
                story: 'The compose project name is rendered, together with the composer kraken icon.',
            },
        },
    },
};

export const SiemensIndustrialEdgeApp: Story = {
    args: {
        process: {
            ...process,
            container: {
                ...container,
                flavor: "com.siemens.industrialedge.app",
                groups: [composeProject],
            },
        },
    },
    parameters: {
        docs: {
            description: {
                story: 'The app compose project and container icons are both shown as Industrial Edge icons.',
            },
        },
    },
};

export const KubernetesPod: Story = {
    args: {
        process: {
            ...process,
            container: {
                ...container,
                type: 'k8s.io/cri-api',
                flavor: 'k8s.io/cri-api',
                groups: [
                    {
                        name: "captn-ahab",
                        type: "io.kubernetes.pod",
                        flavor: "io.kubernetes.pod",
                    } as Group,
                ],
            },
        },
    },
    parameters: {
        docs: {
            description: {
                story: 'The container has a tiller icon and the k8s namespace has a k8s pod/namespace icon.',
            },
        },
    },
};
