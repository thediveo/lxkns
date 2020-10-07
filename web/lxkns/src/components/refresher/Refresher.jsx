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

import { useAtom } from 'jotai'

import { makeStyles } from '@material-ui/core/styles'
import { green } from '@material-ui/core/colors'
import Fade from '@material-ui/core/Fade'
import CircularProgress from '@material-ui/core/CircularProgress'

import IconButton from '@material-ui/core/IconButton'
import Tooltip from '@material-ui/core/Tooltip'
import Menu from '@material-ui/core/Menu'
import MenuItem from '@material-ui/core/MenuItem'
import RefreshIcon from '@material-ui/icons/Refresh'
import SyncIcon from '@material-ui/icons/Sync'
import SyncDisabledIcon from '@material-ui/icons/SyncDisabled'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'

import { discoveryRefreshingAtom, discoveryRefreshIntervalAtom } from 'components/discovery'
import useId from 'hooks/id';

const intervals = [
    { text: 'off', interval: null },
    { text: '500ms', interval: 500 },
    { text: '1s', interval: 1000 },
    { text: '5s', interval: 5 * 1000 },
    { text: '10s', interval: 10 * 1000 },
    { text: '30s', interval: 30 * 1000 },
    { text: '1m', interval: 60 * 1000 },
    { text: '5m', interval: 5 * 60 * 1000 },
];

// returns the textual description for a specific interval.
const intervaltext = interval => intervals.find(intervalitem => interval === intervalitem.interval).text;

// Progress indicator appearing around the refresh button.
const useStyles = makeStyles((theme) => ({
    wrapper: {
        margin: theme.spacing(1),
        position: 'relative',
    },
    discoveryprogress: {
        color: green[500],
        position: 'absolute',
        top: 8,
        left: 8,
        zIndex: 1,
    }
}));

// The refresher component allows users to control the interval between
// refreshes, as well as single-shot on-demand refreshes. Users can also switch
// off automatic refreshing completely. If a refresh takes more than 800ms, then
// a rotating progress indicator appears around the refresh button.
const Refresher = () => {
    const classes = useStyles();
    const menuId = useId('refreshermenu');

    // Refresh interval and status.
    const [refreshInterval, setRefreshInterval] = useAtom(discoveryRefreshIntervalAtom)
    const [refreshing, setRefreshing] = useAtom(discoveryRefreshingAtom)

    // Used for popping up the interval menu.
    const [anchorEl, setAnchorEl] = useState(null);

    // User clicks on the auto-refresh button to pop up the associated menu.
    const handleIntervalButtonClick = (event) => {
        setAnchorEl(event.currentTarget);
    };

    // User selects an auto-refresh interval menu item.
    const handleIntervalMenuItemClick = (event, interval) => {
        setAnchorEl(null);
        console.log("setting auto-refresh to: ", interval ? (interval / 1000) + "s" : "off");
        setRefreshInterval(interval);
    };

    // User clicks outside the popped up interval menu.
    const handleIntervalMenuClose = () => setAnchorEl(null);

    const intervalTitle = refreshInterval !== null 
        ? "auto-refresh interval " + intervaltext(refreshInterval) 
        : "auto-refresh off";

    return (
        <>
            <Tooltip title="refresh">
                <div className={classes.wrapper}>
                    <IconButton color="inherit"
                        onClick={() => setRefreshing(true)}
                    ><RefreshIcon /></IconButton>
                    {refreshing &&
                        <Fade in={true} style={{ transitionDelay: '500ms' }} unmountOnExit>
                            <CircularProgress size={32} className={classes.discoveryprogress} />
                        </Fade>
                    }
                </div>
            </Tooltip>
            <Tooltip title={intervalTitle}>
                <IconButton
                    aria-haspopup="true"
                    aria-controls={menuId}
                    onClick={handleIntervalButtonClick}
                    color="inherit"
                >
                    {refreshInterval !== null ? <SyncIcon /> : <SyncDisabledIcon />}
                    <ExpandMoreIcon />
                </IconButton>
            </Tooltip>
            <Menu
                id={menuId}
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={handleIntervalMenuClose}
            >
                {intervals.map((intervalitem, ) => (
                    <MenuItem
                        key={intervalitem.interval}
                        selected={intervalitem.interval === refreshInterval}
                        onClick={(event) => handleIntervalMenuItemClick(event, intervalitem.interval)}
                    >
                        {intervalitem.text}
                    </MenuItem>
                ))}
            </Menu>
        </>
    );
};

export default Refresher;
