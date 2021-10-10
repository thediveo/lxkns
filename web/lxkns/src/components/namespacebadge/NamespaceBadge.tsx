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

import Tooltip from '@mui/material/Tooltip'

import { Namespace } from 'models/lxkns'

import { darken, alpha, lighten, Theme } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { NamespaceIcon, namespaceTypeInfo } from 'components/namespaceicon'


// Based on the general idea from https://stackoverflow.com/a/53309284 "how to
// increase space between dotted border dots".
const dashedBorder = (fg: string = '#000', bg: string = '#fff') => `
linear-gradient(to right, ${fg} 50%, ${bg} 0%) top/5px 2px repeat-x,
linear-gradient(${fg} 50%, ${bg} 0%) right/2px 5px repeat-y,
linear-gradient(to right, ${fg} 50%, ${bg} 0%) bottom/5px 2px repeat-x,
linear-gradient(${fg} 50%, ${bg} 0%) left/2px 5px repeat-y`

// Creates a dashed border based on the badge (background) color as defined for
// the specified type of namespace. Please note that we need to explicity define
// backgroundColor again, as it gets trashed when setting the background to
// achieve a dashed border; for this reason we return an object with background
// and backgroundColor instead of just a background CSS property value string.
const themedDashedBorder = (nstype: string, theme: Theme, shared?: 'shared') => {
    const color = shared ? alpha(theme.palette.namespace[nstype], 0.15) : theme.palette.namespace[nstype]
    const change = shared ? 0.6 : 0.4
    return {
        background: dashedBorder(
            theme.palette.mode === 'light' ? darken(color, change) : lighten(color, change),
            color),
        backgroundColor: color,
    }
}

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

    },
    shared: {
        color: theme.palette.text.disabled,
        fontWeight: theme.typography.fontWeightLight,
    },
    // ...and now for the namespace-type specific styling.
    cgroup: {
        backgroundColor: theme.palette.namespace.cgroup,
    },
    ipc: {
        backgroundColor: theme.palette.namespace.ipc,
    },
    mnt: {
        backgroundColor: theme.palette.namespace.mnt,
    },
    net: {
        backgroundColor: theme.palette.namespace.net,
    },
    pid: {
        backgroundColor: theme.palette.namespace.pid,
    },
    user: {
        width: '9.5em',
        textAlign: 'center',
        backgroundColor: theme.palette.namespace.user,
        fontWeight: 'bold',
    },
    uts: {
        backgroundColor: theme.palette.namespace.uts,
    },
    time: {
        backgroundColor: theme.palette.namespace.time,
    },
    // initial namespaces get a dashed border, with the dash color derived from
    // the badge background color.
    initialNamespace: {
        '&$cgroup': themedDashedBorder('cgroup', theme),
        '&$cgroup$shared': themedDashedBorder('cgroup', theme, 'shared'),

        '&$ipc': themedDashedBorder('ipc', theme),
        '&$ipc$shared': themedDashedBorder('ipc', theme, 'shared'),

        '&$mnt': themedDashedBorder('mnt', theme),
        '&$mnt$shared': themedDashedBorder('mnt', theme, 'shared'),

        '&$net': themedDashedBorder('net', theme),
        '&$net$shared': themedDashedBorder('net', theme, 'shared'),

        '&$pid': themedDashedBorder('pid', theme),
        '&$pid$shared': themedDashedBorder('pid', theme, 'shared'),

        '&$user': themedDashedBorder('user', theme),
        '&$user$shared': themedDashedBorder('user', theme, 'shared'),

        '&$uts': themedDashedBorder('uts', theme),
        '&$uts$shared': themedDashedBorder('uts', theme, 'shared'),

        '&$time': themedDashedBorder('time', theme),
        '&$time$shared': themedDashedBorder('time', theme, 'shared'),
    },
}))


export interface NamespaceBadgeProps {
    /** namespace with type, identifier and initial namespace indication. */
    namespace: Namespace
    /** is this a namespace shared with other leader processes? */
    shared?: boolean,
    /** optional tooltip prefix text. */
    tooltipprefix?: string,
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
export const NamespaceBadge = ({ namespace, tooltipprefix, shared, className }: NamespaceBadgeProps) => {

    const classes = useStyles()

    // Ouch ... Tooltip won't display its tooltip on a <> child, but
    // instead we have to use a <span> to make it work as expected...

    // Ouch #2: don't put comments into return statements, as this will break
    // the optimized build. Ouch ouch ouch ... see also issue #8687,
    // https://github.com/facebook/create-react-app/issues/8687 ... which still
    // is open.
    return (
        <Tooltip title={`${tooltipprefix ? tooltipprefix + ' ': ''}${shared ? '«shared» ' : ''} ${namespace.initial ? 'initial': ''} ${namespaceTypeInfo[namespace.type].tooltip} namespace`}>
            <span className={clsx(
                classes.namespaceBadge,
                classes[namespace.type],
                shared && classes.shared,
                className,
                namespace.initial && classes.initialNamespace
            )}>
                <NamespaceIcon type={namespace.type} fontSize="inherit" />
                {namespace.type}:[{namespace.nsid}]
            </span>
        </Tooltip>
    )
}

