// Copyright 2021 Harald Albrecht.
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

import type { NamespaceMap } from 'models/lxkns/model'
import { NamespaceInfo } from 'components/namespaceinfo'


const namespacere = /^[a-z]{3,6}:\[(\d+)\]$/


export interface MountpointRootProps {
    /** root path of a mount point */
    root: string
    /** 
     * map of all discovered namespaces for mountpoint namespace root path
     * lookups.
     */
    namespaces: NamespaceMap
}

/**
 * Renders a mount point root path, detecting roots which are Linux-kernel
 * namespaces, coming from the special nsfs filesystem. Such namespace roots are
 * then not rendered as a path but instead as a namespace badge as used in all
 * other places of the UI.
 */
export const MountpointRoot = ({ root, namespaces }: MountpointRootProps) => {

    const match = root.match(namespacere)
    const namespace = match && namespaces[match[1]]

    return namespace
        ? <NamespaceInfo shortprocess={true} namespace={namespace} />
        : <>{root}</>
}

export default MountpointRoot
