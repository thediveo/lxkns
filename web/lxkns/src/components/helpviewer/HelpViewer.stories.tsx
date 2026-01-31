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
import { MuiMarkdown, type MuiMarkdownProps } from "components/muimarkdown"

import { HelpViewer } from "./HelpViewer"

import chintro from "./chapters/01-intro.mdx"
import chfoobar from "./chapters/02-foobar.mdx"
import chnew from "./chapters/03-newchapter.mdx"

const MyMarkdowner = (props: MuiMarkdownProps) => <MuiMarkdown {...props} />

const chapters = [
    { title: "Intro", chapter: chintro },
    { title: "Foo Bar", chapter: chfoobar },
    { title: "A New Chapter", chapter: chnew },
]

const meta: Meta<typeof HelpViewer> = {
    title: "Universal/HelpViewer",
    component: HelpViewer,
    tags: ["autodocs"],
}

export default meta

type Story = StoryObj<typeof HelpViewer>

export const Standard: Story = {
    parameters: {
        routerProps: {
            initialEntries: ["/help"]
        },
    },
    render: () => (
        <HelpViewer
            chapters={chapters}
            baseroute="/help"
            style={{ height: "50ex", maxHeight: "50ex" }}
            markdowner={MyMarkdowner}
        />
    ),
}
