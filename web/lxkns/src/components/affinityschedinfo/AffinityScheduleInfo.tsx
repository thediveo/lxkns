// Copyright 2024 Harald Albrecht.
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

//import clsx from 'clsx'

import { styled } from '@mui/material'

import CPUIcon from '@mui/icons-material/Memory'
import { Process } from 'models/lxkns/model'

const CPUAffSchedInformation = styled('span')(({ theme }) => ({
    marginLeft: '0.5em',
    color: theme.palette.cgroup, // FIXME:
    '&.affschedinformation > .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
        marginRight: 0,
        color: theme.palette.cgroup, // FIXME:
    },

}))

const CPUAffinityIcon = styled(CPUIcon)(({ theme }) => ({
    '&.MuiSvgIcon-root': {
        color: theme.palette.cgroup, // FIXME:
    }
}))

const schedulerPolicies: { [key: string]: string } = {
    '0': 'NORMAL',
    '1': 'FIFO',
    '2': 'RR',
    '3': 'BATCH',
    '5': 'IDLE',
    '6': 'DEADLINE',
}

const hasPriority = (process: Process) => {
    const policy = process.policy || 0
    return policy === 1 || policy === 2
}

const hasNice = (process: Process) => {
    const policy = process.policy || 0
    return policy === 0 || policy === 3
}

export interface AffinityScheduleInfoProps {
    /** information about a discovered Linux OS process. */
    process: Process
}

export const AffinityScheduleInfo = ({ process }: AffinityScheduleInfoProps) => {
    return !!process.affinity && (
        <CPUAffSchedInformation className='affschedinformation'>
            <CPUAffinityIcon fontSize="inherit" />&nbsp;
            {
                process.affinity.map((cpurange, index) => {
                    if (cpurange[0] === cpurange[1]) {
                        return <>{"," && index > 0}{cpurange[0]}</>
                    }
                    return <>{"," && index > 0}{cpurange[0]}‒{cpurange[1]}</>
                })
            }
            {!!process.policy && <>&nbsp;{schedulerPolicies[process.policy]}</>}
            {hasPriority(process) && <>&nbsp;priority {process.priority || 0}</>}{
            hasNice(process) && !!process.nice && <>&nbsp;nice {process.nice}</>}
        </CPUAffSchedInformation>
    )
}

export default AffinityScheduleInfo