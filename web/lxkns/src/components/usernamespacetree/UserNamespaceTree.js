import React, { useEffect, useState, useContext, useRef } from 'react';

import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';

import TreeView from '@material-ui/lab/TreeView';

import { DiscoveryContext } from 'components/discovery';
import { namespaceIdOrder } from 'components/discovery/model';
import { UserNamespaceTreeItem } from './UserNamespaceTreeItem';

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
            const alluserns = Object.values(discovery.namespaces)
                .filter(ns => ns.type === "user")
                .map(ns => ns.nsid.toString())
            setExpanded(alluserns);
        } else if (action.startsWith(COLLAPSEALL_ACTION)) {
            const topuserns = Object.values(discovery.namespaces)
                .filter(ns => ns.type === "user" && ns.parent === null)
                .map(ns => ns.nsid.toString())
            setExpanded(topuserns);
        }
    }, [action, discovery]);

    // After updaing the discovery information, check if there are new user
    // namespaces which we want to automatically expand in the tree view. We
    // won't touch the expansion state of existing user namespace tree nodes.
    useEffect(() => {
        // We want to determine which user namespace tree nodes should be expanded,
        // taking into account which user namespace nodes currently are expanded and
        // which user namespace nodes are now being added anew. The difficulty here
        // is that we only know which nodes are expanded, but we don't know which
        // nodes are collapsed. So we first need to calculate which user namespace
        // nodes are really new; we just need the user namespace IDs, as this is
        // what we're identifying the tree nodes by.
        const oldusernsids = Object.values(discovery.previousNamespaces)
            .filter(ns => ns.type === "user")
            .map(ns => ns.nsid.toString());
        // Initially open all root namespaces, but lateron never touch that state
        // again.
        const fltr = Object.keys(discovery.previousNamespaces).length ?
            (ns => ns.type === "user") : (ns => ns.type === "user" && ns.parent === null);
        const addedusernsids = Object.values(discovery.namespaces)
            .filter(fltr)
            .map(ns => ns.nsid.toString())
            .filter(nsid => !oldusernsids.includes(nsid));
        // Now we need to combine the "set" of existing expanded user namespace
        // nodes with the "set" of the newly added user namespace nodes, as we want
        // all new user namespace to be automatically expanded on arrival.
        setExpanded(prevExpanded =>
            prevExpanded.concat(addedusernsids.filter(nsid => !prevExpanded.includes(nsid))));
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
        .map(ns => <UserNamespaceTreeItem key={ns.nsid.toString()} namespace={ns} />
        );

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
