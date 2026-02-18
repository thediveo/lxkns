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

import React, { type ReactNode } from 'react'

import { HelpViewer, type HelpViewerChapter } from 'components/helpviewer'
import { MuiMarkdown } from 'components/muimarkdown'
import { NamespaceBadge } from 'components/namespacebadge'
import { SmartA } from 'components/smarta'
import { type Namespace, NamespaceType } from 'models/lxkns'
import { Box, styled } from '@mui/material'
import { Card } from '@mui/material'
import { type Atom, Provider, type WritableAtom } from 'jotai'
import { useHydrateAtoms } from 'jotai/utils'
import { expandInitiallyAtom, showSharedNamespacesAtom, showSystemProcessesAtom } from 'views/settings'


/**
 * Convenience wrapper for lazily importing a help chapter MDX module.
 * 
 * @param name name (without .mdx extension and without any path) of a chapter
 * .mdx file; chapter files are located in the chapters/ subdirectory.
 */
const ch = (name: string) => React.lazy(() => import(`./chapters/${name}.mdx`))

const chapters: HelpViewerChapter[] = [
    { title: 'lxkns', chapter: ch('Lxkns'), slug: 'lxkns' },
    { title: 'Refresh', chapter: ch('Refresh'), slug: 'refresh' },
    { title: '“All” View', chapter: ch('Allview'), slug: 'allview' },
    { title: 'Type-Specific Views', chapter: ch('Typedviews'), slug: 'typedviews' },
    { title: 'Mount Namespaces', chapter: ch('Mounts'), slug: 'mounts' },
    { title: 'Containers', chapter: ch('Containers'), slug: 'containers' },
    { title: 'Core Fancy', chapter: ch('Affinities'), slug: 'affinities'},
    { title: 'Application Bar', chapter: ch('Appbar'), slug: 'appbar' },
    { title: 'Namespaces', chapter: ch('Namespaces'), slug: 'namespaces' },
    { title: 'Settings', chapter: ch('Settings'), slug: 'settings' },
]

interface ExampleProps {
    /** optional CSS maximum width for the example card element. */
    maxWidth?: string
    /** optional jōtai states. */
    states?: [Atom<unknown>, unknown][]
    /** the example rendering... */
    children: React.ReactNode
}

const initials: [Atom<unknown>, unknown][] = [
    [showSystemProcessesAtom, false],
    [showSharedNamespacesAtom, true],
    [expandInitiallyAtom, true],
]

interface AtomsHydratorProps {
    atomvals: Iterable<readonly [WritableAtom<unknown, [unknown], unknown>, unknown]>
    children: ReactNode
}

const AtomsHydrator = ({ atomvals, children}: AtomsHydratorProps) => {
    useHydrateAtoms(new Map(atomvals))
    return children
}

// Ensure consist example rendering by providing a dedicated, erm, state
// provider which we prime to a known state, independent of any user
// preferences.
const Example = ({ children, maxWidth, states }: ExampleProps) => {
    const atomvals = [...initials, ...(states ? states : [])] as
        Iterable<readonly [WritableAtom<unknown, [unknown], unknown>, unknown]>

    return (
        <Provider>
            <AtomsHydrator atomvals={atomvals}>
                <Box m={2}>
                    <Card style={{ maxWidth: maxWidth || '100%' }}>
                        <Box m={1} style={{ overflowX: 'auto' }}>
                            {children}
                        </Box>
                    </Card>
                </Box>
            </AtomsHydrator>
        </Provider>
    )
}


const IconBox = styled('span')(({ theme }) => ({
    display: 'inline-block',
    verticalAlign: 'middle',
    fontSize: '70%', // why, CSS, oh why???
    border: `1px solid ${theme.palette.text.disabled}`,
    padding: 1,
    borderRadius: theme.spacing(0.5),

    '& > .MuiSvgIcon-root': {
        verticalAlign: 'middle',
        fontSize: 'calc(100% + 2px)',
    },
}))


const BoxedIcons = ({ children }: { children: ReactNode }) => {
    return <IconBox>{children}</IconBox>
}

const NamespaceExample = ({ type, initial, shared }: {type: NamespaceType; initial: boolean; shared: boolean}) =>
    <NamespaceBadge namespace={{
        nsid: 4026531837,
        type: type,
        ealdorman: {},
        initial: initial,
        parent: null,
        children: [],
    } as unknown as Namespace} shared={shared} />

export const Help = () => (
    <HelpViewer
        chapters={chapters}
        baseroute="/help"
        markdowner={MuiMarkdown}
        shortcodes={{ 
            a: SmartA, 
            BoxedIcons, 
            Example, 
            NamespaceBadge, 
            NamespaceExample 
        }}
        style={{overflow: 'visible'}}
    />
)
