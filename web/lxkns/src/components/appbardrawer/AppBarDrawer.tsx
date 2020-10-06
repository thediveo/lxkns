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

import { makeStyles, useTheme } from '@material-ui/core'
import AppBar from '@material-ui/core/AppBar'
import Toolbar from '@material-ui/core/Toolbar'
import SwipeableDrawer from '@material-ui/core/SwipeableDrawer'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import MenuIcon from '@material-ui/icons/Menu'
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft'
import ChevronRightIcon from '@material-ui/icons/ChevronRight'
import Divider from '@material-ui/core/Divider'

// Width of drawer.
const drawerWidth = 240

// We need some styling for the AppBar in order to correctly flex the title so
// it takes up the available space and pushes the app bar tools to the right.
const useStyles = makeStyles((theme) => ({
    menuButton: { marginRight: theme.spacing(2) },
    drawer: { width: drawerWidth, flexShrink: 0 },
    drawerHeader: {
        display: 'flex',
        alignItems: 'center',
        padding: theme.spacing(0, 1),
        ...theme.mixins.toolbar, // necessary for content to be below app bar
        justifyContent: 'flex-end',
    },
    drawerPaper: { width: drawerWidth },
    title: { flexGrow: 1 }
}))

/**
 * Callback function to call when the drawer needs to be closed.
 */
type drawerCloser = () => void

export interface AppBarDrawerProps {
    /** the app title to show in the app bar. */
    title: React.ReactNode
    /** optional tools (icon buttons, et cetera) to show in the tool bar. */
    tools?: () => React.ReactNode
    /**
     * a function rendering the contents inside the drawer. This function gets
     * passed a callback function so that components inside the drawer are
     * able to close the drawer when necessary. For instance, links typically
     * want to close the drawer whenever the user clicks on them in order to
     * navigate to a different route.
     */
    drawer?: (drawerCloser: drawerCloser) => React.ReactNode
}

/**
 * AppBarDrawer is a high-order component ("hoc" in react parlance) that
 * covers the gory details of giving an application an app bar with tools, as
 * well as a drawer for navigation, et cetera.
 */
const AppBarDrawer = ({ title, tools, drawer }: AppBarDrawerProps) => {

    // Not much state here in ... Denmark?!
    const [drawerOpen, setDrawerOpen] = useState(false)

    // Convenience handlers for dealing with the swipeable drawer, that should
    // keep users busy on a rainy Sunday afternoon.
    const openDrawer = () => { setDrawerOpen(true) }
    const closeDrawer = () => { setDrawerOpen(false) }
    const toggleDrawer = () => { setDrawerOpen(!drawerOpen) }

    const theme = useTheme()
    const classes = useStyles()

    return (<>
        <AppBar position="static">
            <Toolbar>
                <IconButton
                    edge="start"
                    className={classes.menuButton}
                    color="inherit"
                    aria-label="menu"
                    onClick={toggleDrawer}
                >
                    <MenuIcon />
                </IconButton>

                <Typography variant="h6" className={classes.title}>
                    {title}
                </Typography>

                {tools()}
            </Toolbar>
        </AppBar>
        <SwipeableDrawer
            className={classes.drawer}
            classes={{ paper: classes.drawerPaper }}
            open={drawerOpen}
            onOpen={openDrawer}
            onClose={closeDrawer}
        >
            <div className={classes.drawerHeader}>
                <IconButton onClick={closeDrawer}>
                    {theme.direction === 'ltr' ? <ChevronLeftIcon /> : <ChevronRightIcon />}
                </IconButton>
            </div>
            <Divider />
            {drawer && drawer(closeDrawer)}
        </SwipeableDrawer>
    </>)
}

export default AppBarDrawer
