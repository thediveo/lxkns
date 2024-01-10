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

import { Container, Engine } from 'models/lxkns'
import { engineTypeName } from 'utils/engine'
import { ContainerTypeIcon } from 'components/containertypeicon'

const EngineInformation = styled('span')(() => ({
    display: 'inline-block',
    whiteSpace: 'nowrap',
    '& .MuiSvgIcon-root': {
        marginRight: '0.15em',
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.2ex',
    },
}))

const EngineID = styled('span')(() => ({
    fontVariantNumeric: 'tabular-nums',
    fontFamily: 'Roboto Mono',
    '&:before': {
        content: '"«"',
        fontStyle: 'normal',
    },
    '&:after': {
        content: '"»"',
        fontStyle: 'normal',
    },
}))

export interface EngineInfo {
    /** information about a container engine (with workload) */
    engine: Engine
    /** optional CSS class name(s). */
    className?: string
}

export const EngineInfo = ({ engine, className }: EngineInfo) => {
    const typename = engineTypeName(engine.type)

    return !!engine && (
        <EngineInformation className={clsx(className)}>
            <ContainerTypeIcon container={{
                type: engine.type,
                flavor: engine.type,
            } as Container}
                fontSize="inherit"
            />
            {typename} engine ({engine.pid})
            ID: <EngineID>{engine.id}</EngineID>
        </EngineInformation>
    )
}

export default EngineInfo
