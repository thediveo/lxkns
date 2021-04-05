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
import { CypressHistorySupport } from 'cypress-react-router'

import { SnackbarProvider } from 'notistack'

import { Provider as StateProvider, useAtom } from 'jotai'

import CssBaseline from '@material-ui/core/CssBaseline'
import Badge from '@material-ui/core/Badge'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import Tooltip from '@material-ui/core/Tooltip'
import List from '@material-ui/core/List'
import { Box, createMuiTheme, Divider, fade, makeStyles, Theme, ThemeProvider, useMediaQuery, useTheme } from '@material-ui/core'

import SettingsIcon from '@material-ui/icons/Settings'
import HelpIcon from '@material-ui/icons/Help'
import HomeIcon from '@material-ui/icons/Home'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'
import ChevronRightIcon from '@material-ui/icons/ChevronRight'
import InfoIcon from '@material-ui/icons/Info'

import Discovery, { useDiscovery } from 'components/discovery'
import Refresher from 'components/refresher'
import AppBarDrawer, { DrawerLinkItem } from 'components/appbardrawer'
import { NamespaceType } from 'models/lxkns'

import { useTreeAction, EXPANDALL, COLLAPSEALL } from './treeaction'
import { lxknsDarkTheme, lxknsLightTheme } from './appstyles'
import { Settings, themeAtom, THEME_DARK, THEME_USERPREF } from 'views/settings'
import { NamespaceIcon } from 'components/namespaceicon'
import { About } from 'views/about'
import { Help } from 'views/help'
import { AllNamespaces } from 'views/allnamespaces'
import { TypedNamespaces } from 'views/typednamespaces'

interface viewItem {
    icon: JSX.Element /** drawer item icon */
    label: string /** drawer item label */
    path: string /** route path */
    type?: NamespaceType /** type of namespace to show, if any */
}

/**
 * Side drawer items, organized into groups which will later be visually
 * separated by dividers.
 */
const views: viewItem[][] = [
    [
        { icon: <HomeIcon />, label: "all namespaces", path: "/" },
    ], [
        {
            icon: <NamespaceIcon type={NamespaceType.user} />,
            label: "user namespaces", path: "/user", type: NamespaceType.user
        },
        {
            icon: <NamespaceIcon type={NamespaceType.pid} />,
            label: "PID namespaces", path: "/pid", type: NamespaceType.pid
        },
        {
            icon: <NamespaceIcon type={NamespaceType.cgroup} />,
            label: "cgroup namespaces", path: "/cgroup", type: NamespaceType.cgroup
        },
        {
            icon: <NamespaceIcon type={NamespaceType.ipc} />,
            label: "IPC namespaces", path: "/ipc", type: NamespaceType.ipc
        },
        {
            icon: <NamespaceIcon type={NamespaceType.mnt} />,
            label: "mount namespaces", path: "/mnt", type: NamespaceType.mnt
        },
        {
            icon: <NamespaceIcon type={NamespaceType.net} />,
            label: "network namespaces", path: "/net", type: NamespaceType.net
        },
        {
            icon: <NamespaceIcon type={NamespaceType.uts} />,
            label: "UTS namespaces", path: "/uts", type: NamespaceType.uts
        },
        {
            icon: <NamespaceIcon type={NamespaceType.time} />,
            label: "time namespaces", path: "/time", type: NamespaceType.time
        },
    ], [
        { icon: <SettingsIcon />, label: "settings", path: "/settings" },
        { icon: <HelpIcon />, label: "help", path: "/help/lxkns" },
        { icon: <InfoIcon />, label: "about", path: "/about" },
    ]
]


const themedFade = (theme: Theme, el: ('dark' | 'light'), f: number) => (
    theme.palette.type === 'light'
        ? fade(theme.palette.primary[el], f)
        : fade(theme.palette.primary[el], 1 - f)
)

const useStyles = makeStyles((theme) => ({
    drawer: {
        '& .MuiListItem-root.Mui-selected, & .MuiListItem-root.Mui-selected:hover': {
            backgroundColor: themedFade(theme, 'dark', 0.2),
        },
        '& .MuiListItem-root:hover': {
            backgroundColor: themedFade(theme, 'dark', 0.05),
        },
        '& .MuiListItemIcon-root .MuiSvgIcon-root': {
            color: theme.palette.primary.light,
        }
    }
}))

/**
 * The `LxknsApp` component renders the general app layout without thinking
 * about providers for routing, themes, discovery, et cetera. So this component
 * deals with:
 * - app bar with title, number of namespaces badge, quick actions.
 * - drawer for navigating the different views and types of namespaces.
 * - scrollable content area.
 */
const LxknsApp = () => {

    const classes = useStyles()
    const theme = useTheme()

    const [treeaction, setTreeAction] = useTreeAction()

    const path = useLocation().pathname

    // Note: JS returns undefined if the result doesn't turn up a match; that's
    // what we want ... and millions of Gophers are starting to cry (again).
    const [typeview] = views
        .flat()
        .filter(view => view.path === path && view.type)

    const discovery = useDiscovery()

    // Number of namespaces shown ... either type-specific or total number.
    const count = typeview
        ? Object.values(discovery.namespaces)
            .filter(netns => netns.type === typeview.type)
            .length
        : Object.keys(discovery.namespaces).length

    return (
        <Box width="100vw" height="100vh" display="flex" flexDirection="column">
            <AppBarDrawer
                drawerwidth={300}
                swipeAreaWidth={theme.spacing(1)}
                drawerClassName={classes.drawer}
                title={<>
                    <Badge badgeContent={count} color="secondary">
                        <Typography variant="h6">Linux {typeview && <em>{typeview.type} </em>}Namespaces</Typography>
                    </Badge>
                </>}
                tools={() => <>
                    <Tooltip key="collapseall" title="expand only top-level namespace(s)">
                        <IconButton color="inherit"
                            onClick={() => setTreeAction(COLLAPSEALL)}>
                            <ChevronRightIcon />
                        </IconButton>
                    </Tooltip>
                    <Tooltip key="expandall" title="expand all">
                        <IconButton color="inherit"
                            onClick={() => setTreeAction(EXPANDALL)}>
                            <ExpandMoreIcon />
                        </IconButton>
                    </Tooltip>
                    <Refresher />
                </>}
                drawertitle={() =>
                    <Typography variant="h6" style={{ flexGrow: 1 }} color="textSecondary" component="span">
                        lxkns
                    </Typography>
                }
                drawer={closeDrawer =>
                    views.map((group, groupidx) => [
                        groupidx > 0 && <Divider key={`div-${groupidx}`} />,
                        <List onClick={closeDrawer} key={groupidx}>
                            {group.map((viewitem, idx) =>
                                <DrawerLinkItem
                                    key={`${groupidx}-${idx}`}
                                    icon={viewitem.icon}
                                    label={viewitem.label}
                                    path={viewitem.path}
                                />
                            )}
                        </List>
                    ])
                }
            />
            <Box m={0} flex={1} overflow="auto">
                <Switch>
                    <Route exact path="/settings"><Settings /></Route>
                    <Route exact path="/about"><About /></Route>
                    <Route path="/help"><Help /></Route>
                    <Route
                        exact path={
                            views.map(group => group.filter(viewitem => !!viewitem.type))
                                .flat().map(viewitem => viewitem.path)}
                    >
                        <TypedNamespaces discovery={discovery} action={treeaction} />
                    </Route>
                    <Route path="/"><AllNamespaces discovery={discovery} action={treeaction} /></Route>
                </Switch>
            </Box>
        </Box>)
}

// Wrap the Lxkns app component into a theme provider that switches between
// light and dark themes depending on theme type configuration state.
const ThemedApp = () => {

    const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')
    const [theme] = useAtom(themeAtom)
    const themeType = theme === THEME_USERPREF
        ? (prefersDarkMode ? 'dark' : 'light')
        : (theme === THEME_DARK ? 'dark' : 'light')

    const appTheme = React.useMemo(() => createMuiTheme(
        {
            palette: {
                type: themeType,
            },
        },
        themeType === 'dark' ? lxknsDarkTheme : lxknsLightTheme,
    ), [themeType])

    return (
        <ThemeProvider theme={appTheme}>
            <CssBaseline />
            <SnackbarProvider maxSnack={3}>
                <Discovery />
                    <LxknsApp />
            </SnackbarProvider>
        </ThemeProvider>
    )
}

// Finally, the exported App component wraps the themed app component into a
// Jotai state provider, to keep state provision and app theme switching
// separated. And we also place the router high up here, so we can get the
// history object in the ThemedApp for passing it to Cypress, if present.
const App = () => (
    <StateProvider>
        <Router>
            <CypressHistorySupport />
            <ThemedApp />
        </Router>
    </StateProvider>
)

export default App
