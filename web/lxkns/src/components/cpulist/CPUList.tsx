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
    /* list of CPU ranges */
    cpus: number[][] | null
    /* show/hide a CPU icon before the CPU ranges */
    showIcon?: boolean
    /* allow line breaks after range (after the comma) */
    noWrap?: boolean
    /** optional tooltip override */
    tooltip?: string
    /** optional CSS class name(s). */
    className?: string
}

/**
 * The `CPUList` component renders a list of CPU ranges.
 */
export const CPUList = ({ cpus, showIcon, noWrap, tooltip, className }: CPUListProps) => {
    const sep = noWrap ? ',' : ',\u200b'
    tooltip = tooltip || 'CPU list'
    return !!cpus && (
        <Tooltip title={tooltip}>
            <CPURangeList className={clsx('cpulist', className)}>
                {!!showIcon && <CPUIcon fontSize="inherit" />}
                {
                    cpus.map((cpurange, index) => {
                        if (cpurange[0] === cpurange[1]) {
                            return <>{index > 0 && sep}{cpurange[0]}</>
                        }
                        return <>{index > 0 && sep}{cpurange[0]}–{cpurange[1]}</>
                    })
                }
            </CPURangeList>
        </Tooltip>
    )
}

export default CPUList
