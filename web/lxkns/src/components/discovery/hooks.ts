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

import { atom, useAtom, type Setter } from "jotai"
import { fromjson, type Discovery } from "models/lxkns"
import { localStorageAtom } from "utils/persistentsettings"

/** Internal discovery result state; can be used only via useDiscovery(). */
const discoveryResultAtom = atom({
    namespaces: {},
    processes: {},
} as Discovery)

/** 
 * Use the namespace discovery result in a react component; on purpose, there no
 * way to set it (it wouldn't make sense).
 */
export const useDiscovery = () => {
    const [discovery] = useAtom(discoveryResultAtom)
    return discovery
}

/**
 * Internal discovery error state; internally used to display a snackbar
 * message via the Discovery component. 
 */
export const discoveryErrorAtom = atom("")

export const refreshIntervalKey = "lxkns.refresh.interval"

const initialRefreshInterval = (() => {
    try {
        const interval = JSON.parse(localStorage.getItem(refreshIntervalKey) || "");
        if (interval === null || (Number.isInteger(interval) && interval > 500)) {
            return interval
        }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    } catch (e) { /* empty */ }
    return 5000;
})()

/** 
 * The discovery refresh interval state; null means refresh is disabled. This
 * state is automatically synced to the local storage.
 */
export const discoveryRefreshIntervalAtom = localStorageAtom(refreshIntervalKey, initialRefreshInterval)

/** 
 * Discovery refresh status; setting the status to "true" triggers an ad-hoc
 * refresh, unless there is already a refresh ongoing. It is not possible to
 * reset an ongoing refresh.
 */
export const discoveryRefreshingAtom = atom(
    false,
    (get, set, arg) => {
        const refreshing = get(discoveryRefreshingAtom)
        if (arg as boolean && !refreshing) {
            set(discoveryRefreshingAtom, true)
            fetchDiscoveryData(set)
        }
    }
)

// Fetch the namespace+process discovery data from the server, postprocess
// the JSON result, and finally update the discovery data state with the new
// information about all namespaces, adding information about the previous
// discovery state.
const fetchDiscoveryData = (set: Setter) => {
    fetch('api/namespaces') // relative to base href!
        .then(httpresult => {
            // Whatever the server replied, it did reply and we can reset
            // the refreshing indication. 
            set(discoveryRefreshingAtom, false)
            // fetch() doesn't throw an error for non-2xx reponse status codes, so
            // we throw up here instead, so we can catch below later in the promise
            // chain.
            if (!httpresult.ok) {
                throw Error(httpresult.status + " " + httpresult.statusText)
            }
            try {
                return httpresult.json()
            // eslint-disable-next-line @typescript-eslint/no-unused-vars
            } catch (e) {
                throw Error('malformed discovery API response')
            }
        })
        .then(jsondata => fromjson(jsondata))
        .then(discovery => set(discoveryResultAtom, discovery))
        .catch((error) => {
            // Don't forget to reset the refreshing indication and then set the
            // error result, so someone else can pick it up and send a toast to the
            // snackbar. Before 10pm. And only less than six toasts. Just for
            // testing eyesight.
            set(discoveryRefreshingAtom, false)
            set(discoveryErrorAtom,
                'refreshing failed: ' + error.toString().replace(/^[E|e]rror: /, ''))
        })
}
