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
import { MountpointPath } from 'components/mountpointpath'
import { NamespaceInfo } from 'components/namespaceinfo'
import { styled } from '@mui/material'


const MountPathGroup = styled('div')(({ theme }) => ({
    '& + &': {
        marginTop: theme.spacing(1),
    },
}))

const NamespacedMountPaths = styled('ul')(({ theme }) => ({
    listStyleType: 'none',
    margin: 0,
    paddingLeft: theme.spacing(3),
    '& > li': {
    },
}))


export interface GroupedPropagationMembersProps {
    /** list of members of a propagation group */
    members: MountPoint[]
}

/**
 * Renders a list of mount point propagation group members, grouping the mount
 * points by their mount namespaces and then sorting and listing them by path
 * per mount namespace. The mount namespace groups are ordered by their
 * identifiers, that is, by their inode numbers.
 */
export const GroupedPropagationMembers = ({ members }: GroupedPropagationMembersProps) => {
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
    return <>{Object.values(grouped)
        .sort((group1, group2) => group1[0].mountnamespace.nsid - group2[0].mountnamespace.nsid)
        .map(group => {
            const mountns = group[0].mountnamespace
            return <MountPathGroup key={mountns.nsid}>
                <NamespaceInfo shortprocess={true} namespace={mountns} />
                <NamespacedMountPaths>
                    {group
                        .sort(compareMountPeers)
                        .map(peermountpoint =>
                            <li key={peermountpoint.mountpoint}>
                                <MountpointPath showid="hidden" mountpoint={peermountpoint} />
                            </li>)
                    }
                </NamespacedMountPaths>
            </MountPathGroup>
        })
    }</>
}
