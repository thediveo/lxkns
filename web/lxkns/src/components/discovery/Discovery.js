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
        triggerRefresh: null,
    });
    // Allow other components consuming the RefreshContext to change the
    // refreshing interval and to trigger refreshes on demand.
    refresh.setInterval = (interval) => {
        setRefresh({ ...refresh, interval: interval });
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
        setRefresh({ ...refresh, refreshing: true });
        fetch('/api/namespaces')
            .then(httpresult => {
                setRefresh({ ...refresh, refreshing: false });
                return httpresult;
            })
            .then(httpresult => httpresult.json())
            .then(jsondata => postprocessDiscovery(jsondata))
            .then(discovery => setDiscovery(prevDiscovery => {
                discovery.previousNamespaces = prevDiscovery.namespaces;
                return discovery;
            }))
            .catch(() => setRefresh({ ...refresh, refreshing: false }));
    };

    // Get new discovery data after some time; please note that useInterval
    // interprets a null cycle as switching off the timer.
    useInterval(() => fetchDiscoveryData(), refresh.interval);

    // Initially fetch discovery data, unless the cycle is null.
    useEffect(() => {
        if (refresh.interval !== null) {
            fetchDiscoveryData();
        }
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
