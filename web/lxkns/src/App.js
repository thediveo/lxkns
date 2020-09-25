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

import React, { useEffect, useState, useRef } from 'react';
import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import TreeView from '@material-ui/lab/TreeView';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import InfoIcon from '@material-ui/icons/Info';
import { ConfirmProvider, useConfirm } from 'material-ui-confirm';

import './App.css';
import ElevationScroll from './ElevationScroll';
import UsernamespaceItem from './Usernamespace';
import { postDiscovery as finalizeDiscoveryData, namespaceIdOrder } from './model';
import { makeStyles } from '@material-ui/core';


const LxknsApp = () => {
    const confirm = useConfirm();

    const [allns, setAllns] = useState({ namespaces: {}, processes: {} });
    const [expanded, setExpanded] = React.useState([]);

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
            title: 'About Linux Namespaces',
            description:
                <div>
                    <p>Displays all discovered namespaces inside a Linux host.
                    The display is organized following the hierarchy of user
                    namespaces. Namespaces of other types are shown beneath the
                    particular user namespace which is owning them. Owning a
                    namespace here means that a namespace was created by a
                    process while the process was attached to that specific user
                    namespace.</p>
                    <p><a href="https://github.com/thediveo/lxkns"
                        target="_blank" rel="noopener noreferrer">thediveo/lxkns
                    project</a> on Github</p>
                </div>,
            cancellationButtonProps: { className: "hide" }
        }).then(() => { }).catch(() => { });
    };

    // We definitely need "useRef: the unsung Hooks hero"
    // (https://blog.logrocket.com/how-to-get-previous-props-state-with-react-hooks/)
    // here, as we need to calculate the expansion state of the user namespace
    // nodes, based on which nodes were expanded previously and which nodes
    // are to be added anew, and should always be expanded.
    const oldStateRef = useRef();
    useEffect(() => {
        oldStateRef.current = { allns: allns, expanded: expanded };
    });

    // The effect hook runs after the component was rendered for the first
    // time(!) only and we then take the chance to initiate a namespace
    // discovery on the lxkns service API, later updating the discovery state
    // after the data has arrived.
    useEffect(() => {
        // Fetch the namespace+process discovery data from the server,
        // postprocess the JSON result, and finally update the allns state
        // with the new information about all namespaces.
        const fetchDiscoveryData = () => {
            fetch('http://' + window.location.hostname + ':5010/api/namespaces')
                .then(httpresult => httpresult.json())
                .then(jsondata => finalizeDiscoveryData(jsondata))
                .then(discovery => {
                    // We need to determine which user namespace tree nodes
                    // should be expanded, taking into account which user
                    // namespace nodes currently are expanded and which user
                    // namespace nodes are now being added anew.

                    // The difficulty here is that we only know which nodes
                    // are expanded, but we don't know which nodes are
                    // collapsed. So we first need to calculate which user
                    // namespace nodes are really new; we just need the user
                    // namespace IDs, as this is what we're identifying the
                    // tree nodes by.
                    const oldusernsids = Object.values(oldStateRef.current.allns.namespaces)
                        .filter(ns => ns.type === "user")
                        .map(ns => ns.nsid.toString());
                    const addedusernsids = Object.values(discovery.namespaces)
                        .filter(ns => ns.type === "user")
                        .map(ns => ns.nsid.toString())
                        .filter(nsid => !oldusernsids.includes(nsid));
                    // Now we need to combine the "set" of existing expanded
                    // user namespace nodes with the "set" of the newly added
                    // user namespace nodes, as we want all new user namespace
                    // to be automatically expanded on arrival.
                    const expanded = oldStateRef.current.expanded;
                    const expand = expanded.concat(
                        addedusernsids.filter(nsid => !expanded.includes(nsid)));
                    // Phew, finally we can update the state.
                    setAllns(discovery);
                    setExpanded(expand);
                });
        };
        // Initiate getting discovery data NOW...
        fetchDiscoveryData();
        // Set up an interval timer to re-fetch the discovery data from time
        // to time...
        const interval = setInterval(() => fetchDiscoveryData(), 5000);
        return () => clearInterval(interval);
    }, []);

    // In the discovery heap find only the topmost user namespaces; that is,
    // user namespaces without any parent. This should return only one user
    // namespace.
    const rootuserns = Object.values(allns.namespaces)
        .filter(ns => ns.type === "user" && ns.parent === null)
        .sort(namespaceIdOrder)
        .map(ns =>
            <UsernamespaceItem key={ns.nsid.toString()} ns={ns} />
        );

    const classes = useStyles();

    return (
        <React.Fragment>
            <CssBaseline />
            <ElevationScroll>
                <AppBar>
                    <Toolbar>
                        <Typography variant="h6" className={classes.title}>Linux Namespaces</Typography>
                        <IconButton color="inherit" onClick={handleInfo}><InfoIcon /></IconButton>
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
        </React.Fragment>
    );
}

// We need to wrap the application as otherwise we won't get a confirmer...
// ouch.
const Wrapper = () => <ConfirmProvider><LxknsApp /></ConfirmProvider>;

export default Wrapper;

const useStyles = makeStyles((theme) => ({
    title: { flexGrow: 1 }
}));
