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
import { BrowserRouter as Router, Route, Routes, useLocation } from 'react-router-dom'
import { CypressHistorySupport } from 'cypress-react-router'

import { SnackbarProvider } from 'notistack'

import { Provider as StateProvider, useAtom } from 'jotai'

import CssBaseline from '@mui/material/CssBaseline'
import Badge from '@mui/material/Badge'
import Typography from '@mui/material/Typography'
import IconButton from '@mui/material/IconButton'
import Tooltip from '@mui/material/Tooltip'
import List from '@mui/material/List'
import {
    Box,
    createTheme,
    Divider,
    alpha,
    Theme,
    ThemeProvider,
    StyledEngineProvider,
    useMediaQuery,
    useTheme,
    styled,
} from '@mui/material'

import TuneIcon from '@mui/icons-material/Tune'
import HelpIcon from '@mui/icons-material/Help'
import HomeIcon from '@mui/icons-material/Home'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import InfoIcon from '@mui/icons-material/Info'

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

import { basename } from 'utils/basename'


/**
 * Describes properties of an individual sidebar view item, such as its icon to
 * show, label, and the route path it applies to.
 */
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
        { icon: <TuneIcon />, label: "settings", path: "/settings" },
        { icon: <HelpIcon />, label: "help", path: "/help/lxkns" },
        { icon: <InfoIcon />, label: "about", path: "/about" },
    ]
]

const themedFade = (theme: Theme, el: ('dark' | 'light'), f: number) => (
    theme.palette.mode === 'light'
        ? alpha(theme.palette.primary[el], f)
        : alpha(theme.palette.primary[el], 1 - f)
)

const LxknsAppBarDrawer = styled(AppBarDrawer)(({ theme }) => ({
    '& .MuiListItem-root.Mui-selected, & .MuiListItem-root.Mui-selected:hover': {
        backgroundColor: themedFade(theme, 'dark', 0.2),
    },
    '& .MuiListItem-root:hover': {
        backgroundColor: themedFade(theme, 'dark', 0.05),
    },
    '& .MuiListItemIcon-root .MuiSvgIcon-root': {
        color: theme.palette.primary.light,
    },
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
            <LxknsAppBarDrawer
                drawerwidth={300}
                swipeAreaWidth={Number(theme.spacing(1))}
                title={<>
                    <Badge badgeContent={count} color="secondary">
                        <Typography variant="h6">Linux {typeview && <em>{typeview.type} </em>}Namespaces</Typography>
                    </Badge>
                </>}
                tools={() => <>
                    <Tooltip key="collapseall" title="expand only top-level namespace(s)">
                        <IconButton color="inherit" onClick={() => setTreeAction(COLLAPSEALL)} size="large">
                            <ChevronRightIcon />
                        </IconButton>
                    </Tooltip>
                    <Tooltip key="expandall" title="expand all">
                        <IconButton color="inherit" onClick={() => setTreeAction(EXPANDALL)} size="large">
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
                <Routes>
                    <Route path="/settings" element={<Settings />} />
                    <Route path="/about" element={<About />} />
                    <Route path="/help/*" element={<Help />} />
                    {views.map(group => group.filter(viewitem => !!viewitem.type))
                        .flat().map(viewitem =>
                            <Route
                                key={viewitem.path}
                                path={viewitem.path}
                                element={<TypedNamespaces discovery={discovery} action={treeaction} />} />
                        )}
                    <Route path="/" element={<AllNamespaces discovery={discovery} action={treeaction} />} />
                </Routes>
            </Box>
        </Box>
    );
}

// Wrap the Lxkns app component into a theme provider that switches between
// light and dark themes depending on theme type configuration state.
const ThemedApp = () => {
    const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')
    const [theme] = useAtom(themeAtom)
    const themeMode = theme === THEME_USERPREF
        ? (prefersDarkMode ? 'dark' : 'light')
        : (theme === THEME_DARK ? 'dark' : 'light')

    const appTheme = React.useMemo(() => createTheme(
        {
            components: {
                MuiSelect: {
                    defaultProps: {
                        variant: 'standard', // MUI v4 default.
                    },
                },
            },
            palette: {
                mode: themeMode,
                primary: {
                    main: '#3f51b5',
                },
                secondary: {
                    main: '#f50057',
                },
            },
        },
        themeMode === 'dark' ? lxknsDarkTheme : lxknsLightTheme,
    ), [themeMode])

    return (
        <StyledEngineProvider injectFirst>
            <ThemeProvider theme={appTheme}>
                <CssBaseline />
                <SnackbarProvider maxSnack={3}>
                    <Discovery />
                    <LxknsApp />
                </SnackbarProvider>
            </ThemeProvider>
        </StyledEngineProvider>
    )
}

// Finally, the exported App component wraps the themed app component into a
// Jotai state provider, to keep state provision and app theme switching
// separated. And we also place the router high up here, so we can get the
// history object in the ThemedApp for passing it to Cypress, if present.
const App = () => (
    <StateProvider>
        <Router basename={basename}>
            <CypressHistorySupport />
            <ThemedApp />
        </Router>
    </StateProvider>
)

export default App
