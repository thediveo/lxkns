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
import { BrowserRouter as Router, Switch, Route, useLocation } from 'react-router-dom'

import useErrorBoundary from "use-error-boundary"

import { SnackbarProvider } from 'notistack'

import { Provider as StateProvider, useAtom } from 'jotai'

import CssBaseline from '@material-ui/core/CssBaseline'
import Badge from '@material-ui/core/Badge'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import Tooltip from '@material-ui/core/Tooltip'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import { Box, Divider, ThemeProvider } from '@material-ui/core'
import Toggle from '@material-ui/core/Switch'

import HomeIcon from '@material-ui/icons/Home'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'
import ChevronRightIcon from '@material-ui/icons/ChevronRight'
import InfoIcon from '@material-ui/icons/Info'

import './App.css'
import lxknsTheme from './appstyles'

import Discovery, { useDiscovery } from 'components/discovery'
import UserNamespaceTree from 'components/usernamespacetree'
import NamespaceProcessTree from 'components/namespaceprocesstree'
import Refresher from 'components/refresher'
import AppBarDrawer, { DrawerLinkItem } from 'components/appbardrawer'
import { CreateNamespaceTypeIcon } from 'components/namespaceinfo'
import { NamespaceType } from 'models/lxkns'

import version from '../version'
import About from './About'
import { useTreeAction, EXPANDALL, COLLAPSEALL } from './treeaction'
import { showSystemProcessesAtom } from 'components/namespaceprocesstree'

interface viewItem {
    icon: JSX.Element /** drawer item icon */
    label: string /** drawer item label */
    path: string /** route path */
    type?: string /** type of namespace to show, if any */
}

const views: viewItem[] = [
    { icon: <HomeIcon />, label: "all namespaces", path: "/" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.user), label: "user", path: "/user", type: "user" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.pid), label: "PID", path: "/pid", type: "pid" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.cgroup), label: "cgroup", path: "/cgroup", type: "cgroup" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.ipc), label: "IPC", path: "/ipc", type: "ipc" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.mnt), label: "mount", path: "/mnt", type: "mnt" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.net), label: "network", path: "/net", type: "net" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.uts), label: "UTS", path: "/uts", type: "uts" },
    { icon: CreateNamespaceTypeIcon(NamespaceType.time), label: "time", path: "/time", type: "time" },
    { icon: <InfoIcon />, label: "information", path: "/about" },
]

/**
 * The `LxknsApp` component renders the general app layout without thinking
 * about providers for routing, themes, discovery, et cetera. So this component
 * deals with:
 * - app bar with title, number of namespaces badge, quick actions.
 * - drawer for navigating the different views and types of namespaces.
 * - scrollable content area.
 */
const LxknsApp = () => {
    const { ErrorBoundary } = useErrorBoundary()

    const [treeaction, setTreeAction] = useTreeAction()

    const [showSystemProcesses, setShowSystemProcesses] = useAtom(showSystemProcessesAtom)

    const path = useLocation().pathname
    const typeview = views.find(view => view.path === path && view.type)

    const discovery = useDiscovery()

    return (
        <Box width="100vw" height="100vh" display="flex" flexDirection="column">
            <AppBarDrawer
                title={
                    <Badge badgeContent={Object.keys(discovery.namespaces).length} color="secondary">
                        Linux {typeview && `${typeview.type} `}Namespaces
                    </Badge>
                }
                tools={() => <>
                    <Tooltip title="expand initial user namespace(s) only">
                        <IconButton color="inherit"
                            onClick={() => setTreeAction(COLLAPSEALL)}>
                            <ChevronRightIcon />
                        </IconButton>
                    </Tooltip>
                    <Tooltip title="expand all">
                        <IconButton color="inherit"
                            onClick={() => setTreeAction(EXPANDALL)}>
                            <ExpandMoreIcon />
                        </IconButton>
                    </Tooltip>
                    <Refresher />
                </>}
                drawertitle={() => <>
                    <Typography variant="h6" style={{ flexGrow: 1 }} color="textSecondary" component="span">lxkns</Typography>
                    <Typography variant="body2" color="textSecondary" component="span">&#32;{version}</Typography>
                </>}
                drawer={closeDrawer => <>
                    <List onClick={closeDrawer}>
                        {views.map((viewitem, idx) =>
                            <DrawerLinkItem
                                key={idx}
                                icon={viewitem.icon}
                                label={viewitem.label}
                                path={viewitem.path}
                            />
                        )}
                    </List>
                    <Divider />
                    <List>
                        <ListItem>
                            <Toggle
                                checked={showSystemProcesses}
                                onChange={() => setShowSystemProcesses(!showSystemProcesses)}
                                color="primary"
                            />system processes
                        </ListItem>
                    </List>
                </>}
            />
            <Box m={1} flex={1} overflow="auto">
                <ErrorBoundary
                    render={() =>
                        <Switch>
                            <Route exact path="/about" render={() => <About />} />
                            {views.filter(viewitem => !!viewitem.type).map((viewitem, idx) =>
                                <Route
                                    exact path={viewitem.path}
                                    render={() => <NamespaceProcessTree type={viewitem.type} action={treeaction} />}
                                    key={idx}
                                />
                            )}
                            <Route path="/" render={() => <UserNamespaceTree action={treeaction} />} />
                        </Switch>
                    }
                    renderError={({ error }) => <pre>{error.toString()}</pre>}
                />
            </Box>
        </Box>)
}

// We need to wrap the application as otherwise we won't get a confirmer ...
// ouch. And since we're already at wrapping things, let's just wrap up all the
// other wrapping here... *snicker*.
const App = () => (
    <ThemeProvider theme={lxknsTheme}>
        <SnackbarProvider maxSnack={3}>
            <StateProvider>
                <Discovery />
                <Router>
                    <CssBaseline />
                    <LxknsApp />
                </Router>
            </StateProvider>
        </SnackbarProvider>
    </ThemeProvider>
)

export default App
