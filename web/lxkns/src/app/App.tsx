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

/**
 * Side drawer items, organized into groups which will later be visually
 * separated by dividers.
 */
const views: viewItem[][] = [
    [
        { icon: <HomeIcon />, label: "all namespaces", path: "/" },
    ], [
        { icon: CreateNamespaceTypeIcon(NamespaceType.user), label: "user namespaces", path: "/user", type: "user" },
        { icon: CreateNamespaceTypeIcon(NamespaceType.pid), label: "PID namespaces", path: "/pid", type: "pid" },
        { icon: CreateNamespaceTypeIcon(NamespaceType.cgroup), label: "cgroup namespaces", path: "/cgroup", type: "cgroup" },
        { icon: CreateNamespaceTypeIcon(NamespaceType.ipc), label: "IPC namespaces", path: "/ipc", type: "ipc" },
        { icon: CreateNamespaceTypeIcon(NamespaceType.mnt), label: "mount namespaces", path: "/mnt", type: "mnt" },
        { icon: CreateNamespaceTypeIcon(NamespaceType.net), label: "network namespaces", path: "/net", type: "net" },
        { icon: CreateNamespaceTypeIcon(NamespaceType.uts), label: "UTS namespaces", path: "/uts", type: "uts" },
        { icon: CreateNamespaceTypeIcon(NamespaceType.time), label: "time namespaces", path: "/time", type: "time" },
    ], [
        { icon: <InfoIcon />, label: "information", path: "/about" },
    ]
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

    // Note: JS returns undefined if the result doesn't turn up a match; that's
    // what we want ... and millions of Gophers are starting to cry (again).
    const [typeview] = views.filter(group => group.some(view => view.path === path && view.type)).flat()

    const discovery = useDiscovery()

    return (
        <Box width="100vw" height="100vh" display="flex" flexDirection="column">
            <AppBarDrawer
                drawerWidth={300}
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
                    {views.map((group, groupidx) => <>
                        {groupidx > 0 && <Divider/>}
                        <List onClick={closeDrawer}>
                            {group.map((viewitem, idx) =>
                                    <DrawerLinkItem
                                        key={groupidx*100+idx}
                                        icon={viewitem.icon}
                                        label={viewitem.label}
                                        path={viewitem.path}
                                    />
                            )}
                        </List>
                    </>)}
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
                            {views.map(group => group.filter(viewitem => !!viewitem.type).map((viewitem, idx) =>
                                <Route
                                    exact path={viewitem.path}
                                    render={() => <NamespaceProcessTree type={viewitem.type} action={treeaction} />}
                                    key={idx}
                                />
                            )).flat()}
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
