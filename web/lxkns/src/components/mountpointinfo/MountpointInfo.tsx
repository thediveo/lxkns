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
import { IconButton, styled, Tooltip } from '@mui/material'
import { filesystemTypeLink } from './fslinks'
import MenuBookIcon from '@mui/icons-material/MenuBook'
import { MountpointPath } from 'components/mountpointpath'
import { GroupedPropagationMembers } from 'components/groupedpropagationmembers/GroupedPropagationMembers'
import { NamespaceMap } from 'models/lxkns/model'
import { MountpointRoot } from 'components/mountpointroot'
import { NamespaceInfo } from 'components/namespaceinfo'


const MountPathTitle = styled('div')(({ theme }) => ({
    fontSize: '120%',
    fontWeight: theme.typography.fontWeightLight,
    marginBottom: theme.spacing(1),
}))

const MountProperties = styled('div')(({ theme }) => ({
    display: 'grid',
    gridTemplateColumns: 'auto 1fr',
    columnGap: theme.spacing(2),
    rowGap: theme.spacing(0.5),
}))

const ExternalDocumentation = styled('span')(({ theme }) => ({
    position: 'relative',
    width: 0,
    height: 0,
    '& > *': {
        position: 'absolute',
        top: 'calc(-50% + 0.2ex)',
    }
}))

const PropertyName = styled('div')(({ theme }) => ({
    gridColumn: '1/2',
    whiteSpace: 'nowrap',
    alignSelf: 'baseline',
    lineHeight: theme.typography.body1.lineHeight,
}))

const PropertyValue = styled('div')(({ theme }) => ({
    gridColumn: '2/3',
    fontWeight: theme.typography.fontWeightLight,
    alignSelf: 'baseline',
    lineHeight: theme.typography.body1.lineHeight,
}))


interface NameValueRowProps {
    name: React.ReactNode
    value: React.ReactNode
}

/**
 * Renders a single key-value grid row.
 */
const NameValueRow = ({ name, value }: NameValueRowProps) => {
    return <>
        <PropertyName>{name}:</PropertyName>
        <PropertyValue>{value}</PropertyValue>
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


export interface MountpointInfoProps {
    /** mount point information object. */
    mountpoint: MountPoint
    /** 
     * map of all discovered namespaces for mountpoint namespace root path
     * lookups.
     */
    namespaces: NamespaceMap
}

/**
 * Renders detail information about a specific mount point.
 */
export const MountpointInfo = ({ mountpoint, namespaces }: MountpointInfoProps) => {

    const options = mountpoint.mountoptions
        .sort((opt1, opt2) => opt1.localeCompare(opt2, undefined, { numeric: true }))
        .map((opt, idx) => [
            idx > 0 && <>,<br /></>,
            <>{opt}</>
        ])

    // Please note: mount point tags cannot contain spaces in their names or
    // values, as spaces are used as separators between tags. Values are
    // optional.
    const tags = Object.entries(mountpoint.tags)
        .sort(([tagname1,], [tagname2,]) => tagname1.localeCompare(tagname2, undefined, { numeric: true }))
        .map(([tagname, tagvalue], idx) => [
            idx > 0 && <br />,
            <>{tagname}{tagvalue ? `:${tagvalue}` : ''}</>
        ])

    const parent = <>
        {mountpoint.parentid}
        {mountpoint.parent && <> ~ {mountpoint.parent.mountpoint}</>}
    </>

    // Determine the mount point's propagation mode from the tags the Linux
    // kernel is showing us ...or not.
    const propagationmodes = [
        mountpoint.tags['shared'] && 'shared',
        mountpoint.tags['master'] && 'slave',
        mountpoint.tags['unbindable'] && 'unbindable',
        !(mountpoint.tags['shared'] || mountpoint.tags['master'] || mountpoint.tags['unbindable']) && 'private',
    ].filter(propmode => propmode)

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
        <MountPathTitle>
            <MountpointPath drum="always" plainpath={true} mountpoint={mountpoint} />
        </MountPathTitle>
        <MountProperties>
            <NameValueRow
                name={'mount namespace'}
                value={<NamespaceInfo shortprocess={true} namespace={mountpoint.mountnamespace} />}
            />
            <NameValueRow name="device" value={`${mountpoint.major}:${mountpoint.minor}`} />
            <NameValueRow name="filesystem type" value={<>
                {mountpoint.fstype}
                &nbsp;<ExternalDocumentation>
                    <Tooltip title="open external filesystem documentation">
                        <IconButton
                            color="primary"
                            size="small"
                            aria-label="external documentation"
                            href={filesystemTypeLink(mountpoint.fstype)}
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            <MenuBookIcon />
                        </IconButton>
                    </Tooltip>
                </ExternalDocumentation>
            </>}
            />
            <NameValueRow name="root" value={<MountpointRoot root={mountpoint.root} namespaces={namespaces} />} />
            <NameValueRow name="options" value={options} />
            <NameValueRow name="superblock options" value={<Options options={mountpoint.superoptions.split(',')} />} />
            <NameValueRow name="source" value={mountpoint.source} />
            <NameValueRow name="propagation mode" value={propagationmodes.join(', ')} />
            {peers && peers.length > 0 && <NameValueRow
                name="peer mounts"
                value={<GroupedPropagationMembers members={peers} />}
            />}
            {masters && masters.length > 0 && <NameValueRow
                name="master peer mounts"
                value={<GroupedPropagationMembers members={masters} />}
            />}
            {slaves && slaves.length > 0 && <NameValueRow
                name="slave mounts"
                value={<GroupedPropagationMembers members={slaves} />}
            />}
            <NameValueRow name="ID" value={mountpoint.mountid} />
            <NameValueRow name="parent ID" value={parent} />
            <NameValueRow name="tags" value={tags} />
        </MountProperties>
    </>
}

export default MountpointInfo
