// Copyright 2020 by Harald Albrecht.
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

import { Card, makeStyles } from '@material-ui/core'

const useStyles = makeStyles((theme) => ({
    card: {
        display: 'grid',
        flexDirection: 'column',
        alignItems: 'stretch',
        backgroundColor: theme.palette.background.default,

        '&.paragraph': {
            marginBottom: '2ex',
        },

        gridTemplateColumns: `${theme.spacing(2)}px auto minmax(${theme.spacing(2)}px, 1fr)`,
        gridTemplateRows: `${theme.spacing(2)}px auto minmax(${theme.spacing(2)}px, 1fr)`,
        gridTemplateAreas: '"topleft top topright" "middleleft content middleright" "bottomleft bottom bottomright"',

        '& > .top': { 
            gridArea: 'top',
            borderBottom: '1px dashed #ccc',
        },

        '& > .bottom': { 
            gridArea: 'bottom',
            borderTop: '1px dashed #ccc',
        },

        '& > .left': {
            gridArea: 'middleleft',
            borderRight: '1px dashed #ccc',
        },

        '& > .right': { 
            gridArea: 'middleright',
            borderLeft: '1px dashed #ccc',
        },

        '& > .content': {
            gridArea: 'content',
            background: theme.palette.background.paper,
        },

        '& > .top.left': { gridArea: 'topleft' },
        '& > .top.right': { gridArea: 'topright' },
        '& > .bottom.left': { gridArea: 'bottomleft' },
        '& > .bottom.right': { gridArea: 'bottomright' },

        '& + &': { marginTop: theme.spacing(2) }
    }
}))

export interface ComponentCardProps {
    /** optional maximal width of the children. */
    maxwidth?: string
    /** optional bottom margin. */
    paragraph?: boolean
    /** children components to render inside a `ComponentCard`. */
    children: React.ReactNode
}

/**
 * `ComponentCard` renders its children into a (Material design) card and marks
 * the outer edges of the children. This allows to quickly understand how a
 * component behaves with respect to its outer dimensions. For better contrast,
 * the children are rendered on the palette's paper color, while the card color
 * itself is the default color: white child background for the default palette,
 * and a light gray background for the card surrounding the children.
 *
 * The optional `maxwidth` (notice the *pure lowercase* spelling!) property can
 * be used to restrict the width of the card and thus the rendered children,
 * such as `maxwidth="50%"`. This allows quickly testing how components react
 * when there isn't much horizontal space available without the need to switch
 * the browser window width.
 *
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
export const ComponentCard = ({children, maxwidth, paragraph}: ComponentCardProps) => {
    
    const classes = useStyles()

    return (
        <Card className={clsx(classes.card, paragraph)}>
            <div className="top left"/><div className="top"/><div className="top right"/>
            <div className="left"/><div className="content" style={{maxWidth: maxwidth}}>{children}</div><div className="right"/>
            <div className="bottom left"/><div className="bottom"/><div className="bottom right"/>
        </Card>
    )
}

export default ComponentCard
