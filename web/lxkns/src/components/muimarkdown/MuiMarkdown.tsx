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

import React, { memo } from 'react'
import clsx from 'clsx'

import { MDXProvider } from '@mdx-js/react'

import {
    Divider,
    Table, TableBody, TableCell, TableHead, TableRow,
    Typography,
    makeStyles,
    lighten
} from '@material-ui/core'

import { ChapterSkeleton } from './ChapterSkeleton'


// Defines how to map the components emitted by MDX onto Material-UI components,
// and especially the Typography component. See also:
// https://mdxjs.com/advanced/components
const MuiComponents = {
    // Get us rid of that pesky "validateDOMNesting(...): <p> cannot appear as a
    // descendant of <p>" by using a <div> instead of Typography's default <p>.
    p: (() => {
        const P = (props: any) => <Typography {...props} component="div" />
        return memo(P)
    })(),

    h1: (() => {
        const H1 = (props: any) => <Typography {...props} variant="h4" />
        return memo(H1)
    })(),

    h2: (() => {
        const H2 = (props: any) => <Typography {...props} variant="h5" />
        return memo(H2)
    })(),

    h3: (() => {
        const H3 = (props: any) => <Typography {...props} variant="h6" />
        return memo(H3)
    })(),

    h4: (() => {
        const H4 = (props: any) => <Typography {...props} variant="subtitle1" />
        return memo(H4)
    })(),

    h5: (() => {
        const H5 = (props: any) => <Typography {...props} variant="subtitle2" />
        return memo(H5)
    })(),

    h6: (() => {
        const H6 = (props: any) => <Typography {...props} variant="subtitle2" />
        return memo(H6)
    })(),

    // And once more: get us rid of that pesky "validateDOMNesting(...): <p>
    // cannot appear as a descendant of <p>" by using a <div> instead of
    // Typography's default <p>.
    blockquote: (() => {
        const Blockquote = (props: any) => <Typography {...props} component="div" variant="body2" />
        return memo(Blockquote)
    })(),

    ul: (() => {
        const Ul = (props: any) => <Typography {...props} component="ul" />
        return memo(Ul)
    })(),

    ol: (() => {
        const Ol = (props: any) => <Typography {...props} component="ol" />
        return memo(Ol)
    })(),

    li: (() => {
        const Li = (props: any) => <Typography {...props} component="li" />
        return memo(Li)
    })(),

    table: (() => {
        const MuiTable = (props: any) => <Table {...props} />
        return memo(MuiTable)
    })(),

    tr: (() => {
        const Tr = (props: any) => <TableRow {...props} />
        return memo(Tr)
    })(),

    td: (() => {
        const Td = ({ align, ...props }) => (
            <TableCell align={align || undefined} {...props} />
        )
        return memo(Td)
    })(),

    tbody: (() => {
        const TBody = (props: any) => <TableBody {...props} />
        return memo(TBody)
    })(),

    th: (() => {
        const Th = ({ align, ...props }) => (
            <TableCell align={align || undefined} {...props} />
        )
        return memo(Th)
    })(),

    thead: (() => {
        const THead = (props: any) => <TableHead {...props} />
        return memo(THead)
    })(),

    hr: Divider,
}


// Styles Material-UIs typography elements inside am MDX context to our hearts'
// desires. Additionally styles some Mui components, such as Mui SVG icons to
// fit into the overall styling.
const useStyles = makeStyles((theme) => ({
    markdown: {
        // Make sure to properly reset the text color according to the primary
        // text color.
        color: theme.palette.text.primary,
        // ...and now for the details...
        '& .MuiTypography-h1, & .MuiTypography-h2, & .MuiTypography-h3, & .MuiTypography-h4, & .MuiTypography-h5, & .MuiTypography-h6, & .MuiTypography-subtitle1, & .MuiTypography-subtitle2': {
            color: theme.palette.type === 'light'
                ? theme.palette.primary.main
                : theme.palette.primary.light,
        },
        '& .MuiTypography-h4:first-of-type': {
            marginTop: theme.spacing(1),
        },
        '& .MuiTypography-h4, & .MuiTypography-h5, & .MuiTypography-h6': {
            marginTop: theme.spacing(3),
            marginBottom: theme.spacing(2),
        },
        '& .MuiTypography-subtitle1, & .MuiTypography-subtitle2': {
            marginTop: theme.spacing(2),
            marginBottom: theme.spacing(1),
        },
        '& .MuiTypography-body1 + .MuiTypography-body1': {
            marginTop: theme.spacing(1),
        },
        '& .MuiTypography-body2': {
            margin: theme.spacing(2),
            borderLeft: `${theme.spacing(1)}px solid ${theme.palette.primary.main}`,
            paddingLeft: theme.spacing(1),
        },
        '& .MuiSvgIcon-root.icon': {
            verticalAlign: 'middle',
            fontSize: 'calc(100% + 2px)',
            border: `1px solid ${theme.palette.text.disabled}`,
            padding: 1,
            borderRadius: theme.spacing(1) / 2,
        },
        '& a:link': {
            color: theme.palette.type === 'light'
                ? theme.palette.primary.main
                : theme.palette.primary.light
        },
        '& a:visited': {
            color: theme.palette.type === 'light'
                ? theme.palette.primary.dark
                : lighten(theme.palette.primary.light, 0.3)
        },
        '& a:hover, & a:active': { color: theme.palette.secondary.main },
    }
}))


export interface MuiMarkdownProps {
    /** compiled MDX, which can also be lazy loaded. */
    mdx: (props: any) => JSX.Element
    /** shortcodes, that is, available components. */
    shortcodes?: { [key: string]: React.ComponentType<any> }
    /** CSS class name(s). */
    className?: string
    /** fallback components to render when lazily loading the mdx. */
    fallback?: JSX.Element
}

/**
 * Renders the given MDX using Material-UI `Typography` components (where
 * appropriate). The MDX can be either statically imported beforehand or also
 * lazily imported using `React.lazy()`. This component will handle both use
 * cases transparently: it uses a `React.Suspense` child component and shows a
 * `ChapterSkeleton` component while lazily loading MDX.
 *
 * - uses [mdx-js/mdx](https://github.com/mdx-js/mdx).
 * - headings automatically get `id` slugs via
 *   [remark-slug](https://github.com/remarkjs/remark-slug).
 * - some typography goodies via
 *   [remark-textr](https://github.com/remarkjs/remark-textr):
 *   - typographic ellipsis,
 *   - typgraphic quotes,
 *   - number range endashes,
 *   - turns `--` into emdashes.
 *
 * Please see the [`HelpViewer`](#helpviewer) component for a no-frills help
 * document viewer with multiple chapter support and chapter navigation.
 */
export const MuiMarkdown = ({ mdx: Mdx, className, shortcodes, fallback }: MuiMarkdownProps) => {

    const classes = useStyles()

    return (
        <React.Suspense fallback={fallback || ChapterSkeleton}>
            <MDXProvider components={{ ...MuiComponents, ...shortcodes }}>
                <div className={clsx(className, classes.markdown)}>
                    <Mdx />
                </div>
            </MDXProvider>
        </React.Suspense>
    )
}

export default MuiMarkdown
