// Copyright 2020 Harald Albrecht.
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

import React from 'react'

import { HelpViewer, HelpViewerChapter } from 'components/helpviewer'
import { MuiMarkdown } from 'components/muimarkdown'
import { NamespaceBadge } from 'components/namespacebadge'
import { SmartA } from 'components/smarta'
import { Namespace } from 'models/lxkns'
import { Box } from '@material-ui/core'
import { Card } from '@material-ui/core'


/**
 * Convenience wrapper for lazily importing a help chapter MDX module.
 * 
 * @param name name (without .mdx extension and without any path) of a chapter
 * .mdx file; chapter files are located in the chapters/ subdirectory.
 */
const ch = (name: string) => React.lazy(() => import(`!babel-loader!mdx-loader!./chapters/${name}.mdx`))

const chapters: HelpViewerChapter[] = [
    { title: 'lxkns', chapter: ch('Lxkns'), slug: 'lxkns' },
    { title: 'Refresh', chapter: ch('Refresh'), slug: 'refresh' },
    { title: 'Views', chapter: ch('Views'), slug: 'views' },
    { title: 'Namespaces', chapter: ch('Namespaces'), slug: 'namespaces' },
    { title: 'Settings', chapter: ch('Settings'), slug: 'settings' },
]

const Example = ({ children }: { children: React.ReactNode }) => (
    <Box m={2}>
        <Card>
            <Box m={1}>
                {children}
            </Box>
        </Card>
    </Box>
)

const NamespaceExample = ({ type, initial }) =>
    <NamespaceBadge namespace={{
        nsid: 4026531837,
        type: type,
        ealdorman: {},
        initial: initial,
        parent: null,
        children: [],
    } as Namespace} />

export const Help = () => (
    <HelpViewer
        chapters={chapters}
        baseroute="/help"
        markdowner={MuiMarkdown}
        shortcodes={{ a: SmartA, Example, NamespaceBadge, NamespaceExample }}
    />
)
