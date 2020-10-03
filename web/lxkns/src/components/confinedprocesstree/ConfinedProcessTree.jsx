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

import React, { useContext } from 'react';

import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';

import TreeView from '@material-ui/lab/TreeView';
import TreeItem from '@material-ui/lab/TreeItem';

import { DiscoveryContext } from 'components/discovery';
import { namespaceIdOrder, processNameIdOrder } from 'components/lxkns';
import Namespace from 'components/namespace';
import ProcessInfo from 'components/process'

// Returns all processes and child/grandchild/... processes which have different
// controllers, but are still within the same PID namespace of the original
// process.
const findProcesses = (proc) => {
    let procs = []
    proc.children.forEach(childproc => {
        if (childproc.namespaces.pid !== proc.namespaces.pid) {
            return;
        }
        if (childproc.cgroup !== proc.cgroup) {
            // Child process is still within the same PID namespace, but has a
            // different controller: looks interesting 8).
            procs.push(childproc);
        }
        // Check to see if there are any children which are still in the same
        // PID namespace, but which have different controllers.
        procs = procs.concat(findProcesses(childproc))
    });
    return procs;
};

const confinedProcess = (process) =>
    <TreeItem nodeId={process.pid.toString()} label={process.name + " (" + process.pid + ") " + process.cgroup} />;

export const confinedProcessTreeItem = namespace => {

    const processes = namespace.leaders
        .concat(namespace.leaders.map(leader => findProcesses(leader)))
        .flat(1)
        .sort(processNameIdOrder)
        .map(proc => confinedProcess(proc));

    const pidns = namespace.children
        .map(pidns => confinedProcessTreeItem(pidns));

    return (
        <TreeItem nodeId={namespace.nsid.toString()} label={<Namespace namespace={namespace} />}>
            {processes}{pidns}
        </TreeItem>
    );
};

export const ConfinedProcessTree = () => {

    // Discovery data comes in via a dedicated discovery context.
    const discovery = useContext(DiscoveryContext);

    const rootpidnsItems = Object.values(discovery.namespaces)
        .filter(ns => ns.type === "pid" && ns.parent === null)
        .sort(namespaceIdOrder)
        .map(pidns => confinedProcessTreeItem(pidns));

    return (
        <TreeView
            className="namespacetree"
            defaultCollapseIcon={<ExpandMoreIcon />}
            defaultExpandIcon={<ChevronRightIcon />}
        >{rootpidnsItems}</TreeView>
    );

};

export default ConfinedProcessTree;
