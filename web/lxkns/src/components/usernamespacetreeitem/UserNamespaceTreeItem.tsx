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
import { useAtom } from 'jotai'

import TreeItem from '@material-ui/lab/TreeItem'
import Crown from 'mdi-material-ui/Crown'

import { ProcessInfo } from 'components/processinfo'
import { NamespaceInfo } from 'components/namespaceinfo'
import { compareNamespaceById, compareProcessByNameId, ProcessMap, Namespace, NamespaceType } from 'models/lxkns'
import { showSharedNamespacesAtom } from 'views/settings'
import { makeStyles } from '@material-ui/core'
import { NamespaceBadge } from 'components/namespacebadge'

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
    Object.values(usernamespace.tenants)
        .forEach(tenantnamespace => {
            if (tenantnamespace.ealdorman) {
                uniqueprocs[tenantnamespace.ealdorman.pid] = tenantnamespace.ealdorman
            }
        })
    return Object.values(uniqueprocs)
}

const useStyle = makeStyles((theme) => ({
    owninguserns: {
        paddingLeft: '0.5em',
        color: theme.palette.text.disabled,

        '& .MuiSvgIcon-root': {
            fontSize: 'inherit',
            verticalAlign: 'middle',
            marginRight: '0.1em',
        },
    }
}))

export interface UserNamespaceTreeItemProps {
    /** user namespace object */
    namespace: Namespace
}

// Component UserNamespaceTreeItem renders a user namespace tree item, as well
// as the owned non-user namespaces and child user namespaces.
export const UserNamespaceTreeItem = ({ namespace: usernamespace }: UserNamespaceTreeItemProps) => {

    const classes = useStyle()

    const [showSharedNamespaces] = useAtom(showSharedNamespacesAtom)

    // Generally speaking, we now separate the "tenants" into bind-mounted
    // namespaces and namespaces "inhabited" by processes.

    // Bind-mounted namespaces can be found by checking that a namespace has no
    // ealdorman process.
    const bindmounts = Object.values(usernamespace.tenants)
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
    const procs = uniqueProcsOfTenants(usernamespace, showSharedNamespaces)
        .sort(compareProcessByNameId)
        .map(proc =>
            <TreeItem
                className="controlledprocess"
                key={proc.pid}
                nodeId={`${usernamespace.nsid}-${proc.pid}`}
                label={<ProcessInfo process={proc} />}
            >
                {Object.values(proc.namespaces)
                    // either (a) show all non-user namespaces to which a
                    // process is attached, or (b) show only those non-user
                    // namespaces that are specific to this process and not
                    // "shared" with other leaders in the same user namespace. 
                    .filter(showSharedNamespaces
                        ? (procns: Namespace) => (procns.type !== NamespaceType.user || procns !== usernamespace)
                        : (procns: Namespace) => procns.owner === usernamespace && procns.ealdorman === proc)
                    .sort((procns1, procns2) => procns1.type.localeCompare(procns2.type))
                    .map((procns: Namespace) => <TreeItem
                        className="tenant"
                        key={procns.nsid}
                        nodeId={`${usernamespace.nsid}-${proc.pid}-${procns.nsid}`}
                        label={<>
                            <NamespaceInfo
                                shared={procns.owner !== usernamespace || procns.ealdorman !== proc}
                                noprocess={true}
                                shortprocess={procns.ealdorman !== proc}
                                namespace={procns}
                            />
                            {procns.ealdorman === proc
                                && procns.owner && procns.type !== NamespaceType.user
                                && procns.owner !== usernamespace
                                && <span className={classes.owninguserns}>
                                    <Crown fontSize="inherit" />&nbsp;
                                    <NamespaceBadge
                                        namespace={procns.owner}
                                        tooltipprefix="different owning"
                                    />
                                </span>}
                        </>}
                    />)
                }
            </TreeItem>
        )

    const children = Object.values(usernamespace.children)
        .sort(compareNamespaceById)
        .map(childns => <UserNamespaceTreeItem key={childns.nsid} namespace={childns} />)

    // Please note that we need destructure or concatenate the resulting two
    // sets of tenant nodes and children nodes, as otherwise the enclosing
    // tree item gets fooled into thinking it always has child tree nodes
    // (grrr).
    return (
        <TreeItem
            className="namespace"
            key={usernamespace.nsid}
            nodeId={`${usernamespace.nsid}`}
            label={<NamespaceInfo namespace={usernamespace} />}
        >
            {[...procs, ...bindmounts, ...children]}
        </TreeItem>
    )
}

export default UserNamespaceTreeItem
