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

import { useEffect } from 'react'

import { useAtom } from 'jotai'

import { useSnackbar } from 'notistack'

import { useInterval } from 'usehooks-ts'
import { discoveryErrorAtom, discoveryRefreshingAtom, discoveryRefreshIntervalAtom, refreshIntervalKey } from './hooks'

/**
 * The `Discovery` component is needed in order to toast errors to the snack bar
 * as well as refresh discovery information according to the refresh interval
 * set.
 */
const Discovery = () => {
    // In order to report discovery REST API failures... 
    const { enqueueSnackbar } = useSnackbar()

    // Discovery status and control...
    const [discoveryError] = useAtom(discoveryErrorAtom)
    const [interval] = useAtom(discoveryRefreshIntervalAtom)
    const [, setDiscoveryRefreshing] = useAtom(discoveryRefreshingAtom)

    // Get new discovery data after some time; please note that useInterval
    // interprets a null cycle as switching off the timer.
    useInterval(() => setDiscoveryRefreshing(true), interval)

    // Initially fetch discovery data, unless the cycle is null.
    useEffect(() => {
        if (interval !== null) {
            setDiscoveryRefreshing(true)
        }
        localStorage.setItem(refreshIntervalKey, JSON.stringify(interval))
    }, [interval, setDiscoveryRefreshing])

    useEffect(
        () => { 
            // eslint-disable-next-line @typescript-eslint/no-unused-expressions
            discoveryError && enqueueSnackbar(discoveryError, { variant: 'error' }) 
        },
        [discoveryError, enqueueSnackbar])

    // Do not render anything.
    return null
}

export default Discovery
