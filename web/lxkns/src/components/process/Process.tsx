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

import React from 'react'

import Tooltip from '@material-ui/core/Tooltip';

import RunFast from 'mdi-material-ui/RunFast';
import CarCruiseControl from 'mdi-material-ui/CarCruiseControl';

import { Process } from 'components/lxkns'

export interface ProcessInfoProps {
    /** information about a discovered Linux OS process. */
    process: Process
}

/** 
 * The `ProcessInfo` component renders (limited) information about a single
 * Linux OS process:
 *
 * - name of the process, which is has been either set by the process itself,
 *   or has been derived from the process' command line.
 * - PID.
 * - cgroup path, if path is not empty.
 *
 * On purpose, this component doesn't render more comprehensive information
 * (such as parent and children, et cetera), as it is to be used in concise
 * contexts, such as a single process tree node.
 */
export const ProcessInfo = ({ process }: ProcessInfoProps) => (
    <span className="processinfo">
        {!!process && (<>
            <Tooltip title="process"><span>
                <RunFast fontSize="inherit" />&nbsp;
                <span className="processname">"{process.name}"</span> ({process.pid})
            </span></Tooltip>
            {process.cgroup && (
                <Tooltip title="control-group path" className="cgroupinfo">
                    <span>
                        <CarCruiseControl fontSize="inherit" />&nbsp;
                            "<span className="cgroupname">{process.cgroup}</span>"
                        </span>
                </Tooltip>)}
        </>)}
    </span>
)

export default ProcessInfo
