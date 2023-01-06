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

import TreeItem from '@mui/lab/TreeItem'
import { OwnerIcon } from 'icons/Owner'

import { ProcessInfo } from 'components/processinfo'
import { NamespaceInfo } from 'components/namespaceinfo'
import { compareNamespaceById, ProcessMap, Namespace, NamespaceType, isPassive, TaskMap, Busybody, isProcess, compareBusybodies } from 'models/lxkns'
import { showSharedNamespacesAtom } from 'views/settings'
import { NamespaceBadge } from 'components/namespacebadge'
import { styled } from '@mui/material'
import { TaskInfo } from 'components/taskinfo'

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

// Return the unique tasks for all loose threads of the specified user namespace
// as well as all of its tenant namespaces.
const uniqueTasks = (usernamespace: Namespace) => {
    const uniquetasks: TaskMap = {}
    usernamespace.loosethreads
        .concat(usernamespace.tenants.map(tenantnamespace => tenantnamespace.loosethreads).flat())
        .forEach(task => uniquetasks[task.tid] = task)
    return Object.values(uniquetasks)
}

const OwningUserNamespace = styled('span')(({ theme }) => ({
    paddingLeft: '0.5em',
    color: theme.palette.text.disabled,

    '& .MuiSvgIcon-root': {
        fontSize: 'inherit',
        verticalAlign: 'middle',
        marginRight: '0.1em',
    },
}))


export interface UserNamespaceTreeItemProps {
    /** user namespace object */
    namespace: Namespace
    /** information about all processes (for some render support) */
    processes?: ProcessMap,
}

// Component UserNamespaceTreeItem renders a user namespace tree item, as well
// as the owned non-user namespaces and child user namespaces.
export const UserNamespaceTreeItem = ({ namespace: usernamespace, processes }: UserNamespaceTreeItemProps) => {
    const [showSharedNamespaces] = useAtom(showSharedNamespacesAtom)

    // Generally speaking, we now separate the "tenants" into (a) passive
    // namespaces and (b) namespaces with processes or tasks attached to them.
    // The presentation differs between (a) and (b) in that for (b) we present
    // the process or task attached to a particular namespace.

    // "Passive" namespaces (for lack of better terminology) are without any
    // process or task attached to it.
    const passives = Object.values(usernamespace.tenants)
        .filter(tenant => isPassive(tenant))
        .sort(compareNamespaceById)
        .map(tenant => <TreeItem
            className="tenant"
            key={tenant.nsid}
            nodeId={tenant.nsid.toString()}
            label={<NamespaceInfo namespace={tenant} processes={processes} />}
        />);

    // We now want to organize namespaces with processes joined to them by these
    // processes, because that might be better fitting user expectations when
    // navigating the discovery information. So, we collect all ealdorman
    // processes and then sort them by their names and PIDs, and then we start
    // rendering the process nodes.
    //
    // Also render loose threads that are either attached to our user namespace
    // here, or that are attached to one of the namespaces owned by our user
    // namespace.
    const busybodies = (uniqueProcsOfTenants(usernamespace, showSharedNamespaces) as Busybody[])
        .concat(uniqueTasks(usernamespace))
        .sort(compareBusybodies)
        .map(busybody => isProcess(busybody)
            ? <TreeItem
                className="controlledprocess"
                key={busybody.pid}
                nodeId={`${usernamespace.nsid}-${busybody.pid}`}
                label={<ProcessInfo process={busybody} />}
            >
                {Object.values(busybody.namespaces)
                    // either (a) show all non-user namespaces to which a
                    // process is attached, or (b) show only those non-user
                    // namespaces that are specific to this process and not
                    // "shared" with other leaders in the same user namespace. 
                    .filter(showSharedNamespaces
                        ? (procns: Namespace) => (procns.type !== NamespaceType.user || procns !== usernamespace)
                        : (procns: Namespace) => procns.owner === usernamespace && procns.ealdorman === busybody)
                    .sort((procns1, procns2) => procns1.type.localeCompare(procns2.type))
                    .map((procns: Namespace) => <TreeItem
                        className="tenant"
                        key={procns.nsid}
                        nodeId={`${usernamespace.nsid}-${busybody.pid}-${procns.nsid}`}
                        label={<>
                            <NamespaceInfo
                                shared={procns.owner !== usernamespace || procns.ealdorman !== busybody}
                                noprocess={true}
                                shortprocess={procns.ealdorman !== busybody}
                                namespace={procns}
                                processes={processes}
                            />
                            {procns.ealdorman === busybody
                                && procns.owner && procns.type !== NamespaceType.user
                                && procns.owner !== usernamespace
                                && <OwningUserNamespace>
                                    <OwnerIcon fontSize="inherit" />&nbsp;
                                    <NamespaceBadge
                                        namespace={procns.owner}
                                        tooltipprefix="different owning"
                                    />
                                </OwningUserNamespace>}
                        </>}
                    />)
                }
            </TreeItem>
            : <TreeItem
                className="controlledtask"
                key={busybody.tid}
                nodeId={`${usernamespace.nsid}-${busybody.tid}`}
                label={<TaskInfo task={busybody} />}
            >
                {Object.values(busybody.namespaces)
                    .filter(showSharedNamespaces
                        ? (taskns: Namespace) => (taskns.type !== NamespaceType.user || taskns !== usernamespace)
                        : (taskns: Namespace) => taskns.owner === usernamespace /* FIXME:??? */)
                    .sort((taskns1, taskns2) => taskns1.type.localeCompare(taskns2.type))
                    .map((taskns: Namespace) => {
                        const selftask = taskns.loosethreads.includes(busybody)
                        return <TreeItem
                            className="tenant"
                            key={taskns.nsid}
                            nodeId={`${usernamespace.nsid}-${busybody.tid}-${taskns.nsid}`}
                            label={<>
                                <NamespaceInfo
                                    shared={taskns.owner !== usernamespace || !selftask}
                                    noprocess={true}
                                    shortprocess={!selftask}
                                    namespace={taskns}
                                    processes={processes}
                                />
                                {/* TODO:??? */}
                            </>}
                        />
                    })
                }
            </TreeItem>
        )

    const children = Object.values(usernamespace.children)
        .sort(compareNamespaceById)
        .map(childns => <UserNamespaceTreeItem
            key={childns.nsid}
            namespace={childns}
            processes={processes}
        />)

    // Please note that we need destructure or concatenate the resulting two
    // sets of tenant nodes and children nodes, as otherwise the enclosing
    // tree item gets fooled into thinking it always has child tree nodes
    // (grrr).
    return (
        <TreeItem
            className="namespace"
            key={usernamespace.nsid}
            nodeId={`${usernamespace.nsid}`}
            label={<NamespaceInfo namespace={usernamespace} processes={processes} />}
        >
            {[...busybodies, ...passives, ...children]}
        </TreeItem>
    )
}

export default UserNamespaceTreeItem
