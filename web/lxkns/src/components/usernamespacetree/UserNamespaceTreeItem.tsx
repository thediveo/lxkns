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

import React from 'react';
import TreeItem from '@material-ui/lab/TreeItem';

import NamespaceInfo from 'components/namespaceinfo/NamespaceInfo';
import ProcessInfo from 'components/processinfo'
import { compareNamespaceById, compareProcessByNameId, ProcessMap, Namespace } from 'models/lxkns';

// Return the ealdormen processes attached to namespaces owned by the specified
// user namespace.
export const uniqueProcsOfTenants = (usernamespace: Namespace) => {
    const uniqueprocs = {} as ProcessMap
    Object.values(usernamespace.tenants)
        .filter(tenant => tenant.ealdorman !== null)
        .forEach(tenant =>
            uniqueprocs[tenant.ealdorman.pid] = tenant.ealdorman);
    return Object.values(uniqueprocs);
};

export interface UserNamespaceTreeItemProps {
    namespace: Namespace
}

// Component UserNamespaceTreeItem renders a user namespace tree item, as well
// as the owned non-user namespaces and child user namespaces.
export const UserNamespaceTreeItem = ({ namespace }: UserNamespaceTreeItemProps) => {
    // Generally speaking, we now separate the "tenants" into bind-mounted
    // namespaces and namespaces "inhabited" by processes.

    // Bind-mounted namespaces can be found by checking that a namespace has no
    // ealdorman process.
    const bindmounts = Object.values(namespace.tenants)
        .filter(tenant => tenant.ealdorman === null)
        .sort(compareNamespaceById)
        .map(tenant => <TreeItem
            className="tenant"
            key={tenant.nsid}
            nodeId={tenant.nsid.toString()}
            label={<NamespaceInfo namespace={tenant} />}
        />);

    // We now want to organize namespaces with processes joined to them by these
    // processes, because that might be better fitting user expectations when
    // navigating the discovery information. So, we collect all ealdorman
    // processes and then sort them by their names and PIDs, and then we start
    // rendering the process nodes.
    const procs = uniqueProcsOfTenants(namespace)
        .sort(compareProcessByNameId)
        .map(proc =>
            <TreeItem
                className="tenantprocess"
                key={proc.pid}
                nodeId={namespace.nsid.toString() + '-' + proc.pid.toString()}
                label={<ProcessInfo process={proc} />}>{
                    Object.values(proc.namespaces)
                        .filter(tenant => tenant.owner === namespace && tenant.ealdorman === proc)
                        .sort((tenant1, tenant2) => tenant1.type.localeCompare(tenant2.type))
                        .map(tenant => <TreeItem
                            className="tenant"
                            key={tenant.nsid}
                            nodeId={tenant.nsid.toString()}
                            label={<NamespaceInfo noprocess={true} namespace={tenant} />}
                        />)
                }</TreeItem>
        );

    const children = Object.values(namespace.children)
        .sort(compareNamespaceById)
        .map(childns => <UserNamespaceTreeItem key={childns.nsid} namespace={childns} />);

    // Please note that we need destructure or concatenate the resulting two
    // sets of tenant nodes and children nodes, as otherwise the enclosing
    // tree item gets fooled into thinking it always has child tree nodes
    // (grrr).
    return (
        <TreeItem
            key={namespace.nsid}
            nodeId={namespace.nsid.toString()}
            label={<NamespaceInfo namespace={namespace} />}
        >
            {[...procs, ...bindmounts, ...children]}
        </TreeItem>
    )
}

export default UserNamespaceTreeItem;
