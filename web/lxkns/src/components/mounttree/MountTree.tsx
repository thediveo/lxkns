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
import { TreeItem } from '@material-ui/lab'
import { NamespaceProcessTreeDetailComponentProps, NamespaceProcessTreeTreeDetails } from 'components/namespaceprocesstree'
import { compareMountPaths, compareMounts, MountPath, MountPoint, unescapeMountPath } from 'models/lxkns/mount'
import { Namespace, NamespaceMap } from 'models/lxkns'
import { Button, lighten, makeStyles, Tooltip } from '@material-ui/core'
import ChildrenIcon from 'icons/Children'
import FilesystemtypeIcon from 'icons/Filesystemtype'
import { useMountpointInfoModal } from 'components/mountpointinfomodal'
import FolderOutlinedIcon from '@material-ui/icons/FolderOutlined'
import PeerIcon from 'icons/propagation/Peer'
import SlaveIcon from 'icons/propagation/Slave'
import UnbindableIcon from 'icons/propagation/Unbindable'
import ReadonlyIcon from 'icons/Readonly'


const useStyles = makeStyles((theme) => ({
    mounttreedetails: {
    },
    label: {
        whiteSpace: 'nowrap',
        fontWeight: theme.typography.fontWeightLight,
        '& .MuiSvgIcon-root': {
            verticalAlign: 'baseline',
            position: 'relative',
            top: '0.3ex',
        },
    },
    mountpointpath: {
        fontWeight: theme.typography.fontWeightLight,
        borderRadius: '0.3ex',
        marginRight: '0.5em',
    },
    rootpath: {
        fontWeight: theme.typography.fontWeightBold,
        padding: '0 0.2em',
        background: lighten(theme.palette.namespace.mnt, 0.2),
    },
    hiddenmountpoint: {
        fontWeight: theme.typography.fontWeightLight,
        color: theme.palette.text.disabled,
        '& $mountpointpath': {
            textDecoration: 'line-through solid',
        },
    },
    notamountpoint: {
        fontWeight: theme.typography.fontWeightLight,
        fontStyle: 'italic',
        marginLeft: '0.2em',
        marginRight: '0.7em',
        '&$mountpointpath': {
            border: 'none',
            background: 'none',
        }
    },
    childcount: {
        marginRight: '0.5em',
    },
    ro: {
        color: theme.palette.fstype,
        marginRight: '0.3em',
    },
    fstype: {
        color: theme.palette.fstype,
    },
    propmode: {
        marginLeft: '0.5em',
        color: theme.palette.fstype,
    },
    more: {
        marginLeft: '0.5em',
        paddingTop: 0,
        paddingBottom: 0,
        '&.MuiButton-root': {
            minWidth: 0,
        },
    },
}))

// Reduce function returning the sum of all mount points in for this mount path
// as well as for all its child mount paths.
const countMounts = (sum: number, mp: MountPath) =>
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

    const classes = useStyles()

    const tooltip = `${mountpoint.hidden ? 'overmounted ' : ''}${unescapeMountPath(mountpoint.mountpoint)}`

    const setMountpoint = useMountpointInfoModal()

    const propagationmodes = [
        mountpoint.tags['shared'] &&
        <Tooltip title="propagation between peers and to slaves">
            <span className={classes.propmode} ><PeerIcon fontSize="inherit" />&nbsp;({mountpoint.tags['shared']})</span>
        </Tooltip>,
        mountpoint.tags['master'] &&
        <Tooltip title="propagation from master(s)">
            <span className={classes.propmode} ><SlaveIcon fontSize="inherit" />&nbsp;({mountpoint.tags['master']})</span>
        </Tooltip>,
        mountpoint.tags['unbindable'] &&
        <Tooltip title="unbindable mount point">
            <span className={classes.propmode} ><UnbindableIcon fontSize="inherit" /></span>
        </Tooltip>,
    ].filter(propmode => propmode)

    const handleMore = (event: React.MouseEvent<HTMLButtonElement>) => {
        event.stopPropagation()
        setMountpoint(mountpoint)
    }

    return (
        <span className={clsx(classes.label, mountpoint.hidden && classes.hiddenmountpoint)}>
            <Tooltip title={tooltip}>
                <span>
                    <span className={clsx(classes.mountpointpath, tail === '/' && classes.rootpath)}>{unescapeMountPath(tail)}</span>
                </span>
            </Tooltip>
            {!mountpoint.hidden && childmountcount > 0 &&
                <span className={classes.childcount}>[<ChildrenIcon fontSize="inherit" />&nbsp;{childmountcount}]</span>}
            {mountpoint.mountoptions.includes('ro') &&
                <Tooltip title="read-only">
                    <span className={classes.ro}><ReadonlyIcon fontSize="inherit" />&nbsp;</span>
                </Tooltip>}
            <Tooltip title={`filesystem type «${mountpoint.fstype}»`}>
                <span className={classes.fstype}>
                    <FilesystemtypeIcon fontSize="inherit" />&#8239;{mountpoint.fstype}
                </span>
            </Tooltip>
            {propagationmodes}
            <Tooltip title="mountpoint details">
                <Button className={classes.more} onClick={handleMore}>···</Button>
            </Tooltip>
        </span>
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

    const classes = useStyles()

    return (
        <span className={classes.label}>
            <FolderOutlinedIcon fontSize="inherit" color="disabled" />
            <span className={clsx(classes.notamountpoint, classes.mountpointpath)}>{tail}</span>
            [<ChildrenIcon fontSize="inherit" />&#8239;{childmountcount}]
        </span>
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

    const classes = useStyles()

    const path = mountpath.path
    const tail = path.substr(parentpath.length)
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
                className={classes.mounttreedetails}
                nodeId={`${namespace.nsid}-${path}`}
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
                className={classes.mounttreedetails}
                key={mountpoint.mountid}
                nodeId={`${namespace.nsid}-${path}-${mountpoint.mountid}`}
                label={<MountPointLabel mountpoint={mountpoint} tail={tail} childmountcount={childmountcount} />}
            >
                {idx === mountpath.mounts.length - 1 && childitems}
            </TreeItem>)
    }</>
}


export interface MountTreeProps extends NamespaceProcessTreeDetailComponentProps { }

/**
 * Renders the tree of all mount paths with its mount points from the specified
 * mount namespace.
 */
export const MountTree = ({ namespace }: MountTreeProps) => {
    return namespace.mountpaths
        ? <MountPathTreeItem
            namespace={namespace}
            mountpath={namespace.mountpaths['/']}
            parentpath="" />
        : null
}

/**
 * Returns the list of all tree node ids to be expanded. However, mount points
 * which cross a certain threshold of child mount points won't be expanded
 * though. 
 */
const expandAll = (namespaces: NamespaceMap) => Object.values(namespaces)
    .map(ns => ns.mountpaths
        ? Object.values(ns.mountpaths)
            .map(mountpath => mountpath.mounts
                    .filter(mountpoint => mountpoint.children.length > 0 && mountpoint.children.length <= 50)
                    .map(mountpoint => `${ns.nsid}-${mountpoint.mountpoint}-${mountpoint.mountid}`)
                ).flat()
        : [])
    .flat()

/**
 * This detailer:
 * - provides a factory to render the mount point details of mount namespaces,
 * - supports expanding all detail nodes (well, at least if they don't contain
 *   more than a certain maximum of child mount points).
 * - supports collapsing all detail nodes.
 */
export const MountTreeDetailer: NamespaceProcessTreeTreeDetails = {
    factory: MountTree,
    expandAll: expandAll,
    collapseAll: null,
}
