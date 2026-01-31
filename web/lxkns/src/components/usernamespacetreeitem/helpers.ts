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

import type { Namespace, ProcessMap } from "models/lxkns"

// Return the ealdormen processes attached to namespaces owned by the specified
// user namespace.
export const uniqueProcsOfTenants = (usernamespace: Namespace, showSharedNamespaces?: boolean) => {
    const uniqueprocs: ProcessMap = {}
    // When users want to see shared namespaces, then we need to add the
    // ealdorman of this user namespace to its list as a (pseudo) tenant for
    // convenience.
    if (showSharedNamespaces && usernamespace.ealdorman) {
        uniqueprocs[usernamespace.ealdorman.pid] = usernamespace.ealdorman
    }
    usernamespace.tenants.forEach(tenantnamespace => {
        if (tenantnamespace.ealdorman) {
            uniqueprocs[tenantnamespace.ealdorman.pid] = tenantnamespace.ealdorman
        }
    })
    return Object.values(uniqueprocs)
}

