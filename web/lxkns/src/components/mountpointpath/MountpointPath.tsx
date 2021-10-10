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

import React from 'react'
import clsx from 'clsx'

import makeStyles from '@mui/styles/makeStyles';

import { MountPoint, unescapeMountPath } from 'models/lxkns/mount'
import HiddenmountIcon from 'icons/Hiddenmount'
import MountIcon from 'icons/namespaces/Mount'


const useStyles = makeStyles((theme) => ({
    mountpointpath: {
        display: 'inline-block',
        position: 'relative',
        '& .MuiSvgIcon-root': {
            verticalAlign: 'baseline',
            position: 'relative',
            top: '0.3ex',
            marginRight: '0.3em',
        }
    },
    hidden: {
        color: theme.palette.text.disabled,
        '& > span': {
            textDecoration: 'line-through solid',
        }
    },
}))


export interface MountpointPathProps {
    /** mount point with mount path. */
    mountpoint: MountPoint
    /** optionally show only "tail" path instead of full mount path. */
    tail?: string
    /** optionally change mount path name styling when hidden? */
    plainpath?: boolean
    /**
     * when to show a drum icon? When unspecified, defaults to "hidden", which
     * will render a drum icon only when the mountpoint is hidden.
     */
    drum?: 'never' | 'always' | 'hidden'
    /** 
     * when to show the mount point ID? When unspecified, defaults to "never".
     */
    showid?: 'never' | 'always' | 'hidden'
    /** optional CSS class(es). */
    className?: string
}

/**
 * Renders the path name of a mount point, with an (optional) additional
 * (broken) drum icon and crossed-out path name in case the mount point is
 * hidden.
 */
export const MountpointPath = ({
    mountpoint,
    tail,
    plainpath,
    drum,
    showid,
    className
}: MountpointPathProps) => {

    const classes = useStyles()

    drum = drum || 'hidden'

    const drumicon =
        (mountpoint.hidden && drum !== 'never' && <HiddenmountIcon fontSize="inherit" />)
        || (!mountpoint.hidden && drum !== 'never' && drum !== 'hidden' && <MountIcon fontSize="inherit" />)

    return <span className={clsx(
        className,
        classes.mountpointpath,
        mountpoint.hidden && !plainpath && classes.hidden)}
    >
        {drumicon}
        <span>{unescapeMountPath(tail || mountpoint.mountpoint)}</span>
        {(showid === 'always' || (showid === 'hidden' && mountpoint.hidden)) && ` (${mountpoint.mountid})`}
    </span>
}

export default MountpointPath
