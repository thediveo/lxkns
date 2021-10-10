// Copyright 2021 Harald Albrecht.
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

import clsx from 'clsx'

import makeStyles from '@mui/styles/makeStyles';
import { Pause } from '@mui/icons-material'

import { Container, containerGroup } from 'models/lxkns'
import { ContainerTypeIcon } from 'utils/containericon'

import ComposerProjectIcon from 'icons/containers/ComposerProject'
import PodIcon from 'icons/containers/Pod'
import IEAppIcon from 'icons/containers/IEApp'


const useStyles = makeStyles((theme) => ({
    // The whole component as such...
    containerInfo: {
        fontWeight: theme.typography.fontWeightLight,
        display: 'inline-block',
        whiteSpace: 'nowrap',
        '& .MuiSvgIcon-root': {
            marginRight: '0.15em',
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.2ex',
            color: [theme.palette.container, '!important'],
        },
    },
    containerName: {
        fontStyle: 'italic',
        color: theme.palette.container,
        '&::before': {
            content: '"«"',
            fontStyle: 'normal',
        },
        '&::after': {
            content: '"»"',
            fontStyle: 'normal',
            paddingLeft: '0.1em', // avoid italics overlapping with guillemet
        },
    },
    groupInfo: {
        marginLeft: '0.5em',
    },
    groupName: {
        color: theme.palette.container,
        '&::before': {
            content: '"«"',
            fontStyle: 'normal',
        },
        '&::after': {
            content: '"»"',
            fontStyle: 'normal',
            paddingLeft: '0.1em', // avoid italics overlapping with guillemet
        },
    },
}))

/**
 * The `ContainerInfo` component expects only a single property: the container to
 * render information about.
 */
export interface ContainerInfoProps {
    /** information about a discovered container. */
    container: Container
    /** optional CSS class name(s). */
    className?: string
}

/** 
 * The `ContainerInfo` component renders information about the name of a
 * container as well as an optional group.
 */
export const ContainerInfo = ({ container, className }: ContainerInfoProps) => {

    const classes = useStyles()

    const ContainerIcon = ContainerTypeIcon(container)

    var groupicon = null
    var groupname = ""
    const project = containerGroup(container, 'com.docker.compose.project')
    if (project) {
        groupname = project.name
        groupicon = container.flavor === 'com.siemens.industrialedge.app'
            ? <IEAppIcon fontSize="inherit" />
            : <ComposerProjectIcon fontSize="inherit" />
    }
    const pod = containerGroup(container, 'io.kubernetes.pod')
    if (pod) {
        groupname = pod.name
        groupicon = <PodIcon fontSize="inherit" />
    }

    const paused = container.paused && <Pause fontSize="inherit" />

    return !!container && (
        <span className={clsx(classes.containerInfo, className)}>
            <ContainerIcon fontSize="inherit" />
            {paused}
            <span className={classes.containerName}>
                {container.name}
            </span>
            {groupicon && <span className={classes.groupInfo}>
                in {groupicon}<span className={classes.groupName}>{groupname}</span>
            </span>}
        </span>
    )
}

export default ContainerInfo
