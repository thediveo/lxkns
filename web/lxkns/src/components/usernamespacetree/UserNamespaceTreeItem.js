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

import Namespace from 'components/namespace';
import { namespaceIdOrder, namespaceNameTypeIdOrder } from 'components/discovery/model';

// Component UserNamespaceTreeItem renders a user namespace tree item, as well
// as the owned non-user namespaces and child user namespaces.
export const UserNamespaceTreeItem = ({namespace}) => {
    const tenants = Object.values(namespace.tenants)
        .sort(namespaceNameTypeIdOrder)
        .map(tenantns => <TreeItem
            className="tenant"
            key={tenantns.nsid}
            nodeId={tenantns.nsid.toString()}
            label={<Namespace namespace={tenantns} />}
        />);

    const children = Object.values(namespace.children)
        .sort(namespaceIdOrder)
        .map(childns => <UserNamespaceTreeItem key={childns.nsid} namespace={childns} />);

    // Please note that we need destructure or concatenate the resulting two
    // sets of tenant nodes and children nodes, as otherwise the enclosing
    // tree item gets fooled into thinking it always has child tree nodes
    // (grrr).
    return (
        <TreeItem
            key={namespace.nsid}
            nodeId={namespace.nsid.toString()}
            label={<Namespace namespace={namespace} />}
        >
            {[...tenants, ...children]}
        </TreeItem>);
}

export default UserNamespaceTreeItem;
