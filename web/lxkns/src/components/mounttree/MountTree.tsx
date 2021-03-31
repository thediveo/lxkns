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
import { MountPath, unescapeMountPath } from 'models/lxkns/mount'
import { Namespace } from 'models/lxkns'
import { makeStyles, Tooltip } from '@material-ui/core'
import ChildrenIcon from 'icons/Children'


const useStyles = makeStyles((theme) => ({
    mounttreedetails: {
        '& .MuiSvgIcon-root': {
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.05ex',
        },
    },
    hiddenmountpoint: {
        color: theme.palette.text.disabled,
        textDecoration: 'line-through',
    },
    notamountpoint: {
        fontStyle: 'italic',
    }
}))

// Reduce function returning the sum of all mount points in for this mount path
// as well as for all its child mount paths.
const countMounts = (sum: number, mp: MountPath) =>
    mp.mounts.length + mp.children.reduce(countMounts, sum)

// Calculate the sum of all mount points in all child mount paths of this mount
// path, excluding the mount points of this mount path itself.
const countChildMounts = (mp: MountPath) =>
    mp.children.reduce(countMounts, 0)


/**
 * Renders a single mount path with its child mount paths.
 */
const MountPathTreeItem = (namespace: Namespace, mountpath: MountPath, parentpath: string) => {

    const classes = useStyles()

    const path = mountpath.path
    const tail = path.substr(parentpath.length)
    const prefix = path === '/' ? path : path + "/"

    const childmountcount = countChildMounts(mountpath)
    const label = <>
        <Tooltip title={unescapeMountPath(path)}>
            <span className={clsx(
                mountpath.mounts.length === 0 && classes.notamountpoint)}>
                {unescapeMountPath(tail)}
            </span>
        </Tooltip>
        {childmountcount > 0 && <> [<ChildrenIcon fontSize="inherit" />&#8239;{childmountcount}]</>}
    </>

    const childitems = mountpath.children
        .sort((childA, childB) => childA.path.localeCompare(childB.path))
        .map(child => MountPathTreeItem(namespace, child, prefix))

    return <TreeItem
        className={classes.mounttreedetails}
        nodeId={`${namespace.nsid}-${path}`}
        label={label}
    >{childitems}</TreeItem>
}


export interface MountTreeProps extends NamespaceProcessTreeDetailComponentProps { }

export const MountTree = ({ namespace }: MountTreeProps) => {

    return namespace.mountpaths
        ? MountPathTreeItem(namespace, namespace.mountpaths['/'], '')
        : <></>
}
