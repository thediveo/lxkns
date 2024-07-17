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

import { styled, Tooltip } from '@mui/material'
import CPUIcon from '@mui/icons-material/Memory'

import { Process } from 'models/lxkns/model'

const CPURangeList = styled('span')(({ theme }) => ({
    color: theme.palette.cpulist,
    '&.cpulist > .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
        marginRight: '0.2em',
        color: theme.palette.cpulist,
    },
}))

export interface CPUListProps {
    /** information about a discovered Linux OS process. */
    process: Process
    showIcon?: boolean
    noWrap?: boolean
    /** optional tooltip override */
    tooltip?: string
    /** optional CSS class name(s). */
    className?: string
}

export const CPUList = ({ process, showIcon, noWrap, tooltip, className }: CPUListProps) => {
    const sep = noWrap ? ',' : ',\u200b'
    tooltip = tooltip || 'CPU list'
    return !!process.affinity && (
        <Tooltip title={tooltip}>
            <CPURangeList className={clsx('cpulist', className)}>
                {!!showIcon && <CPUIcon fontSize="inherit" />}
                {
                    process.affinity.map((cpurange, index) => {
                        if (cpurange[0] === cpurange[1]) {
                            return <>{sep && index > 0}{cpurange[0]}</>
                        }
                        return <>{sep && index > 0}{cpurange[0]}–{cpurange[1]}</>
                    })
                }
            </CPURangeList>
        </Tooltip>
    )
}

export default CPUList
