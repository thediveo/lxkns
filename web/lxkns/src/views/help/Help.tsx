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
import { Box, makeStyles } from '@material-ui/core'
import { Card } from '@material-ui/core'
import { Provider } from 'jotai'
import { expandInitiallyAtom, showSharedNamespacesAtom, showSystemProcessesAtom } from 'views/settings'
import { AnyAtom } from 'jotai/core/types'


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
    { title: 'All View', chapter: ch('Allview'), slug: 'allview' },
    { title: 'Type-Specific Views', chapter: ch('Typedviews'), slug: 'typedviews' },
    { title: 'Mount Namespaces', chapter: ch('Mounts'), slug: 'mounts' },
    { title: 'Application Bar', chapter: ch('Appbar'), slug: 'appbar' },
    { title: 'Namespaces', chapter: ch('Namespaces'), slug: 'namespaces' },
    { title: 'Settings', chapter: ch('Settings'), slug: 'settings' },
]

interface ExampleProps {
    /** optional CSS maximum width for the example card element. */
    maxWidth?: string
    /** optional jōtai states. */
    states?: [AnyAtom, unknown][]
    /** the example rendering... */
    children: React.ReactNode
}

const initials: [AnyAtom, unknown][] = [
    [showSystemProcessesAtom, false],
    [showSharedNamespacesAtom, true],
    [expandInitiallyAtom, true],
]

// Ensure consist example rendering by providing a dedicated, erm, state
// provider which we prime to a known state, independent of any user
// preferences.
const Example = ({ children, maxWidth, states }: ExampleProps) => (
    // Luckily, the jōtai provider initializes in order, so later settings
    // override earlier ones, if necessary.
    <Provider initialValues={[...initials, ...(states ? states : [])]}>
        <Box m={2}>
            <Card style={{ maxWidth: maxWidth || '100%' }}>
                <Box m={1} style={{ overflowX: 'auto' }}>
                    {children}
                </Box>
            </Card>
        </Box>
    </Provider>
)

const useStyles = makeStyles((theme) => ({
    iconbox: {
        display: 'inline-block',
        verticalAlign: 'middle',
        fontSize: '70%', // why, CSS, oh why???
        border: `1px solid ${theme.palette.text.disabled}`,
        padding: 1,
        borderRadius: theme.spacing(1) / 2,

        '& > .MuiSvgIcon-root': {
            verticalAlign: 'middle',
            fontSize: 'calc(100% + 2px)',
        },
    }
}))

const BoxedIcons = ({ children }: { children: React.ReactNode }) => {
    const classes = useStyles()

    return <span className={classes.iconbox}>{children}</span>
}

const NamespaceExample = ({ type, initial, shared }) =>
    <NamespaceBadge namespace={{
        nsid: 4026531837,
        type: type,
        ealdorman: {},
        initial: initial,
        parent: null,
        children: [],
    } as Namespace} shared={shared} />

export const Help = () => (
    <HelpViewer
        chapters={chapters}
        baseroute="/help"
        markdowner={MuiMarkdown}
        shortcodes={{ a: SmartA, BoxedIcons, Example, NamespaceBadge, NamespaceExample }}
    />
)
