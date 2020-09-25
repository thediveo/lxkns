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
import { postDiscovery, namespaceIdOrder } from './model';
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

    const handleInfo = () => {
        console.log("info...", confirm);
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

    // The effect hook runs after the component was rendered and we use that
    // chance to initiate a namespace discovery on the lxkns service API and
    // then update everything...
    useEffect(() => {
        const namespaceDiscovery = async () => {
            const response = await fetch(
                'http://' + window.location.hostname + ':5010/api/namespaces');
            const jsondata = await response.json();
            const newallns = postDiscovery(jsondata);
    
            // Expand all new user namespace nodes and keep all existing user
            // nodes in their current expansion state: for this, we need to
            // look at the delta between the new list of user namespaces and
            // the old list of namespaces.
            const olduserns = Object.values(allns.namespaces)
                .filter(ns => ns.type === "user")
                .map(ns => ns.nsid.toString());
            const reallynewuserns = Object.values(newallns.namespaces)
                .filter(ns => ns.type === "user")
                .map(ns => ns.nsid.toString())
                .filter(nsid => !olduserns.includes(nsid));
            const toexpand = reallynewuserns.concat(
                expanded.filter(nsid => !reallynewuserns.includes(nsid)));
    
            setAllns(newallns);
            //setExpanded(toexpand);
        };
    
        namespaceDiscovery();
        const interval = setInterval(() => namespaceDiscovery(), 2000);
        return () => clearInterval(interval);
    });

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
