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

import React, { useState } from 'react'

import MenuIcon from '@mui/icons-material/Menu'
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import { AppBar, Box, Divider, IconButton, styled, SwipeableDrawer, Theme, Toolbar, useTheme } from '@mui/material'


// Width of drawer.
const defaultDrawerWidth = 240

// We need some styling for the AppBar in order to correctly flex the title so
// it takes up the available space and pushes the app bar tools to the right.
// The drawer width can be parameterized; for this we need to define the
// properties getting passed later to styled() 
interface SwipeableDrawerStyleProps {
    theme?: Theme,
    /** width of app drawer in pixels */
    drawerwidth?: number,
}

const ToolbarActionButton = styled(IconButton)(({ theme }) => ({
    marginRight: theme.spacing(2),
}))

const DrawerHeader = styled('div')(({ theme }) => ({
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    padding: theme.spacing(0, 1),
    ...theme.mixins.toolbar, // necessary for content to be below app bar
    justifyContent: 'flex-end',
}))

const SwappyDrawer = styled(SwipeableDrawer)<SwipeableDrawerStyleProps>(({ theme, drawerwidth }: SwipeableDrawerStyleProps) => ({
    width: drawerwidth || defaultDrawerWidth,
    flexShrink: 0,

    '& .MuiDrawer-paper': {
        width: drawerwidth || defaultDrawerWidth,
    },

    '& .MuiListSubheader-root': {
        background: theme?.palette.background.paper,
    },
}))

/**
 * Callback function to call when the drawer needs to be closed.
 */
type drawerCloser = () => void

export interface AppBarDrawerProps {
    /** app title in the app bar. */
    title: React.ReactNode | (() => React.ReactNode)
    /** 
     * optional tools (icon buttons, et cetera) to place in the tool bar,
     * aligned to the end (right) of the app bar.
     */
    tools?: React.ReactNode | (() => React.ReactNode)
    /** 
     * app title in drawer (as opposed to the app bar title). This can be
     * arbitrary content, such as the app title and version (see also example).
     */
    drawertitle?: React.ReactNode | (() => React.ReactNode)
    /**
     * a function rendering the contents inside the drawer. This function gets
     * passed a callback function so that components inside the drawer are
     * able to close the drawer when necessary. For instance, links typically
     * want to close the drawer whenever the user clicks on them in order to
     * navigate to a different route.
     */
    drawer?: (drawerCloser: drawerCloser) => React.ReactNode
    /**
     * optionally sets the width in pixels of the drawer. Defaults to 240 pixels
     * if unspecified.
     */
    drawerwidth?: number
    /** CSS style class name(s) for drawer. */
    drawerClassName?: string
    /** touch area width in pixels for swiping the drawer open. */
    swipeAreaWidth?: number
}

/**
 * `AppBarDrawer` provides an application bar ("app bar") with a navigation
 * drawer. The app bar comes with title and optional action buttons.
 *
 * The navigation drawer can be opened by swiping from the left side or by
 * clicking/tapping on the drawer icon (☰) to the left of the app bar. It can be
 * closed either by swiping to the left or clicking on the close (<) button in
 * the drawer. The drawer close button is automatically added. The navigation
 * drawer takes arbitrary content, yet you typically will want to fill it with
 * [`DrawerLinkItem`](#DrawerLinkItem)s.
 *
 * Please note that the `drawer=` property expects a function rendering the
 * drawer contents on request; it gets passed a `closeDrawer` handler argument
 * which should called as an event handler to close the drawer when clicking on
 * navigation buttons, et cetera. Please see the example for usage.
 *
 * When using
 * [IconButton](https://material-ui.com/api/icon-button/#iconbutton-api) as app
 * bar action buttons don't forget to set `color="inherit"` on the icon button:
 * the icons then will take on the appropriate appbar foreground color (usually
 * as opposed to the default primary color).
 * 
 * **IMPORTANT:** `AppBarDrawer` must be directly or indirectly enclosed inside
 * a [`(Browser)Router`](https://reactrouter.com/web/api/BrowserRouter)
 * component.
 *
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
const AppBarDrawer = ({
    title,
    tools,
    drawertitle,
    drawer,
    drawerwidth,
    drawerClassName,
    swipeAreaWidth,
}: AppBarDrawerProps) => {

    // Not much state here in ... Denmark?!
    const [drawerOpen, setDrawerOpen] = useState(false)

    // Convenience handlers for dealing with the swipeable drawer, that should
    // keep users busy on a rainy Sunday afternoon.
    const openDrawer = () => { setDrawerOpen(true) }
    const closeDrawer = () => { setDrawerOpen(false) }
    const toggleDrawer = () => { setDrawerOpen(!drawerOpen) }

    const theme = useTheme()

    return <>
        <AppBar position="static" elevation={0}>
            <Toolbar>
                <>
                    <ToolbarActionButton
                        edge="start"
                        color="inherit"
                        aria-label="menu"
                        onClick={toggleDrawer}
                        size="large">
                        <MenuIcon />
                    </ToolbarActionButton>

                    {title && ((typeof title === 'function' && title()) || title)}

                    <Box component="span" sx={{ flexGrow: 1 }} />

                    {tools && ((typeof tools === 'function' && tools()) || tools)}
                </>
            </Toolbar>
        </AppBar>
        <SwappyDrawer
            className={drawerClassName}
            swipeAreaWidth={swipeAreaWidth || 20}
            drawerwidth={drawerwidth}
            open={drawerOpen}
            onOpen={openDrawer}
            onClose={closeDrawer}
        >
            <DrawerHeader>
                {drawertitle &&
                    <Box component="span" sx={{ flexGrow: 1 }}>{typeof drawertitle === 'function' ? drawertitle() : drawertitle}</Box>}
                <IconButton onClick={closeDrawer} size="large">
                    {theme.direction === 'ltr' ? <ChevronLeftIcon /> : <ChevronRightIcon />}
                </IconButton>
            </DrawerHeader>
            <Divider />
            {drawer && drawer(closeDrawer)}
        </SwappyDrawer>
    </>
}

export default AppBarDrawer
