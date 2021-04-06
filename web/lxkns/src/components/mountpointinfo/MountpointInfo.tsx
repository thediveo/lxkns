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

import { MountPoint } from 'models/lxkns/mount'
import { makeStyles } from '@material-ui/core'
import MountIcon from 'icons/namespaces/Mount'
import HiddenmountIcon from 'icons/Hiddenmount'
import FilesystemtypeIcon from 'icons/Filesystemtype'
import { filesystemTypeLink } from './fslinks'
import { ExtLink } from 'components/extlink'


const useStyle = makeStyles((theme) => ({
    props: {
        display: 'grid',
        gridTemplateColumns: 'auto 1fr',
        columnGap: theme.spacing(2),
        rowGap: theme.spacing(1) / 2,
        '& .MuiSvgIcon-root': {
            verticalAlign: 'baseline',
            position: 'relative',
            top: '0.1ex',
        },
    },
    propname: {
        gridColumn: '1/2',
        whiteSpace: 'nowrap',
    },
    propvalue: {
        gridColumn: '2/3',
        fontWeight: theme.typography.fontWeightLight,
    },
    fullwidthpropvalue: {
        gridColumn: '1/3',
    },
    mountpath: {
        fontSize: '120%',
        fontWeight: theme.typography.fontWeightLight,
        marginBottom: theme.spacing(1),

        '& .MuiSvgIcon-root': {
            verticalAlign: 'baseline',
            position: 'relative',
            top: '0.2ex',
        },
    }
}))


interface NameValueRowProps {
    name: React.ReactNode
    value: React.ReactNode
}

/**
 * Renders a single key-value grid row.
 */
const NameValueRow = ({ name, value }: NameValueRowProps) => {

    const classes = useStyle()

    return <>
        <div className={classes.propname}>{name}:</div>
        <div className={classes.propvalue}>{value}</div>
    </>
}


const Options = ({ options }: { options: string[] }) =>
    <>{options
        .sort((opt1, opt2) => opt1.localeCompare(opt2, undefined, { numeric: true }))
        .map((opt, idx) => [
            idx > 0 && <>,<br /></>,
            <>{opt}</>
        ])
    }</>

export interface MountpointInfoProps {
    /** mount point information object. */
    mountpoint: MountPoint
}

/**
 * Renders detail information about a specific mount point.
 */
export const MountpointInfo = ({ mountpoint }: MountpointInfoProps) => {

    const classes = useStyle()

    const options = mountpoint.mountoptions
        .sort((opt1, opt2) => opt1.localeCompare(opt2, undefined, { numeric: true }))
        .map((opt, idx) => [
            idx > 0 && <>,<br /></>,
            <>{opt}</>
        ])

    const tags = Object.entries(mountpoint.tags)
        .sort(([tagname1,], [tagname2,]) => tagname1.localeCompare(tagname2, undefined, { numeric: true }))
        .map(([tagname, tagvalue], idx) => [
            idx > 0 && <br />,
            <>{tagname}: {tagvalue}</>
        ])

    return <>
        <div className={classes.mountpath}>
            {mountpoint.hidden ? <HiddenmountIcon fontSize="inherit" /> : <MountIcon fontSize="inherit" />}
            &nbsp;{mountpoint.mountpoint}
        </div>
        <div className={classes.props}>
            <NameValueRow name={"device"} value={`${mountpoint.major}:${mountpoint.minor}`} />
            <NameValueRow name={"filesystem type"} value={<>
                <FilesystemtypeIcon fontSize="inherit" />
                &nbsp;<ExtLink href={filesystemTypeLink(mountpoint.fstype)} iconposition="after">{mountpoint.fstype}</ExtLink>
            </>} />
            <NameValueRow name={"root"} value={mountpoint.root} />{/* TODO: detect namespaces, render using badge */}
            <NameValueRow name={"options"} value={options} />
            <NameValueRow name={"superblock options"} value={<Options options={mountpoint.superoptions.split(',')} />} />
            <NameValueRow name={"source"} value={mountpoint.source} />
            <NameValueRow name={"ID"} value={mountpoint.mountid} />
            <NameValueRow name={"parent ID"} value={mountpoint.parentid} />
            <NameValueRow name={"tags"} value={tags} />
        </div>
    </>
}

export default MountpointInfo
