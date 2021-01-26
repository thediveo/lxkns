// Copyright 2020 Harald Albrecht.
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

import Tooltip from '@material-ui/core/Tooltip'

import { Namespace } from 'models/lxkns'

import { makeStyles } from '@material-ui/core'
import { NamespaceIcon, namespaceTypeInfo } from 'components/namespaceicon'


// https://stackoverflow.com/a/53309284
const dashedBorder = (fg: string = '#000', bg: string = '#fff') => `
linear-gradient(to right, ${fg} 50%, ${bg} 0%) top/5px 2px repeat-x,
linear-gradient(${fg} 50%, ${bg} 0%) right/2px 5px repeat-y,
linear-gradient(to right, ${fg} 50%, ${bg} 0%) bottom/5px 2px repeat-x,
linear-gradient(${fg} 50%, ${bg} 0%) left/2px 5px repeat-y`

const useStyles = makeStyles((theme) => ({
    namespaceBadge: {
        minWidth: '11.5em',
        verticalAlign: 'middle',

        display: 'inline-flex',
        justifyContent: 'space-between',
        alignItems: 'center',

        marginTop: '0.2ex',
        marginBottom: '0.2ex',
        paddingLeft: '0.2em',
        paddingRight: '0.2em',
        paddingTop: '0.2ex',
        borderRadius: '0.2em',

        // type icon placement...
        '& .MuiSvgIcon-root': {
            verticalAlign: 'text-top',
            position: 'relative',
            top: '0.05ex',
        },

        // ...and now for the namespace-type specific styling.
        '&$cgroup': {
            backgroundColor: '#fce1e1',
        },
        '&$ipc': {
            backgroundColor: '#f5ffcc',
        },
        '&$mnt': {
            backgroundColor: '#e4f2f5',
        },
        '&$net': {
            backgroundColor: '#e0ffe0',
        },
        '&$pid': {
            backgroundColor: '#daddf2',
        },
        '&$user': {
            width: '9.5em',
            textAlign: 'center',
            backgroundColor: '#e9e8e8',
            fontWeight: 'bold',
        },
        '&$uts': {
            backgroundColor: '#fff2d9',
        },
        '&$time': {
            backgroundColor: '#bdffe8',
        },
    },
    initialNamespace: {
        '&$cgroup': {
            background: dashedBorder('#a68383', '#fce1e1'),
            backgroundColor: '#fce1e1',
        },
        '&$ipc': {
            background: dashedBorder('#a1a885', '#f5ffcc'),
            backgroundColor: '#f5ffcc',
        },
        '&$mnt': {
            background: dashedBorder('#a2adb0', '#e4f2f5'),
            backgroundColor: '#e4f2f5',
        },
        '&$net': {
            background: dashedBorder('#879c87', '#e0ffe0'),
            backgroundColor: '#e0ffe0',
        },
        '&$pid': {
            background: dashedBorder('#9a9dad', '#daddf2'),
            backgroundColor: '#daddf2',
        },
        '&$user': {
            background: dashedBorder('#808080', '#e9e8e8'),
            backgroundColor: '#e9e8e8',
        },
        '&$uts': {
            background: dashedBorder('#a68546', '#fff2d9'),
            backgroundColor: '#fff2d9',
        },
        '&$time': {
            background: dashedBorder('#84b3a2', '#bdffe8'),
            backgroundColor: '#bdffe8',
        },
    },
    // The following is required so we can reference and thus combine
    // selectors for namespace type-specific styling of the "pill".
    cgroup: {},
    ipc: {},
    mnt: {},
    net: {},
    pid: {},
    user: {},
    uts: {},
    time: {}
}))


export interface NamespaceBadgeProps {
    /** namespace with type, identifier and initial namespace indication. */
    namespace: Namespace
    /** optional CSS class name(s). */
    className?: string
}

/**
 * Renders a namespace "badge" (or "pill") consisting of the namespace's type
 * and identifier, in the typical `nstype:[nsid]` textual notation (with only
 * the inode number, but without the device number where the inode lives in).
 *
 * Additionally, the badge gets a namespace type-specific icon. Finally, if the
 * namespace is an initial namespace then it gets visually marked using a dashed
 * border.
 */
export const NamespaceBadge = ({ namespace, className }: NamespaceBadgeProps) => {

    const classes = useStyles()

    // Ouch ... Tooltip won't display its tooltip on a <> child, but
    // instead we have to use a <span> to make it work as expected...

    // Ouch #2: don't put comments into return statements, as this will break
    // the optimized build. Ouch ouch ouch ... see also issue #8687,
    // https://github.com/facebook/create-react-app/issues/8687 ... which still
    // is open.
    return (
        <Tooltip title={`${namespaceTypeInfo[namespace.type].tooltip} namespace`}>
            <span className={clsx(
                classes.namespaceBadge,
                classes[namespace.type],
                className,
                namespace.initial && classes.initialNamespace
            )}>
                <NamespaceIcon type={namespace.type} fontSize="inherit" />
                {namespace.type}:[{namespace.nsid}]
            </span>
        </Tooltip>
    )
}

