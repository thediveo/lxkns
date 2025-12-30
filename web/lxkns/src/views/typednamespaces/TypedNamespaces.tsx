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

import { useLocation } from 'react-router-dom'

import { type Discovery, NamespaceType } from 'models/lxkns'
import { NamespaceProcessTree } from 'components/namespaceprocesstree'
import { MountTreeDetailer } from 'components/mounttree'
import { Box } from '@mui/material'
import type { TreeAPI } from 'app/treeapi'


export interface TypedNamespacesProps {
    /** tree API for expansion, collapsing */
    apiRef?: React.Ref<TreeAPI>
    /** lxkns discovery data */
    discovery: Discovery
}

export const TypedNamespaces = ({ apiRef, discovery }: TypedNamespacesProps) => {

    const loc = useLocation()

    const nstype = loc
        && Object.values(NamespaceType).includes(loc.pathname.substring(1) as NamespaceType)
        && loc.pathname.substring(1)

    return (
        // On changing the type of namespaces to render, we need to force
        // unmounting the existing tree component and remounting a fresh one in
        // order to clear the namespace tree's internal state completely. Yes,
        // this is slightly (w)hacky.
        nstype && <Box pl={1}>
            <NamespaceProcessTree
                key={nstype}
                type={nstype}
                discovery={discovery}
                apiRef={apiRef}
                details={(nstype === 'mnt' && MountTreeDetailer) || undefined}
            />
        </Box>
    )

}
