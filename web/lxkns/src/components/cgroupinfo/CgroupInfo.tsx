// Copyright 2022 Harald Albrecht.
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

import { Pause, PlayArrow } from '@mui/icons-material'
import { styled, Tooltip } from '@mui/material'
import { Busybody } from 'models/lxkns'
import CgroupNamespace from 'icons/namespaces/Cgroup'
import clsx from 'clsx'


const CgroupInformation = styled('span')(({ theme }) => ({
    marginLeft: '0.5em',
    '&.cgroupinformation > .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
        marginRight: 0,
        color: theme.palette.cgroup,
    },
    '& > .MuiSvgIcon-root.frozen': {
        color: theme.palette.freezer.frozen,
    },
    '& > .MuiSvgIcon-root.running': {
        color: theme.palette.freezer.run,
    },
}))

const CgroupIcon = styled(CgroupNamespace)(({ theme }) => ({
    '&.MuiSvgIcon-root': {
        color: theme.palette.cgroup,
    }
}))

const CgroupPath = styled('span')(({ theme }) => ({
    color: theme.palette.cgroup,
    '&::before': { content: '"«"' },
    '&::after': { content: '"»"' },
}))


export interface CgroupInfoProps {
    /** process or task. */
    busybody: Busybody
    /** optional CSS class name(s). */
    className?: string
}

/**
 * The `CgroupInfo` component renders information about a process's or task's
 * control group.
 */
export const CgroupInfo = ({ busybody, className }: CgroupInfoProps) => {

    const fridge = busybody.fridgefrozen
        ? <Pause className="frozen" fontSize="inherit" />
        : <PlayArrow className="running" fontSize="inherit" />

    return <Tooltip title="control-group path">
        <CgroupInformation className={clsx('cgroupinformation', className)}>
            <CgroupIcon fontSize="inherit" />
            {fridge}
            <CgroupPath>{busybody.cpucgroup}</CgroupPath>
        </CgroupInformation>
    </Tooltip>
}

export default CgroupInfo
