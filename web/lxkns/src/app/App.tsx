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

import React, { useRef, useState } from 'react'
import { BrowserRouter as Router, Route, Routes, useLocation, Navigate } from 'react-router-dom'

import { SnackbarProvider, useSnackbar } from 'notistack'

import { Provider as StateProvider, useAtom } from 'jotai'

import CssBaseline from '@mui/material/CssBaseline'
import Typography from '@mui/material/Typography'
import IconButton from '@mui/material/IconButton'
import Tooltip from '@mui/material/Tooltip'
import List from '@mui/material/List'
import {
    Box,
    createTheme,
    Divider,
    alpha,
    type Theme,
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

import { lxknsDarkTheme, lxknsLightTheme } from 'styles/themes'
import { Settings, themeAtom, THEME_DARK, THEME_USERPREF } from 'views/settings'
import { NamespaceIcon } from 'components/namespaceicon'
import { About } from 'views/about'
import { Help } from 'views/help'
import { AllNamespaces } from 'views/allnamespaces'
import { TypedNamespaces } from 'views/typednamespaces'
import Logo from 'app/lxkns.svg'

import { basename } from 'utils/basename'
import ContainerIcon from 'icons/containers/Container'
import { Containers } from 'views/containers'
import type { TreeAPI } from './treeapi'
import CPUAffinityIcon from 'icons/CPUAffinity'
import { Affinities } from 'views/affinities'
import { CondBadge } from 'components/condbadge'
import DownloadIcon from 'icons/Download'
import UploadIcon from 'icons/Upload'
import { discoveryRefreshIntervalAtom, useRawDiscoveryJSON } from 'components/discovery/hooks'
import { generateFilename } from 'utils/generatefilename'
import { DiscoveryUploader } from 'components/discoveryuploader/DiscoveryUploader'
import { rgba } from 'utils/rgba'
import BlockIcon from '@mui/icons-material/Block'

interface tooltips {
    collapseall?: string
    expandall?: string
}

interface individualTreeactions {
    collapseall?: boolean
    expandall?: boolean
}

const isIndividualTreeActions = (i: boolean | individualTreeactions): i is individualTreeactions =>
    typeof i === 'object'

// treeactionEnabled returns true if either the passed configuration value is
// true, or an object where the additionally specified field/key has the value
// true; otherwise, false is returned.
const treeactionEnabled = (enableTreeActions: boolean | individualTreeactions, action: keyof individualTreeactions) =>
    isIndividualTreeActions(enableTreeActions) ? Boolean(enableTreeActions[action]) : enableTreeActions

/**
 * Describes properties of an individual sidebar view item, such as its icon to
 * show, label, and the route path it applies to.
 */
interface viewItem {
    icon: React.JSX.Element /** drawer item icon */
    label: string /** drawer item label text */
    path: string /** route path */

    title: string /** title */

    type?: NamespaceType /** type of namespace to show, if any */

    badge?: boolean /** show namespaces count badge */
    saveload?: boolean /** show download/upload buttons */
    treeactions?: boolean | individualTreeactions /** show tree expand/collapse buttons */
    tooltips?: tooltips /** tooltip text if differing from default */
}

/**
 * Side drawer items, organized into groups which will later be visually
 * separated by dividers.
 */
const views: viewItem[][] = [
    [
        {
            icon: <HomeIcon />, label: "all namespaces", path: "/",
            title: "All Linux Namespaces",
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <ContainerIcon />, label: "all containers", path: "/containers",
            title: "All Container Namespaces",
            badge: true, saveload: true, treeactions: true
        },
    ], [
        {
            icon: <NamespaceIcon type={NamespaceType.user} />,
            title: "Linux User Namespaces",
            label: "user namespaces", path: "/user", type: NamespaceType.user,
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <NamespaceIcon type={NamespaceType.pid} />,
            title: "Linux PID Namespaces",
            label: "PID namespaces", path: "/pid", type: NamespaceType.pid,
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <NamespaceIcon type={NamespaceType.cgroup} />,
            title: "Linux Cgroup Namespaces",
            label: "cgroup namespaces", path: "/cgroup", type: NamespaceType.cgroup,
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <NamespaceIcon type={NamespaceType.ipc} />,
            title: "Linux IPC Namespaces",
            label: "IPC namespaces", path: "/ipc", type: NamespaceType.ipc,
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <NamespaceIcon type={NamespaceType.mnt} />,
            title: "Linux Mnt Namespaces",
            label: "mount namespaces", path: "/mnt", type: NamespaceType.mnt,
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <NamespaceIcon type={NamespaceType.net} />,
            title: "Linux Net Namespaces",
            label: "network namespaces", path: "/net", type: NamespaceType.net,
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <NamespaceIcon type={NamespaceType.uts} />,
            title: "Linux UTS Namespaces",
            label: "UTS namespaces", path: "/uts", type: NamespaceType.uts,
            badge: true, saveload: true, treeactions: true
        },
        {
            icon: <NamespaceIcon type={NamespaceType.time} />,
            title: "Linux Time Namespaces",
            label: "time namespaces", path: "/time", type: NamespaceType.time,
            badge: true, saveload: true, treeactions: true
        },
    ], [
        {
            icon: <CPUAffinityIcon />, label: "core fancy", path: "/affinities",
            title: "Core Fancy",
            saveload: true,
            treeactions: {
                collapseall: true,
            },
            tooltips: {
                collapseall: "collapse to show only CPU nodes with PID1 and PID2",
            }
        },
    ], [
        { icon: <TuneIcon />, label: "settings", path: "/settings", title: "Settings" },
        { icon: <HelpIcon />, label: "help", path: "/help", title: "Help" },
        { icon: <InfoIcon />, label: "about", path: "/about", title: "About lxkns" },
    ]
]

const viewProperties = (location: string) => (
    (views.flat().filter((vi) => location.startsWith(vi.path)) // grab only matching paths
        .sort((a, b) => b.path.length - a.path.length))[0] // then sort with longest match first
)

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

const NoDropZone = styled(Box)(({ theme }) => ({
    zIndex: theme.zIndex.modal - 1, // this allows drag&drop into a modal dialog
    position: 'absolute',
    inset: 0, // attach to top-left and bottom-right.
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: rgba(theme.palette.background.paper, 0.5),
    pointerEvents: 'none',
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
    const { enqueueSnackbar } = useSnackbar()

    const [showUploaderDialog, setShowUploaderDialog] = useState(false)

    const path = useLocation().pathname
    const view = viewProperties(path)
    const typeview = view?.type ? view : undefined

    const discovery = useDiscovery()
    const hasDiscovered = Object.keys(discovery.processes).length !== 0
    const [rawJSON, setRawJSON] = useRawDiscoveryJSON()
    const [, setRefreshInterval] = useAtom(discoveryRefreshIntervalAtom)

    const [inDrag, setInDrag] = useState(false)

    // Number of namespaces shown ... either type-specific or total number.
    const count = typeview
        ? Object.values(discovery.namespaces)
            .filter(netns => netns.type === typeview.type)
            .length
        : Object.keys(discovery.namespaces).length

    // keep the various tree API instances, so we can route the collapse/expand
    // all button actions to the tree in the currently selected view.
    const forrestRef = useRef(new Map<string, TreeAPI | null>())

    // calls the passed fn with the currently visible tree API instance, if any.
    // Otherwise, does nothing.
    const currentAPI = (fn: (api: TreeAPI) => void) => {
        const api = forrestRef.current?.get(basename + path)
        if (api) {
            fn(api)
        }
    }

    // get the discovery raw JSON and attach it to a short-lived document link
    // that we then automatically click in order to initiate the download of the
    // raw JSON.
    const handleDownload = () => {
        const blob = new Blob([rawJSON], { type: 'application/json' })
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = generateFilename('lxkns', 'json')
        link.click()
        enqueueSnackbar('successfully downloaded discovery data', {
            variant: 'success',
            autoHideDuration: 2000,
        })
        document.removeChild(link)
    }

    // we're asked to import the passed discovery raw JSON, so we disable any
    // automatic refresh and set the discovery raw JSON. This then automatically
    // triggers JSON parsing and the generation of the fully interconnected
    // in-RAM discovery model.
    const handleImport = (content: string) => {
        setRefreshInterval(null)
        setRawJSON(content)
    }

    // when dragging something over the app window (excluding a modal dialog) we
    // simply block dropping, preventing the browser from switching away from
    // the app.
    const handleDragOver = (event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault()
        setInDrag(true)
    }

    // when leaving dragging, ensure to remove the "no drag" overlay.
    const handleDragLeave = (event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault()
        setInDrag(false)
    }

    // when dropping something over the app window (excluding a modal dialog) we
    // simply block dropping, preventing the browser from switching away from
    // the app.
    const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
        event.preventDefault()
        setInDrag(false)
    }

    return (
        <Box
            width="100vw" height="100vh"
            display="flex" flexDirection="column"
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
        >
            <NoDropZone sx={{ display: inDrag ? undefined : 'none' }}>
                <BlockIcon sx={{
                    fontSize: "500%",
                    color: theme.palette.error.main
                }} />
            </NoDropZone>
            <LxknsAppBarDrawer
                drawerwidth={300}
                swipeAreaWidth={Number(theme.spacing(1))}
                title={
                    <CondBadge show={view.badge || false} badgeContent={count} color="secondary">
                        <Typography variant="h6">{view.title}</Typography>
                    </CondBadge>
                }
                tools={() => <>
                    {view.treeactions && <span>
                        {treeactionEnabled(view.treeactions, 'collapseall') &&
                            <Tooltip key="collapseall" title={view.tooltips?.collapseall || "expand only top-level namespace(s)"}>
                                <IconButton color="inherit" size="large" disabled={!hasDiscovered}
                                    onClick={() => {
                                        currentAPI((api) => api?.collapseAll())
                                    }}>
                                    <ChevronRightIcon />
                                </IconButton>
                            </Tooltip>}
                        {treeactionEnabled(view.treeactions, 'expandall') &&
                            <Tooltip key="expandall" title={view.tooltips?.expandall || "expand all"}>
                                <IconButton color="inherit" size="large" disabled={!hasDiscovered}
                                    onClick={() => {
                                        currentAPI((api) => api?.expandAll())
                                    }}>
                                    <ExpandMoreIcon />
                                </IconButton>
                            </Tooltip>}
                    </span>}
                    {view.saveload && <span>
                        <Tooltip title="Download discovery data">
                            <IconButton color="inherit" size="large" disabled={!hasDiscovered}
                                onClick={handleDownload}
                            >
                                <DownloadIcon />
                            </IconButton>
                        </Tooltip>
                        <Tooltip title="Import discovery data">
                            <IconButton color="inherit" size="large"
                                onClick={() => setShowUploaderDialog(true)}
                            >
                                <UploadIcon />
                            </IconButton>
                        </Tooltip>
                    </span>}
                    <Refresher />
                </>}
                drawertitle={() =>
                    <Typography variant="h6" style={{ flexGrow: 1 }} color="textSecondary" component="span">
                        <img alt="lxkns logo" src={Logo} style={{ height: '2ex', position: 'relative', top: '0.4ex' }} />&nbsp;lxkns
                    </Typography>
                }
                drawer={(closeDrawer: React.MouseEventHandler<HTMLUListElement> | undefined) =>
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
            <DiscoveryUploader
                open={showUploaderDialog}
                onClose={() => setShowUploaderDialog(false)}
                onImport={handleImport}
            />
            <Box m={0} flex={1} overflow="auto">
                <Routes>
                    <Route path="/settings" element={<Settings />} />
                    <Route path="/about" element={<About />} />
                    <Route path="/help" element={<Navigate to="/help/lxkns" replace />} />
                    <Route path="/help/*" element={<Help />} />
                    <Route
                        path="/containers"
                        element={<Containers
                            discovery={discovery}
                            apiRef={(apiref) => {
                                forrestRef.current?.set(basename + "/containers", apiref)
                                return () => { forrestRef.current?.delete(basename + "/containers") }
                            }}
                        />}
                    />
                    <Route
                        path="/affinities"
                        element={<Affinities
                            discovery={discovery}
                            apiRef={(apiref) => {
                                forrestRef.current?.set(basename + "/affinities", apiref)
                                return () => { forrestRef.current?.delete(basename + "/affinities") }
                            }}
                        />}
                    />
                    {views.map(group => group.filter(viewitem => !!viewitem.type))
                        .flat().map(viewitem => {
                            return <Route
                                key={viewitem.path}
                                path={viewitem.path}
                                element={<TypedNamespaces
                                    discovery={discovery}
                                    apiRef={(apiref) => {
                                        forrestRef.current?.set(basename + viewitem.path, apiref)
                                        return () => { forrestRef.current?.delete(basename + viewitem.path) }
                                    }}
                                />}
                            />
                        })}
                    <Route
                        path="/"
                        element={<AllNamespaces
                            discovery={discovery}
                            apiRef={(apiref) => {
                                forrestRef.current?.set(basename + "/", apiref)
                                return () => { forrestRef.current?.delete(basename + "/") }
                            }}
                        />}
                    />
                </Routes>
            </Box>
        </Box>
    )
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
// separated. And we also place the router high up here.
const App = () => (
    <StateProvider>
        <Router basename={basename}>
            <ThemedApp />
        </Router>
    </StateProvider>
)

export default App
