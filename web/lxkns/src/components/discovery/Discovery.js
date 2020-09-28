import React, { createContext, useEffect, useState } from 'react';

import { postprocessDiscovery } from 'components/discovery/model';
import useInterval from 'hooks/interval';


const initialDiscoveryState = {
    namespaces: {},
    processes: {},
    previousNamespaces: {},
}

// DiscoveryContext provides information about the most recent namespaces
// discovery from the /api/namespaces endpoint.
export const DiscoveryContext = createContext(initialDiscoveryState);

// RefreshContext provides information about the refresh configuration and state
// of discovery.
export const RefreshContext = createContext({
    interval: null,
    refreshing: false,
    setInterval: (interval) => { },
    refresh: () => { },
});

// The Discovery component renders all its children, passing them two contexts:
// a DiscoveryContext, as well as a RefreshContext.
const Discovery = ({ children }) => {

    const [refresh, setRefresh] = useState({
        interval: 5000,
        refreshing: false,
        setInterval: null,
        refresh: () => {}
    });
    const [discovery, setDiscovery] = useState(initialDiscoveryState);

    refresh.setInterval = (interval) => {
        console.log("setting new interval", interval);
        setRefresh(interval);
    };

    // Fetch the namespace+process discovery data from the server, postprocess
    // the JSON result, and finally update the discovery data state with the new
    // information about all namespaces, adding information about the previous
    // discovery state.
    const fetchDiscoveryData = () => {
        fetch('/api/namespaces')
            .then(httpresult => httpresult.json())
            .then(jsondata => postprocessDiscovery(jsondata))
            .then(discovery => setDiscovery(prevDiscovery => {
                discovery.previousNamespaces = prevDiscovery.namespaces;
                return discovery;
            }));
    };

    // Get new discovery data after some time; please note that useInterval
    // interprets a null cycle as switching off the timer.
    useInterval(() => fetchDiscoveryData(), refresh.interval);

    // Initially fetch discovery data, unless the cycle is null.
    useEffect(() => {
        refresh.interval !== null && fetchDiscoveryData()
    }, [refresh.interval]);

    return (
        <RefreshContext.Provider value={refresh}>
            <DiscoveryContext.Provider value={discovery}>
                {children}
            </DiscoveryContext.Provider>
        </RefreshContext.Provider>
    );
};

export default Discovery;
