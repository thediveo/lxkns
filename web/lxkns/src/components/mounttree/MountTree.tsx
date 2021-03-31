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

import { TreeItem } from '@material-ui/lab'
import { NamespaceProcessTreeDetailComponentProps } from 'components/namespaceprocesstree'
import { MountPath, unescapeMountPath } from 'models/lxkns/mount'
import { Namespace } from 'models/lxkns'
import { MountpointIcon } from 'icons/Mountpoint'
import { makeStyles } from '@material-ui/core'
import MountpointchildrenIcon from 'icons/Mountpointchildren'


const useStyles = makeStyles({
    mounttreedetails: {
        '& .MuiSvgIcon-root': {
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.05ex',
        },
    }
})

// Reduce function returning the (recursive) sum of children and grand-children
// plus this namespace itself.
const countMounts = (sum: number, mp: MountPath) =>
    sum + mp.children.reduce(countMounts, mp.mounts.length)


const MountPathTreeItem = (namespace: Namespace, mountpath: MountPath, parentpath: string) => {

    const classes = useStyles()

    const path = mountpath.path
    const tail = path.substr(parentpath.length)
    const prefix = path === '/' ? path : path + "/"

    const childmountcount = countMounts(-1, mountpath) // FIXME: incorrect sum
    const label = <>
        {unescapeMountPath(tail)}
        {childmountcount > 0 && <> [<MountpointchildrenIcon fontSize="inherit" />&#8239;{childmountcount}]</>}
    </>

    return <TreeItem
        className={classes.mounttreedetails}
        nodeId={`${namespace.nsid}-${path}`}
        label={label}
    >{mountpath.children
        .sort((childA, childB) => childA.path.localeCompare(childB.path))
        .map(child => MountPathTreeItem(namespace, child, prefix))}</TreeItem>
}


export interface MountTreeProps extends NamespaceProcessTreeDetailComponentProps { }

export const MountTree = ({ namespace }: MountTreeProps) => {

    return namespace.mountpaths
        ? MountPathTreeItem(namespace, namespace.mountpaths['/'], '')
        : <></>
}
