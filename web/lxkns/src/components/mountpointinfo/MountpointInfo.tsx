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

import { compareMountPeers, MountPoint } from 'models/lxkns/mount'
import { IconButton, makeStyles } from '@material-ui/core'
import MountIcon from 'icons/namespaces/Mount'
import HiddenmountIcon from 'icons/Hiddenmount'
import { filesystemTypeLink } from './fslinks'
import MenuBookIcon from '@material-ui/icons/MenuBook'
import { NamespaceBadge } from 'components/namespacebadge'
import ProcessInfo from 'components/processinfo'


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
        alignSelf: 'baseline',
        lineHeight: theme.typography.body1.lineHeight,
    },
    propvalue: {
        gridColumn: '2/3',
        fontWeight: theme.typography.fontWeightLight,
        alignSelf: 'baseline',
        lineHeight: theme.typography.body1.lineHeight,
        '& .MuiIconButton-root': {
            verticalAlign: 'bottom',
        },
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
    },
    extdoc: {
        position: 'relative',
        width: 0,
        height: 0,
        '& > *': {
            position: 'absolute',
            top: 'calc(-50% + 0.2ex)',
        }
    },
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
        .map((opt, idx) => <>
            {idx > 0 && <>,<br /></>}
            {opt}
        </>)
    }</>


/**
 * Renders a list of (peer/master/slave) mount points, grouping them by mount
 * namespace and then sorting and listing them by path per mount namespace. The
 * grouping mount namespaces are sorted by their identifiers, that is, by their
 * inode numbers.
 */
const GroupedPropagationMembers = (members: MountPoint[]) => {
    // We use an object as our map (or dictionary): indexed by mount namespace
    // identifier we then map to the list of mount points belonging to that
    // particular mount namespace. As for the code: reduce() to the rescue,
    // which gives us a nice and compact way to iterate over all mount points
    // and building the index at the same time.
    const grouped = members.reduce((m, mountpoint) => ({
        ...m,
        [mountpoint.mountnamespace.nsid]: m[mountpoint.mountnamespace.nsid]
            ? m[mountpoint.mountnamespace.nsid].concat(mountpoint)
            : [mountpoint]
    }), {} as { [nsid: string]: MountPoint[] })
    // Given the map we can now render the grouping mount namespace badges with
    // short process info, as well as the per-mount namespace sorted list of
    // (peer/master/slave) mount points.
    return Object.values(grouped)
        .sort((group1, group2) => group1[0].mountnamespace.nsid - group2[0].mountnamespace.nsid)
        .map(group => {
            const mountns = group[0].mountnamespace
            return <div key={mountns.nsid}>
                <NamespaceBadge namespace={mountns} />
                &nbsp;<ProcessInfo short process={mountns.ealdorman} />
                <ul style={{ margin: 0 }}>{/* FIXME: proper styling */}
                    {group
                        .sort(compareMountPeers)
                        .map(peermountpoint =>
                            <li key={peermountpoint.mountpoint}>{/* FIXME: proper styling */}
                                {peermountpoint.hidden ? '(hidden) ' : ''}
                                {peermountpoint.mountpoint}
                            </li>)
                    }
                </ul>
            </div>
        })
}

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

    const parent = <>
        {mountpoint.parentid}
        {mountpoint.parent && <> ~ {mountpoint.parent.mountpoint}</>}
    </>

    // The mount point propagation peergroup actually does not only contain
    // peers but also slaves. Here, we want to only see true peers that aren't
    // slaves. And especially we don't want to see ourself.
    const peers = mountpoint.peergroup && mountpoint.peergroup.members
        .filter(member => member !== mountpoint && member.peergroup === mountpoint.peergroup)

    // The peergroup acting as our master(s) again not only contains masters
    // (=true peers), but also us and other slaves. So we need to filter us and
    // the other slaves out, keeping only the master (peers).
    const masters = mountpoint.mastergroup && mountpoint.mastergroup.members
        .filter(member => member !== mountpoint && member.peergroup === mountpoint.mastergroup)

    // And finally for the slaves: these are those members of our peergroup
    // which are not true peers.
    const slaves = mountpoint.peergroup && mountpoint.peergroup.members
        .filter(member => member !== mountpoint && member.mastergroup === mountpoint.peergroup)

    return <>
        <div className={classes.mountpath}>
            {mountpoint.hidden ? <HiddenmountIcon fontSize="inherit" /> : <MountIcon fontSize="inherit" />}
            &nbsp;{mountpoint.mountpoint}
        </div>
        <div className={classes.props}>
            <NameValueRow
                name={'mount namespace'}
                value={<>
                    <NamespaceBadge namespace={mountpoint.mountnamespace} /> <ProcessInfo short process={mountpoint.mountnamespace.ealdorman} />
                </>}
            />
            <NameValueRow name="device" value={`${mountpoint.major}:${mountpoint.minor}`} />
            <NameValueRow name="filesystem type" value={<>
                {mountpoint.fstype}
                &nbsp;<span className={classes.extdoc}>
                    <IconButton
                        size="small"
                        aria-label="external documentation"
                        href={filesystemTypeLink(mountpoint.fstype)}
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        <MenuBookIcon />
                    </IconButton>
                </span>
            </>}
            />
            <NameValueRow name="root" value={mountpoint.root} />{/* TODO: detect namespaces, render using badge */}
            <NameValueRow name="options" value={options} />
            <NameValueRow name="superblock options" value={<Options options={mountpoint.superoptions.split(',')} />} />
            <NameValueRow name="source" value={mountpoint.source} />
            {mountpoint.tags['unbindable'] && <NameValueRow name="propagation type" value="unbindable" />}
            {peers && peers.length > 0 && <NameValueRow
                name="peer mounts"
                value={GroupedPropagationMembers(peers)}
            />}
            {masters && masters.length > 0 && <NameValueRow
                name="master peer mounts"
                value={GroupedPropagationMembers(masters)}
            />}
            {slaves && slaves.length > 0 && <NameValueRow
                name="slave mounts"
                value={GroupedPropagationMembers(slaves)}
            />}
            <NameValueRow name="ID" value={mountpoint.mountid} />
            <NameValueRow name="parent ID" value={parent} />
            <NameValueRow name="tags" value={tags} />
        </div>
    </>
}

export default MountpointInfo
