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
import { NamespaceProcessTreeDetailComponentProps } from 'components/namespaceprocesstree'
import { compareMountPaths, compareMounts, MountPath, MountPoint, unescapeMountPath } from 'models/lxkns/mount'
import { Namespace } from 'models/lxkns'
import { Button, makeStyles, Tooltip } from '@material-ui/core'
import ChildrenIcon from 'icons/Children'
import FilesystemtypeIcon from 'icons/Filesystemtype'


const useStyles = makeStyles((theme) => ({
    mounttreedetails: {
        '& .MuiSvgIcon-root': {
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.1ex',
        },
    },
    mountpointpath: {
        fontWeight: theme.typography.fontWeightLight,
        border: `1px solid ${theme.palette.text.hint}`,
        borderRadius: '0.3ex',
        padding: '0 0.2em',
    },
    rootpath: {
        border: `1px solid ${theme.palette.text.hint}`,
        borderRadius: '0.3ex',
        padding: '0 0.2em',
        fontWeight: theme.typography.fontWeightBold,
    },
    hiddenmountpoint: {
        fontWeight: theme.typography.fontWeightLight,
        color: theme.palette.text.disabled,
        '& $mountpointpath': {
            textDecoration: 'line-through solid',
            borderStyle: 'dotted',
        },
    },
    notamountpoint: {
        fontWeight: theme.typography.fontWeightLight,
        fontStyle: 'italic',
        '&$mountpointpath': {
            border: 'none',
        }
    },
    more: {
        marginLeft: '0.5em',
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
    mountpoint: MountPoint
    tail: string
    childmountcount: number
}

// Renders a mount point tree label with information about a mount point.
const MountPointLabel = ({ mountpoint, tail, childmountcount }: MountPointLabelProps) => {

    const classes = useStyles()

    const tooltip = `${mountpoint.hidden ? 'overmounted ' : ''}${unescapeMountPath(mountpoint.mountpoint)}`

    const handleMore = () => {
        
    }

    return (
        <span className={clsx(mountpoint.hidden && classes.hiddenmountpoint)}>
            <Tooltip title={tooltip}>
                <span>
                    <span className={clsx(classes.mountpointpath, tail === '/' && classes.rootpath)}>{unescapeMountPath(tail)}</span>
                </span>
            </Tooltip>
            {!mountpoint.hidden && childmountcount > 0 &&
                <> [<ChildrenIcon fontSize="inherit" />&#8239;{childmountcount}]</>}
            <Tooltip title={`filesystem type «${mountpoint.fstype}»`}>
                <span>
                    {' '}
                    <FilesystemtypeIcon fontSize="inherit" />&#8239;{mountpoint.fstype}
                </span>
            </Tooltip>
            <Button
                className={classes.more}
                onClick={handleMore}
            >···</Button>
        </span>
    )
}

interface MountPathLabelProps {
    tail: string
    childmountcount: number
}

// Renders a mount path tree label with a few pieces of information, namely the
// path (tail) and the number of mount points(!) below this (fake) mount path.
const MountPathLabel = ({ tail, childmountcount }: MountPathLabelProps) => {

    const classes = useStyles()

    return (
        <span>
            <span className={clsx(classes.notamountpoint, classes.mountpointpath)}>{tail}</span>
            {' '}[<ChildrenIcon fontSize="inherit" />&#8239;{childmountcount}]
        </span>
    )
}

interface MountPathTreeItemProps {
    namespace: Namespace
    mountpath: MountPath
    parentpath: string
}

/**
 * Renders a single mount path with its child mount paths.
 */
const MountPathTreeItem = ({ namespace, mountpath, parentpath }: MountPathTreeItemProps) => {

    const classes = useStyles()

    const path = mountpath.path
    const tail = path.substr(parentpath.length)
    const prefix = path === '/' ? path : path + "/"

    const childmountcount = countChildMounts(mountpath)

    const childitems = mountpath.children
        .sort(compareMountPaths)
        .map(child => <MountPathTreeItem
            key={`${namespace.nsid}-${child.path}`}
            namespace={namespace}
            mountpath={child}
            parentpath={prefix}
        />)

    if (!mountpath.mounts.length) {
        return (
            <TreeItem
                className={classes.mounttreedetails}
                nodeId={`${namespace.nsid}-${path}`}
                label={<MountPathLabel tail={tail} childmountcount={childmountcount} />}
            >
                {childitems}
            </TreeItem>)
    }
    return <>{mountpath.mounts.sort(compareMounts)
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

export const MountTree = ({ namespace }: MountTreeProps) => {

    return namespace.mountpaths
        ? <MountPathTreeItem namespace={namespace} mountpath={namespace.mountpaths['/']} parentpath="" />
        : <></>
}
