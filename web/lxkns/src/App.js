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

import React, { useState } from 'react';

import { makeStyles } from '@material-ui/core';

import CssBaseline from '@material-ui/core/CssBaseline';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Badge from '@material-ui/core/Badge';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import Tooltip from '@material-ui/core/Tooltip';

import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import InfoIcon from '@material-ui/icons/Info';
import LaunchIcon from '@material-ui/icons/Launch';

import { ConfirmProvider, useConfirm } from 'material-ui-confirm';

import './App.css';
import LxknsIcon from './lxkns.svg';
import Discovery, { DiscoveryContext } from 'components/discovery';
import UserNamespaceTree from 'components/usernamespacetree';
import { EXPANDALL_ACTION, COLLAPSEALL_ACTION, treeAction } from 'components/usernamespacetree/UserNamespaceTree';
import ElevationScroll from 'components/elevationscroll';
import Refresher from 'components/refresher';

const LxknsApp = () => {
    const confirm = useConfirm();

    const [treeaction, setTreeAction] = useState("");

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

    const classes = useStyles();

    return (
        <Discovery>
            <CssBaseline />
            <ElevationScroll>
                <AppBar>
                    <Toolbar>
                        <DiscoveryContext.Consumer>
                            {value =>
                                <Typography variant="h6" className={classes.title}>
                                    <Badge badgeContent={Object.keys(value.namespaces).length} color="secondary">
                                        Linux Namespaces
                                    </Badge>
                                </Typography>
                            }
                        </DiscoveryContext.Consumer>
                        <Tooltip title="expand initial user namespace(s) only">
                            <IconButton color="inherit"
                                onClick={() => setTreeAction(treeAction(COLLAPSEALL_ACTION))}>
                                <ChevronRightIcon />
                            </IconButton>
                        </Tooltip>
                        <Tooltip title="expand all">
                            <IconButton color="inherit"
                                onClick={() => setTreeAction(treeAction(EXPANDALL_ACTION))}>
                                <ExpandMoreIcon />
                            </IconButton>
                        </Tooltip>
                        <Refresher />
                        <Tooltip title="about lxkns">
                            <IconButton color="inherit" onClick={handleInfo}><InfoIcon /></IconButton>
                        </Tooltip>
                    </Toolbar>
                </AppBar>
            </ElevationScroll>
            <Toolbar />
            <UserNamespaceTree action={treeaction} />
        </Discovery >
    );
};

// We need to wrap the application as otherwise we won't get a confirmer...
// ouch.
const Wrapper = () => <ConfirmProvider><LxknsApp /></ConfirmProvider>;

export default Wrapper;

const useStyles = makeStyles((theme) => ({
    title: { flexGrow: 1 }
}));
