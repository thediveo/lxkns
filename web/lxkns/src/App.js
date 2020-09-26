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

import React, { useEffect, useState } from 'react';
import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Badge from '@material-ui/core/Badge';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import TreeView from '@material-ui/lab/TreeView';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import InfoIcon from '@material-ui/icons/Info';
import LaunchIcon from '@material-ui/icons/Launch';
import { ConfirmProvider, useConfirm } from 'material-ui-confirm';
import { makeStyles, Tooltip } from '@material-ui/core';

import './App.css';
import LxknsIcon from './lxkns.svg';
import ElevationScroll from './tools/ElevationScroll';
import UsernamespaceItem from './Usernamespace';
import { postDiscovery as finalizeDiscoveryData, namespaceIdOrder } from './model';
import useInterval from './tools/useInterval';


const LxknsApp = () => {
    const confirm = useConfirm();

    const [discovery, setDiscovery] = useState({ namespaces: {}, processes: {} });
    const [expanded, setExpanded] = useState([]);
    const [cycle, setCyle] = useState(5000);

    // Whenever the user clicks on the expand/close icon next to a tree item,
    // update the tree's expand state accordingly. This allows us to
    // explicitly take back control (ha ... hah ... HAHAHAHA!!!) of the expansion
    // state of the tree.
    const handleToggle = (event, nodeIds) => {
        setExpanded(nodeIds);
    };

    // Shows the "About" dialog with a short description of this application.
    const handleInfo = () => {
        confirm({
            title: <><img src={LxknsIcon} alt="lxkns app logo" style={{ verticalAlign: 'text-bottom' }} />
            &nbsp;About Linux Namespaces</>,
            description:
                <div>
                    <p>This app displays all discovered namespaces inside a
                    Linux host.</p>
                    <p>The display is organized following the hierarchy of user
                    namespaces. Namespaces of other types are shown beneath the
                    particular user namespace which is owning them. Owning a
                    namespace here means that a namespace was created by a
                    process while the process was attached to that specific user
                    namespace.</p>
                    <p>Find the <LaunchIcon fontSize="inherit" className="inlineicon" /><a href="https://github.com/thediveo/lxkns"
                        target="_blank" rel="noopener noreferrer">thediveo/lxkns
                    project</a> on Github</p>
                </div>,
            cancellationButtonProps: { className: "hide" }
        });
    };

    // Fetch the namespace+process discovery data from the server,
    // postprocess the JSON result, and finally update the allns state
    // with the new information about all namespaces. And then update also
    // the expansion state of the user namespace tree nodes. And all this
    // in a react-allowed stateless manner...
    const fetchDiscoveryData = () => {
        fetch('/api/namespaces')
            .then(httpresult => httpresult.json())
            .then(jsondata => finalizeDiscoveryData(jsondata))
            .then(discovery => setDiscovery(prevAllns => {
                // We need to determine which user namespace tree nodes
                // should be expanded, taking into account which user
                // namespace nodes currently are expanded and which user
                // namespace nodes are now being added anew. The
                // difficulty here is that we only know which nodes are
                // expanded, but we don't know which nodes are collapsed.
                // So we first need to calculate which user namespace
                // nodes are really new; we just need the user namespace
                // IDs, as this is what we're identifying the tree nodes
                // by.
                const oldusernsids = Object.values(prevAllns.namespaces)
                    .filter(ns => ns.type === "user")
                    .map(ns => ns.nsid.toString());
                const fltr = Object.keys(prevAllns.namespaces).length ?
                    (ns => ns.type === "user") : (ns => ns.type === "user" && ns.parent === null);
                const addedusernsids = Object.values(discovery.namespaces)
                    .filter(fltr)
                    .map(ns => ns.nsid.toString())
                    .filter(nsid => !oldusernsids.includes(nsid));
                // Now we need to combine the "set" of existing expanded
                // user namespace nodes with the "set" of the newly added
                // user namespace nodes, as we want all new user namespace
                // to be automatically expanded on arrival.
                setExpanded(prevExpanded => prevExpanded.concat(
                    addedusernsids.filter(nsid => !prevExpanded.includes(nsid))));
                // Finally return new discovery state to be set.
                return discovery;
            }));
    };

    // Get new discovery data after some time; please note that useInterval
    // interprets a null cycle as switching off the timer.
    useInterval(() => fetchDiscoveryData(), cycle);

    // Initially fetch discovery data, unless the cycle is null.
    useEffect(() => fetchDiscoveryData(), [cycle]);

    // Collapse (almost) all user namespace nodes, except for the top-level user
    // namespace.
    const handleCollapseAll = () => {
        const topuserns = Object.values(discovery.namespaces)
            .filter(ns => ns.type === "user" && ns.parent === null)
            .map(ns => ns.nsid.toString())
        setExpanded(topuserns);
    };

    // Expand all user namespaces nodes.
    const handleExpandAll = () => {
        console.log("allns", Object.values(discovery.namespaces).map(ns => ns.nsid));
        const alluserns = Object.values(discovery.namespaces)
            .filter(ns => ns.type === "user")
            .map(ns => ns.nsid.toString())
        setExpanded(alluserns);
    };

    // In the discovery heap find only the topmost user namespaces; that is,
    // user namespaces without any parent. This should return only one user
    // namespace.
    const rootuserns = Object.values(discovery.namespaces)
        .filter(ns => ns.type === "user" && ns.parent === null)
        .sort(namespaceIdOrder)
        .map(ns =>
            <UsernamespaceItem key={ns.nsid.toString()} ns={ns} />
        );

    const classes = useStyles();

    return (
        <>
            <CssBaseline />
            <ElevationScroll>
                <AppBar>
                    <Toolbar>
                        <Typography variant="h6" className={classes.title}>
                            <Badge badgeContent={Object.keys(discovery.namespaces).length} color="secondary">
                                Linux Namespaces
                            </Badge>
                        </Typography>
                        <Tooltip title="expand initial user namespace(s) only">
                            <IconButton color="inherit" onClick={handleCollapseAll}><ChevronRightIcon /></IconButton>
                        </Tooltip>
                        <Tooltip title="expand all">
                            <IconButton color="inherit" onClick={handleExpandAll}><ExpandMoreIcon /></IconButton>
                        </Tooltip>
                        <Tooltip title="about lxkns">
                            <IconButton color="inherit" onClick={handleInfo}><InfoIcon /></IconButton>
                        </Tooltip>
                    </Toolbar>
                </AppBar>
            </ElevationScroll>
            <Toolbar />
            <TreeView
                className="namespacetree"
                onNodeToggle={handleToggle}
                defaultCollapseIcon={<ExpandMoreIcon />}
                defaultExpandIcon={<ChevronRightIcon />}
                expanded={expanded}
            >{rootuserns}</TreeView>
        </>
    );
}

// We need to wrap the application as otherwise we won't get a confirmer...
// ouch.
const Wrapper = () => <ConfirmProvider><LxknsApp /></ConfirmProvider>;

export default Wrapper;

const useStyles = makeStyles((theme) => ({
    title: { flexGrow: 1 }
}));
