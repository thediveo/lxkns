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

import clsx from 'clsx'

import { styled } from '@mui/material'

import { Process } from 'models/lxkns/model'
import CPUList from 'components/cpulist/CPUList'

const CPUAffSchedInformation = styled('span')(({ theme }) => ({
    marginLeft: '0.5em',
    color: theme.palette.cpulist,
    '&.affschedinformation > .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
        marginRight: 0,
        color: theme.palette.cpulist,
    },
    '& .policy': {
        fontSize: '80%',
    },
    '& .normal,& .batch,& .idle': {
        color: theme.palette.relaxedsched,
    },
    '& .fifo,& .rr': {
        color: theme.palette.stressedsched,
    },
    '& .nice': {
        color: theme.palette.nice,
    },
    '& .notnice': {
        color: theme.palette.notnice,
    },
    '& .prio': {
        color: theme.palette.prio,
    },
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
    const schedpol = schedulerPolicies[process.policy || 0]
    const prio = process.priority || 0
    return !!process.affinity && (
        <CPUAffSchedInformation className="affschedinformation">
            <CPUList cpus={process.affinity} showIcon noWrap tooltip="CPU affinity list" />
            {!!process.policy && <span className={clsx('policy', schedpol.toLowerCase())}>&nbsp;{schedpol}</span>}
            {hasPriority(process) && <span className={clsx(prio > 0 && 'prio')}>&nbsp;priority {prio}</span>}{
                hasNice(process) && !!process.nice &&
                <span className={process.nice >= 0 ? 'nice' : 'notnice'}>&nbsp;nice {process.nice}</span>}
        </CPUAffSchedInformation>
    )
}

export default AffinityScheduleInfo