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
import { TreeItem } from '@mui/x-tree-view'
import type { NamespaceProcessTreeDetailComponentProps } from 'components/namespaceprocesstree'
import { compareMountPaths, compareMounts, type MountPath, type MountPoint, unescapeMountPath } from 'models/lxkns/mount'
import type { Namespace } from 'models/lxkns'
import { Button, lighten, styled, Tooltip } from '@mui/material'

import ChildrenIcon from 'icons/Children'
import FilesystemtypeIcon from 'icons/Filesystemtype'
import { useMountpointInfoModal } from 'components/mountpointinfomodal'
import FolderOutlinedIcon from '@mui/icons-material/FolderOutlined'
import PeerIcon from 'icons/propagation/Peer'
import SlaveIcon from 'icons/propagation/Slave'
import UnbindableIcon from 'icons/propagation/Unbindable'
import ReadonlyIcon from 'icons/Readonly'


const mpHidden = 'mountpoint-hidden'

const Label = styled('span')(({ theme }) => ({
    whiteSpace: 'nowrap',
    fontWeight: theme.typography.fontWeightLight,
    '& .MuiSvgIcon-root': {
        verticalAlign: 'baseline',
        position: 'relative',
        top: '0.3ex',
    },

    [`&.${mpHidden}`]: {
        fontWeight: theme.typography.fontWeightLight,
        color: theme.palette.text.disabled,
    },
}))

const Count = styled('span')(() => ({
    marginRight: '0.5em',
}))

const ReadOnly = styled('span')(({ theme }) => ({
    color: theme.palette.fstype,
    marginRight: '0.3em',
}))

const FsType = styled('span')(({ theme }) => ({
    color: theme.palette.fstype,
}))

const MoreButton = styled(Button)(() => ({
    marginLeft: '0.5em',
    paddingTop: 0,
    paddingBottom: 0,
    '&.MuiButton-root': {
        minWidth: 0,
    },
}))

const mppRoot = 'mountpoint-root-path'

const MountPointPath = styled('span')(({ theme }) => ({
    fontWeight: theme.typography.fontWeightLight,
    borderRadius: '0.3ex',
    marginRight: '0.5em',

    [`&.${mppRoot}`]: {
        fontWeight: theme.typography.fontWeightBold,
        padding: '0 0.2em',
        background: lighten(theme.palette.namespace.mnt, 0.2),
    },

    [`.${mpHidden} &`]: {
        textDecoration: 'line-through solid',
    },
}))

const NotAMountpoint = styled(MountPointPath)(() => ({
    fontStyle: 'italic',
    marginLeft: '0.2em',
    marginRight: '0.7em',
    border: 'none',
    background: 'none',
}))

const PropagationMode = styled(MountPointPath)(({ theme }) => ({
    marginLeft: '0.5em',
    color: theme.palette.fstype,
    textDecoration: 'none !important',
}))


// Reduce function returning the sum of all mount points in for this mount path
// as well as for all its child mount paths.
const countMounts = (sum: number, mp: MountPath): number =>
    mp.mounts.length + mp.children.reduce(countMounts, sum)

// Calculate the sum of all mount points in all child mount paths of this mount
// path, excluding the mount points of this mount path itself.
const countChildMounts = (mp: MountPath) =>
    mp.children.reduce(countMounts, 0)


interface MountPointLabelProps {
    /** mount point object with lots of details. */
    mountpoint: MountPoint
    /** mount point tail path to render instead of full mount path. */
    tail: string
    /** number of mount points in all child mount paths */
    childmountcount: number
}

// Renders a mount point tree label with information about a mount point.
const MountPointLabel = ({ mountpoint, tail, childmountcount }: MountPointLabelProps) => {
    const tooltip = `${mountpoint.hidden ? 'overmounted ' : ''}${unescapeMountPath(mountpoint.mountpoint)}`

    const setMountpoint = useMountpointInfoModal()

    const propagationmodes = [
        mountpoint.tags['shared'] &&
        <Tooltip key="shared" title="propagation between peers and to slaves">
            <PropagationMode><PeerIcon fontSize="inherit" />&nbsp;({mountpoint.tags['shared']})</PropagationMode>
        </Tooltip>,
        mountpoint.tags['master'] &&
        <Tooltip key="master" title="propagation from master(s)">
            <PropagationMode><SlaveIcon fontSize="inherit" />&nbsp;({mountpoint.tags['master']})</PropagationMode>
        </Tooltip>,
        mountpoint.tags['unbindable'] &&
        <Tooltip key="unbindable" title="unbindable mount point">
            <PropagationMode><UnbindableIcon fontSize="inherit" /></PropagationMode>
        </Tooltip>,
    ].filter(propmode => propmode)

    const handleMore = (event: React.MouseEvent<HTMLButtonElement>) => {
        event.stopPropagation()
        if (setMountpoint) setMountpoint(mountpoint)
    }

    return (
        <Label className={clsx(mountpoint.hidden && mpHidden)}>
            <Tooltip title={tooltip}>
                <MountPointPath className={clsx(tail === '/' && mppRoot)}>{unescapeMountPath(tail)}</MountPointPath>
            </Tooltip>
            {!mountpoint.hidden && childmountcount > 0 &&
                <Count>[<ChildrenIcon fontSize="inherit" />&nbsp;{childmountcount}]</Count>}
            {mountpoint.mountoptions.includes('ro') &&
                <Tooltip title="read-only">
                    <ReadOnly><ReadonlyIcon fontSize="inherit" />&nbsp;</ReadOnly>
                </Tooltip>}
            <Tooltip title={`filesystem type «${mountpoint.fstype}»`}>
                <FsType>
                    <FilesystemtypeIcon fontSize="inherit" />&#8239;{mountpoint.fstype}
                </FsType>
            </Tooltip>
            {propagationmodes}
            <Tooltip title="mountpoint details">
                <MoreButton onClick={handleMore}>···</MoreButton>
            </Tooltip>
        </Label>
    )
}

interface MountPathLabelProps {
    /** mount point tail path to render instead of full mount path. */
    tail: string
    /** number of mount points in all child mount paths */
    childmountcount: number
}

/**
 * Renders a mount path tree label with a few pieces of information, namely the
 * path (tail) and the number of mount points(!) below this (fake) mount path.
 */
const MountPathLabel = ({ tail, childmountcount }: MountPathLabelProps) => {
    return (
        <Label>
            <FolderOutlinedIcon fontSize="inherit" color="disabled" />
            <NotAMountpoint>{tail}</NotAMountpoint>
            [<ChildrenIcon fontSize="inherit" />&#8239;{childmountcount}]
        </Label>
    )
}

interface MountPathTreeItemProps {
    /** namespace the mount path belongs to. */
    namespace: Namespace
    /** mount path object with zero or more mount points. */
    mountpath: MountPath
    /** 
     * parent mount path, used to render only the mount path tail after the
     * parent mount "base" path. 
     */
    parentpath: string
}

/**
 * Renders a single mount path – which may consist of multiple mount points –
 * with all the child mount paths. In case of multiple mount points for the same
 * mount path, multiple tree items are rendered and the mount points get sorted
 * by visibility (hidden/overmounted mount points first). Any child mount paths
 * are always rendered only under the last mount point item.
 */
const MountPathTreeItem = ({ namespace, mountpath, parentpath }: MountPathTreeItemProps) => {
    const path = mountpath.path
    const tail = path.substring(parentpath.length)
    const prefix = path === '/' ? path : path + "/"

    const childmountcount = countChildMounts(mountpath)

    // Regardless whether this is a mount path with mount points or just a
    // "fake" intermediate mount path node without any mount points, we always
    // need to recursively render any mount points for the child mount paths.
    const childitems = mountpath.children
        .sort(compareMountPaths)
        .map(child => <MountPathTreeItem
            key={`${namespace.nsid}-${child.path}`}
            namespace={namespace}
            mountpath={child}
            parentpath={prefix}
        />)

    if (!mountpath.mounts.length) {
        // This is a "fake" mount path node that has no mount points and instead
        // acts as a folder node. It represents the longest common prefix path
        // of all child mount path nodes.
        return (
            <TreeItem
                key={`${namespace.nsid}-${path}`}
                itemId={`${namespace.nsid}-${path}`}
                label={<MountPathLabel tail={tail} childmountcount={childmountcount} />}
            >
                {childitems}
            </TreeItem>)
    }
    // A mount path with one or more mount points on the same mount path.
    return <>{mountpath.mounts
        .sort(compareMounts)
        .map((mountpoint, idx) =>
            <TreeItem
                key={mountpoint.mountid}
                itemId={`${namespace.nsid}-${path}-${mountpoint.mountid}`}
                label={<MountPointLabel mountpoint={mountpoint} tail={tail} childmountcount={childmountcount} />}
            >
                {idx === mountpath.mounts.length - 1 && childitems}
            </TreeItem>)
    }</>
}


/**
 * Renders the tree of all mount paths with its mount points from the specified
 * mount namespace.
 */
export const MountTree = ({ namespace }: NamespaceProcessTreeDetailComponentProps) => {
    return namespace.mountpaths
        ? <MountPathTreeItem
            namespace={namespace}
            mountpath={namespace.mountpaths['/']}
            parentpath="" />
        : <></>
}
