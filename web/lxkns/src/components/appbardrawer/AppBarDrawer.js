import React, { useState } from 'react';
import { useLocation, Link } from "react-router-dom";

import { makeStyles, useTheme } from '@material-ui/core';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import SwipeableDrawer from '@material-ui/core/SwipeableDrawer';
import Typography from '@material-ui/core/Typography';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import Divider from '@material-ui/core/Divider';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';

import ElevationScroll from 'components/elevationscroll';

const drawerWidth = 240;

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
}));

// AppBarDrawer is a high-order component ("hoc" in react parlance) that covers
// the gory details of giving an application an app bar with tools and a drawer
// for navigation, et cetera. The app bar stays at the top and gets elevated as
// soon as the user scrolls down even a iota.
//
// - title: AppBar title.
// - tools: items/tools to show in the app bar.
// - drawer: content of drawer.
//   - gets passed {closeDrawer()}.
//
const AppBarDrawer = ({ title, tools, drawer }) => {

    const [drawerOpen, setDrawerOpen] = useState(false);

    const closeDrawer = () => { setDrawerOpen(false) };
    const toggleDrawer = () => { setDrawerOpen(!drawerOpen) };

    const theme = useTheme();
    const classes = useStyles();

    return (<>
        <ElevationScroll>
            <AppBar>
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

                    {tools}
                </Toolbar>
            </AppBar>
        </ElevationScroll>
        <Toolbar />
        <SwipeableDrawer
            className={classes.drawer}
            classes={{ paper: classes.drawerPaper }}
            open={drawerOpen}
            onClose={closeDrawer}
        >
            <div className={classes.drawerHeader}>
                <IconButton onClick={closeDrawer}>
                    {theme.direction === 'ltr' ? <ChevronLeftIcon /> : <ChevronRightIcon />}
                </IconButton>
            </div>
            <Divider />
            {drawer(closeDrawer)}
        </SwipeableDrawer>
    </>);
};

export default AppBarDrawer;

export const DrawerLinkItem = ({ icon, label, path }) => {
    const location = useLocation();
    const selected = location.pathname === path;

    return (
        <ListItem
            button
            component={Link}
            to={path}
            selected={selected}
        >
            {icon && <ListItemIcon>{icon}</ListItemIcon>}
            <Typography>{label}</Typography>
        </ListItem>
    )
};
