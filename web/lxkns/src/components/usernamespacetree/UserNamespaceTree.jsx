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

import React, { useEffect, useState, useContext, useRef } from 'react';

import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';

import TreeView from '@material-ui/lab/TreeView';

import { DiscoveryContext } from 'components/discovery';
import { namespaceIdOrder } from 'components/lxkns';
import { UserNamespaceTreeItem, uniqueProcsOfTenants } from './UserNamespaceTreeItem';

export const EXPANDALL_ACTION = "expandall";
export const COLLAPSEALL_ACTION = "collapseall";

// treeAction returns the specified tree action with some noise tacked on,
// ensuring that the tree component state will change and the component then can
// pick up the "new" command. This IS ugly, no chance to paint enough lipstick
// on this pig.
export const treeAction = (action) => action + Math.floor(100000 + Math.random() * 900000).toString();

// The UserNamespaceTree component renders a tree of user namespaces, including
// owned non-user namespaces. Furthermore, it renders additional information,
// such as about the most-senior leader processes in the namespaces.
//
// The discovery information to be rendered is picked up via a discovery
// context.
//
// This component also supports sending action commands for expanding or
// collapsing (almost) all user namespaces via the properties mechanism.
export const UserNamespaceTree = ({ action }) => {

    // Discovery data comes in via a dedicated discovery context.
    const discovery = useContext(DiscoveryContext);

    // Previous discovery information, if any.
    const previousDiscovery = useRef({ namespaces: {}, processes: {} });

    // Tree node expansion is a component-local state.
    const [expanded, setExpanded] = useState([]);

    // To emulate actions via react's properties architecture and then getting
    // the dependencies correct, we need to store the previous action. Sigh,
    // bloat react-ion.
    const oldaction = useRef("");

    // Trigger an action when the action "state" changes; we are ignoing any
    // stuff appended to the commands, as we need to add noise to the commands
    // in order to make state changes trigger. Oh, well, bummer.
    useEffect(() => {
        if (action === oldaction.current) {
            return;
        }
        oldaction.current = action;
        if (action.startsWith(EXPANDALL_ACTION)) {
            // expand all user namespaces and all included process nodes.
            const alluserns = Object.values(discovery.namespaces)
                .filter(ns => ns.type === "user")
                .map(ns => ns.nsid.toString())
            const allealdormen = Object.values(discovery.namespaces)
                .filter(ns => ns.type !== "user" && ns.ealdorman !== null)
                .map(ns => ns.owner.nsid.toString() + "-" + ns.ealdorman.pid.toString());
            setExpanded(alluserns.concat(allealdormen));
        } else if (action.startsWith(COLLAPSEALL_ACTION)) {
            const topuserns = Object.values(discovery.namespaces)
                .filter(ns => ns.type === "user" && ns.parent === null)
                .map(ns => ns.nsid.toString())
            setExpanded(topuserns);
        }
    }, [action, discovery]);

    // After updaing the discovery information, check if there are any new user
    // namespaces (including their sub items grouping non-user namespaces by
    // processes) which we want to automatically expand in the tree view. We
    // won't touch the expansion state of existing user namespace tree nodes.
    useEffect(() => {
        // We want to determine which user namespace tree nodes should be expanded,
        // taking into account which user namespace nodes currently are expanded and
        // which user namespace nodes are now being added anew. The difficulty here
        // is that we only know which nodes are expanded, but we don't know which
        // nodes are collapsed. So we first need to calculate which user namespace
        // nodes are really new; we just need the user namespace IDs, as this is
        // what we're identifying the tree nodes by.
        const previousNamespaces = previousDiscovery.current.namespaces;
        const oldUsernsIds = Object.values(previousNamespaces)
            .filter(ns => ns.type === 'user')
            .map(ns => ns.nsid);
        // Initially open all root namespaces, but lateron never touch that
        // state again. For this, we set up a filter function either initially
        // letting pass only the root user namespaces, lateron we let pass all
        // user namespaces; we'll next sort out which user namespaces are
        // actually new, as to not touch existing user namespaces.
        const usernsCandidatesFilter = Object.keys(previousNamespaces).length ?
            (ns => ns.type === 'user') : (ns => ns.type === 'user' && ns.parent === null);
        const expandingUserns = Object.values(discovery.namespaces)
            .filter(usernsCandidatesFilter)
            .filter(ns => !oldUsernsIds.includes(ns.nsid));
        // Additionally also open any process child nodes below the new user
        // namespace tree nodes.
        const expandingProcIds = expandingUserns
            .map(userns => uniqueProcsOfTenants(userns)
                .map(proc => userns.nsid.toString() + "-" + proc.pid.toString()))
            .flat();
        // Finally update the expansion state of the tree; this must include the
        // already expanded nodes (state), and we add our to-be-expanded-soon
        // nodes.
        const expandNodeIds = expandingUserns.map(userns => userns.nsid.toString())
            .concat(expandingProcIds);
        setExpanded(previouslyExpanded => previouslyExpanded.concat(expandNodeIds));
        previousDiscovery.current = discovery;
    }, [discovery]);

    // Whenever the user clicks on the expand/close icon next to a tree item,
    // update the tree's expand state accordingly. This allows us to
    // explicitly take back control (ha ... hah ... HAHAHAHA!!!) of the expansion
    // state of the tree.
    const handleToggle = (event, nodeIds) => {
        setExpanded(nodeIds);
    };

    // In the discovery heap find only the topmost user namespaces; that is,
    // user namespaces without any parent. This should return only one user
    // namespace (but covers its a** in case a discovery might someday turn up
    // multiple user namespaces without parents, due to bind-mounting some which
    // are ourside the reach of the discoverer).
    const rootusernsItems = Object.values(discovery.namespaces)
        .filter(ns => ns.type === "user" && ns.parent === null)
        .sort(namespaceIdOrder)
        .map(ns => <UserNamespaceTreeItem key={ns.nsid.toString()} namespace={ns} />);

    return (
        <TreeView
            className="namespacetree"
            onNodeToggle={handleToggle}
            defaultCollapseIcon={<ExpandMoreIcon />}
            defaultExpandIcon={<ChevronRightIcon />}
            expanded={expanded}
        >{rootusernsItems}</TreeView>
    );
};

export default UserNamespaceTree;
