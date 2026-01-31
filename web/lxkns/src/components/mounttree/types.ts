// Copyright 2025 Harald Albrecht.
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

import type { NamespaceProcessTreeDetailComponentProps, NamespaceProcessTreeTreeDetails } from "components/namespaceprocesstree";
import type { NamespaceMap } from "models/lxkns";
import { MountTree } from "./MountTree";

export type MountTreeProps = NamespaceProcessTreeDetailComponentProps

/**
 * Returns the list of all tree node ids to be expanded. However, mount points
 * which cross a certain threshold of child mount points won't be expanded
 * though. 
 */
const expandAll = (namespaces: NamespaceMap) => Object.values(namespaces)
    .map(ns => ns.mountpaths
        ? Object.values(ns.mountpaths)
            .map(mountpath => mountpath.mounts
                .filter(mountpoint => mountpoint.children.length > 0 && mountpoint.children.length <= 50)
                .map(mountpoint => `${ns.nsid}-${mountpoint.mountpoint}-${mountpoint.mountid}`)
            ).flat()
        : [])
    .flat()

/**
 * This detailer:
 * - provides a factory to render the mount point details of mount namespaces,
 * - supports expanding all detail nodes (well, at least if they don't contain
 *   more than a certain maximum of child mount points).
 * - supports collapsing all detail nodes.
 */
export const MountTreeDetailer: NamespaceProcessTreeTreeDetails = {
    factory: MountTree,
    expandAll: expandAll,
    collapseAll: undefined,
}

