import React, { useEffect, useState, useContext, useRef } from 'react';

import IconButton from '@material-ui/core/IconButton';
import Tooltip from '@material-ui/core/Tooltip';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import RefreshIcon from '@material-ui/icons/Refresh';
import SyncIcon from '@material-ui/icons/Sync';
import SyncDisabledIcon from '@material-ui/icons/SyncDisabled';

import ExpandMoreIcon from '@material-ui/icons/ExpandMore';

import { RefreshContext } from 'components/discovery';

const intervals = [
    { text: 'off', interval: null },
    { text: '1s', interval: 1000 },
    { text: '5s', interval: 5 * 1000 },
    { text: '10s', interval: 10 * 1000 },
    { text: '30s', interval: 30 * 1000 },
    { text: '1m', interval: 60 * 1000 },
    { text: '5m', interval: 5 * 60 * 1000 },
];

// returns the textual description for a specific interval.
const intervaltext = interval => intervals.find(intervalitem => interval === intervalitem.interval).text;

const Refresher = () => {

    // Used for popping up the interval menu.
    const [anchorEl, setAnchorEl] = useState(null);

    // User clicks on the auto-refresh button to pop up the associated menu.
    const handleIntervalButtonClick = (event) => {
        setAnchorEl(event.currentTarget);
    };

    // User selects an auto-refresh interval menu item.
    const handleIntervalMenuItemClick = (event, interval) => {
        setAnchorEl(null);
        console.log("setting auto-refresh to ", interval ? (interval / 1000) + "s" : "off");
        // FIXME: setCycle(interval);
    };

    // User clicks outside the popped up interval menu.
    const handleIntervalMenuClose = () => setAnchorEl(null);

    return (
        <RefreshContext.Consumer>
            {refresh => {
                const intervalTitle = refresh.interval !== null ? "auto-refresh interval " + intervaltext(refresh.interval) : "auto-refresh off";
                return (<>
                    <Tooltip title="refresh">
                        <IconButton color="inherit" onClick={() => { }/* FIXME: */}><RefreshIcon /></IconButton>
                    </Tooltip>
                    <Tooltip title={intervalTitle}>
                        <IconButton
                            aria-haspopup="true"
                            aria-controls="cyclesmenu"
                            onClick={handleIntervalButtonClick}
                            color="inherit"
                        >
                            {refresh.interval !== null ? <SyncIcon /> : <SyncDisabledIcon />}
                            <ExpandMoreIcon />
                        </IconButton>
                    </Tooltip>
                    <Menu
                        id="cyclesmenu"
                        anchorEl={anchorEl}
                        keepMounted
                        open={Boolean(anchorEl)}
                        onClose={handleIntervalMenuClose}
                    >
                        {intervals.map((intervalitem, index) => (
                            <MenuItem
                                key={intervalitem.interval}
                                selected={intervalitem.interval === refresh.interval}
                                onClick={(event) => handleIntervalMenuItemClick(event, intervalitem.interval)}
                            >
                                {intervalitem.text}
                            </MenuItem>
                        ))}
                    </Menu>
                </>);
            }}
        </RefreshContext.Consumer>
    );
};

export default Refresher;
