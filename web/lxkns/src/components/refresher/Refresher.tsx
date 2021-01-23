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

import { CircularProgress, Fade, IconButton, makeStyles, Menu, MenuItem, Tooltip } from '@material-ui/core'
import RefreshIcon from '@material-ui/icons/Refresh'
import SyncIcon from '@material-ui/icons/Sync'
import SyncDisabledIcon from '@material-ui/icons/SyncDisabled'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'

import { discoveryRefreshingAtom, discoveryRefreshIntervalAtom } from 'components/discovery'
import useId from 'hooks/id'


const defaultThrobberThreshold = 500 /* ms */

export interface RefresherInterval {
    /** 
     * interval label to show; if undefined then a suitable label will be
     * derived from the interval number.
     */
    label?: string
    /** interval in milliseconds. */
    interval: number
}

const defaultIntervals = [
    { interval: null },
    { interval: 500 },
    { interval: 1000 },
    { interval: 5 * 1000 },
    { interval: 10 * 1000 },
    { interval: 30 * 1000 },
    { interval: 60 * 1000 },
    { interval: 5 * 60 * 1000 },
] as RefresherInterval[]

/**
 * Converts an interval number in milliseconds into a suitable textual label,
 * such as 500ms, 30s, et cetera. An interval value of null is taken as
 * "off".
 *
 * @param interval milliseconds
 */
const intervalToLabel = (interval: number) => {
    if (interval === null) {
        return "off"
    }
    const ms = interval % 1000
    const t = ms ? [`${ms}ms`] : []
    interval = Math.floor(interval / 1000)
    const sec = interval % 60
    if (sec) {
        t.unshift(`${sec}s`)
    }
    const min = Math.floor(interval / 60)
    if (min) {
        t.unshift(`${min}min`)
    }
    return t.join(' ')
}

// Progress indicator appearing around the refresh button.
const useStyles = makeStyles((theme) => ({
    wrapper: {
        margin: theme.spacing(1),
        position: 'relative',
    },
    discoveryprogress: {
        color: theme.palette.secondary.main,
        position: 'absolute',
        top: 8,
        left: 8,
        zIndex: 1,
    },
    interval: {
        // Unfortunately, Material-UI's 50% border radius results in an ugly
        // oblong oval-like shape, but we want proper 50% of height radii. See
        // https://stackoverflow.com/a/29966500 for the rescue by setting an
        // incredibly large border radius in pixels which then triggers a
        // dedicated "50% of the smaller axis" rule. Something we want! 
        borderRadius: '999px',
    }
}))

export interface RefresherProps {
    /** 
     * show throbber if refresh takes longer than the specified threshold;
     * defaults to 500ms. 
     */
    throbberThreshold?: number
    /**
     * an array of refresh intervals; if left undefined, then a default array is
     * applied.
     */
    intervals?: RefresherInterval[]
}

/**
 * A refresher that doesn't stink. This component gives users control over the
 * interval between refreshes, as well as a chance to fire off single-shot
 * on-demand refreshes. Users can switch off automatic refreshing completely. If
 * a refresh takes more than a certain threshold (defaults to 500ms), then a
 * rotating progress indicator appears around the refresh button.
 *
 * This component actually renders two buttons:
 * - on-demand refresh button,
 * - refresh interval selector button, which shows an interval selection menu
 *   when pressed (clicked, touched, ...).
 *
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
const Refresher = ({ throbberThreshold, intervals }: RefresherProps) => {
    const classes = useStyles()
    const menuId = useId('refreshermenu')

    // Refresh interval and status (is a refresh ongoing?).
    const [refreshInterval, setRefreshInterval] = useAtom(discoveryRefreshIntervalAtom)
    const [refreshing, setRefreshing] = useAtom(discoveryRefreshingAtom)

    // Used for popping up the interval menu.
    const [anchorEl, setAnchorEl] = useState(null)

    // Create the final list of interval values and labels, based on what we
    // were given, or rather, no given.
    intervals = [...(intervals || defaultIntervals)]
        .map(i => ({
            interval: i.interval,
            label: i.label || intervalToLabel(i.interval)
        } as RefresherInterval))

    // User clicks on the auto-refresh button to pop up the associated menu.
    const handleIntervalButtonClick = (event) => {
        setAnchorEl(event.currentTarget)
    };

    // User selects an auto-refresh interval menu item.
    const handleIntervalMenuChange = (interval: RefresherInterval) => {
        setAnchorEl(null)
        console.log("setting auto-refresh to:", interval.label)
        setRefreshInterval(interval.interval)
    };

    // User clicks outside the popped up interval menu.
    const handleIntervalMenuClose = () => setAnchorEl(null);

    const intervalTitle = refreshInterval !== null
        ? "auto-refresh interval " + intervalToLabel(refreshInterval)
        : "auto-refresh off"

    return (
        <>
            <Tooltip title="refresh">
                <div className={classes.wrapper}>
                    <IconButton color="inherit"
                        onClick={() => setRefreshing(true)}
                    ><RefreshIcon /></IconButton>
                    {refreshing &&
                        <Fade
                            in={true}
                            style={{ transitionDelay: `${throbberThreshold || defaultThrobberThreshold}ms` }}
                            unmountOnExit
                        >
                            <CircularProgress size={32} className={classes.discoveryprogress} />
                        </Fade>
                    }
                </div>
            </Tooltip>
            <Tooltip title={intervalTitle}>
                <IconButton
                    className={classes.interval}
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
                {intervals.map(i => (
                    <MenuItem
                        key={i.interval}
                        value={i.interval}
                        selected={i.interval === refreshInterval}
                        onClick={() => handleIntervalMenuChange(i)}
                    >
                        {i.label}
                    </MenuItem>
                ))}
            </Menu>
        </>
    )
}

export default Refresher
