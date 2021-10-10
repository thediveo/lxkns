// Adapted from: https://material-ui.com/components/app-bar/#elevate-app-bar,
// minus the iframe-related elements not necessary in this context.

import React from 'react';
import useScrollTrigger from '@mui/material/useScrollTrigger';
import PropTypes from 'prop-types';

const ElevationScroll = (props) => {
    const { children } = props;

    const trigger = useScrollTrigger({
        disableHysteresis: true,
        threshold: 0
    });

    return React.cloneElement(children, {
        elevation: trigger ? 4 : 0,
    });
}

export default ElevationScroll;

ElevationScroll.propTypes = {
    children: PropTypes.element.isRequired,
};
