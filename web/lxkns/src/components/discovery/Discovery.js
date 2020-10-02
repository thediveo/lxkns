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

import React, { createContext, useEffect, useState } from 'react';

import { useSnackbar } from 'notistack';

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

const localStorageKey = "lxkns.refresh.interval";

const initialInterval = (() => {
    try {
        const interval = JSON.parse(localStorage.getItem(localStorageKey));
        if (interval === null || (Number.isInteger(interval) && interval > 500)) {
            return interval
        }
    } catch (e) { }
    return 5000;
})();

// The Discovery component renders all its children, passing them two contexts:
// a DiscoveryContext, as well as a RefreshContext.
const Discovery = ({ children }) => {
    const { enqueueSnackbar } = useSnackbar();

    const [refresh, setRefresh] = useState({
        interval: initialInterval,
        refreshing: false,
    });
    // Allow other components consuming the RefreshContext to change the
    // refreshing interval and to trigger refreshes on demand.
    refresh.setInterval = interval => {
        setRefresh(prevRefresh => {return { ...prevRefresh, interval: interval }});
    };
    refresh.triggerRefresh = () => {
        fetchDiscoveryData();
    };

    // The discovery state to share to consumers of the DiscoveryContext.
    const [discovery, setDiscovery] = useState(initialDiscoveryState);

    // Fetch the namespace+process discovery data from the server, postprocess
    // the JSON result, and finally update the discovery data state with the new
    // information about all namespaces, adding information about the previous
    // discovery state.
    const fetchDiscoveryData = () => {
        if (refresh.refreshing) {
            return;
        }
        setRefresh(prevRefresh => {return { ...refresh, refreshing: true }});
        fetch('/api/namespaces')
            .then(httpresult => {
                // Whatever the server replied, it did reply and we can reset
                // the refreshing indication. 
                setRefresh(prevRefresh => {return { ...refresh, refreshing: false }});
                // fetch() doesn't throw an error for non-2xx reponse status
                // codes...
                if (!httpresult.ok) {
                    console.log(httpresult);
                    throw Error(httpresult.status + " " + httpresult.statusText);
                }
                try {
                    return httpresult.json()
                } catch (e) {
                    throw Error('malformed discovery API response');
                }
            })
            .then(jsondata => postprocessDiscovery(jsondata))
            .then(discovery => setDiscovery(prevDiscovery => {
                discovery.previousNamespaces = prevDiscovery.namespaces;
                return discovery;
            }))
            .catch((error) => {
                // Don't forget to reset the refreshing indication.
                setRefresh(prevRefresh => {return { ...refresh, refreshing: false }})
                enqueueSnackbar('refreshing failed: ' +
                    error.toString().replace(/^[E|e]rror: /, ''), { variant: 'error' });
            });
    };

    // Get new discovery data after some time; please note that useInterval
    // interprets a null cycle as switching off the timer.
    useInterval(() => fetchDiscoveryData(), refresh.interval);

    // Initially fetch discovery data, unless the cycle is null.
    useEffect(() => {
        if (refresh.interval !== null) {
            fetchDiscoveryData();
        }
        localStorage.setItem(localStorageKey, JSON.stringify(refresh.interval));
        // Trust me, I know what I'm doing...
        //
        // eslint-disable-next-line react-hooks/exhaustive-deps
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
