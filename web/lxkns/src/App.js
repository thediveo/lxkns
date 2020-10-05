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
import { HashRouter as Router, Switch, Route } from 'react-router-dom';

import CssBaseline from '@material-ui/core/CssBaseline';
import Badge from '@material-ui/core/Badge';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import Tooltip from '@material-ui/core/Tooltip';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import Divider from '@material-ui/core/Divider';

import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import PersonIcon from '@material-ui/icons/Person';
import RunFast from 'mdi-material-ui/RunFast';
import InfoIcon from '@material-ui/icons/Info';

import './App.css';
import Discovery, { DiscoveryContext } from 'components/discovery';
import UserNamespaceTree from 'components/usernamespacetree';
import { EXPANDALL_ACTION, COLLAPSEALL_ACTION, treeAction } from 'components/usernamespacetree/UserNamespaceTree';
import ConfinedProcessTree from 'components/confinedprocesstree';
import Refresher from 'components/refresher';
import AppBarDrawer, { DrawerLinkItem } from 'components/appbardrawer';

import version from './version';
import About from './About';

const LxknsApp = () => {
    const [treeaction, setTreeAction] = useState("");

    return (<>
        <AppBarDrawer
            title={
                <DiscoveryContext.Consumer>
                    {value =>
                        <Badge badgeContent={Object.keys(value.namespaces).length} color="secondary">
                            Linux Namespaces
                    </Badge>
                    }
                </DiscoveryContext.Consumer>
            }
            tools={<>
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
            </>}
            drawer={(closeDrawer) => <>
                <List onClick={closeDrawer}>
                    <ListItem>
                        <Typography variant="h6" color="textSecondary">
                            lxkns
                        </Typography>
                    </ListItem>
                    <ListItem>
                        <Typography variant="body2" color="textSecondary">
                            version {version}
                        </Typography>
                    </ListItem>
                    <Divider />
                    <DrawerLinkItem
                        icon={<PersonIcon />}
                        label="user namespaces"
                        path="/"
                    />
                    <DrawerLinkItem
                        icon={<RunFast />}
                        label="pid namespaces"
                        path="/pid"
                    />
                    <DrawerLinkItem
                        icon={<RunFast />}
                        label="control-group namespaces"
                        path="/cgroup"
                    />
                    <DrawerLinkItem
                        icon={<RunFast />}
                        label="network namespaces"
                        path="/net"
                    />
                    <DrawerLinkItem
                        icon={<InfoIcon />}
                        label="information"
                        path="/about"
                    />
                </List>
            </>}
        />
        <Switch>
            <Route exact path="/about" render={() => <About />} />
            {['pid', 'cgroup', 'net'].map(
                type => <Route
                    exact path={`/${type}`}
                    render={() => <ConfinedProcessTree type={type} />}
                />)}
            <Route path="/" render={() => <UserNamespaceTree action={treeaction} />} />
        </Switch>
    </>);
};

// We need to wrap the application as otherwise we won't get a confirmer ...
// ouch. And since we're already at wrapping things, let's just wrap up all the
// other wrapping here... *snicker*.
const App = () => (
    <Router>
        <Discovery>
            <CssBaseline />
            <LxknsApp />
        </Discovery>
    </Router>
);

export default App;
