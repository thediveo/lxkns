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
import { useLocation, Link } from "react-router-dom"

import Typography from '@material-ui/core/Typography'
import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'

export interface DrawerLinkItemProps {
    /** drawer item icon, which automatically will be enclosed in ListItemIcon
     * components.
     */
    icon?: React.ReactNode
    /** label of the drawer item. */
    label: React.ReactNode
    /** route path to activate when the user clicks on this drawer item. */
    path: string
}

/**
 * `DrawerLinkItem` renders an individual item inside an
 * [`AppBarDrawer`](#appbardrawer) and links this item to a specific route
 * path. It is a convenience component that simplifies describing the drawer
 * items with their icons and route paths.
 */
export const DrawerLinkItem = ({ icon, label, path }: DrawerLinkItemProps) => {
    const location = useLocation()
    const selected = location.pathname === path

    return (
        <ListItem
            button
            component={Link}
            to={path}
            selected={selected}
        >
            {icon && <ListItemIcon>{icon}</ListItemIcon>}
            <Typography>{label}</Typography>
        </ListItem>
    )
}
