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

import {
    Divider,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
    lighten,
    styled,
} from '@mui/material'
import { ChapterSkeleton } from 'components/chapterskeleton'
import type { MDXComponents, MDXContent } from 'mdx/types'


// Defines how to map the components emitted by MDX onto Material-UI components,
// and especially the Typography component. See also:
// https://mdxjs.com/advanced/components
const muiComponents: MDXComponents = {
    // Get us rid of that pesky "validateDOMNesting(...): <p> cannot appear as a
    // descendant of <p>" by using a <div> instead of Typography's default <p>.
    p: (props: object) => (<Typography {...props} component="div" />),

    h1: (props: object) => (<Typography {...props} variant="h4" />),
    h2: (props: object) => (<Typography {...props} variant="h5" />),
    h3: (props: object) => (<Typography {...props} variant="h6" />),
    h4: (props: object) => (<Typography {...props} variant="subtitle1" />),
    h5: (props: object) => (<Typography {...props} variant="subtitle2" />),
    h6: (props: object) => (<Typography {...props} variant="subtitle2" />),

    // And once more: get us rid of that pesky "validateDOMNesting(...): <p>
    // cannot appear as a descendant of <p>" by using a <div> instead of
    // Typography's default <p>.
    blockquote: (props: object) => (<Typography {...props} component="div" variant="body2" />),

    ul: (props: object) => (<Typography {...props} component="ul" />),
    ol: (props: object) => (<Typography {...props} component="ol" />),
    li: (props: object) => (<Typography {...props} component="li" />),

    table: (props: object) => (<Table {...props} />),
    tr: (props: object) => (<TableRow {...props} />),
    td: (props: object) => {
        const { align, ...otherprops }: { align?: 'inherit' | 'left' | 'center' | 'right' | 'justify' } = props
        return <TableCell align={align || undefined} {...otherprops} />
    },
    tbody: (props: object) => (<TableBody {...props} />),
    th: (props: object) => {
        const { align, ...otherprops }: { align?: 'inherit' | 'left' | 'center' | 'right' | 'justify' } = props
        return <TableCell align={align || undefined} {...otherprops} />
    },
    thead: (props: object) => (<TableHead {...props} />),

    hr: () => (<Divider />),
}


// Styles Material-UIs typography elements inside am MDX context to our hearts'
// desires. Additionally styles some Mui components, such as Mui SVG icons to
// fit into the overall styling.
const MarkdownArea = styled('div')(({ theme }) => ({
    // Make sure to properly reset the text color according to the primary
    // text color.
    color: theme.palette.text.primary,
    // ...and now for the details...
    '& .MuiTypography-h1, & .MuiTypography-h2, & .MuiTypography-h3, & .MuiTypography-h4, & .MuiTypography-h5, & .MuiTypography-h6, & .MuiTypography-subtitle1, & .MuiTypography-subtitle2': {
        color: theme.palette.mode === 'light'
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
        borderLeft: `${theme.spacing(1)} solid ${theme.palette.primary.main}`,
        paddingLeft: theme.spacing(1),
    },
    '& .MuiSvgIcon-root.icon': {
        verticalAlign: 'middle',
        fontSize: 'calc(100% + 2px)',
        border: `1px solid ${theme.palette.text.disabled}`,
        padding: 1,
        borderRadius: theme.spacing(0.5),
    },
    '& a:link': {
        color: theme.palette.mode === 'light'
            ? theme.palette.primary.main
            : theme.palette.primary.light
    },
    '& a:visited': {
        color: theme.palette.mode === 'light'
            ? theme.palette.primary.dark
            : lighten(theme.palette.primary.light, 0.3)
    },
    '& a:hover, & a:active': {
        color: theme.palette.secondary.main
    },
    '& code': {
        fontFamily: 'Roboto Mono',
    },
}))


export interface MuiMarkdownProps {
    /** compiled MDX, which can also be lazy loaded. */
    mdx: MDXContent
    /** shortcodes, that is, available components. */
    shortcodes?: MDXComponents // { [key: string]: React.ComponentType<any> }
    /** CSS class name(s). */
    className?: string
    /** fallback components to render when lazily loading the mdx. */
    fallback?: React.JSX.Element
}

/**
 * `MuiMarkdown` renders the given [MDX](https://mdxjs.com/) using Material-UI
 * `Typography` components (where appropriate). The MDX can be either statically
 * imported beforehand or alternatively lazily imported when needed using
 * `React.lazy()`. This component will handle both use cases transparently: it
 * uses a `React.Suspense` child component and shows a
 * [ChapterSkeleton](?path=/docs/universal-chapterskeleton--description)
 * component while lazily loading MDX.
 *
 * - uses [mdx.js](https://github.com/mdx-js/mdx).
 * - headings automatically get `id` slugs via
 *   [rehype-slug](https://github.com/rehypejs/rehype-slug).
 * - typography goodies:
 *   - turns `--` into _endashes_.
 *   - number range _endashes_,
 * - some more typography goodies via
 *   [remark-textr](https://github.com/remarkjs/remark-textr):
 *   - typographic ellipsis,
 *   - typgraphic quotes,
 *   - and more.
 *
 * Please see the [HelpViewer](?path=/docs/universal-helpviewer--description)
 * component for a no-frills help document viewer with multiple chapter support
 * and chapter navigation.
 * 
 * We recommend configuring your MDX integration as follows; pass
 * `mdxConfiguration` to your particular `mdx` setup like in
 * `mdx(mdxConfiguration)`.
 * 
 * ```ts
 * import remarkGfm from 'remark-gfm'
 * import remarkImages from 'remark-images'
 * import remarkTextr from 'remark-textr'
 * import remarkGEmoji from 'remark-gemoji'
 * 
 * import textrTypoApos from 'typographic-apostrophes'
 * import textrTypoQuotes from 'typographic-quotes'
 * import textrTypoPossPluralsApos from 'typographic-apostrophes-for-possessive-plurals'
 * import textrTypoEllipses from 'typographic-ellipses'
 * import textrTypoNumberEnDashes from 'typographic-en-dashes'
 * 
 * import rehypeSlug from 'rehype-slug'
 * import type { Options } from '@mdx-js/rollup'
 * 
 * const textrTypoEnDashes = (input: string) => {
 *     return input
 *         .replace(/ -- /gim, ' – ')
 * }
 * 
 * export const mdxConfiguration: Readonly<Options> = {
 *     remarkPlugins: [
 *         remarkGfm,
 *         remarkImages,
 *         remarkGEmoji,
 *         [remarkTextr, {
 *             plugins: [
 *                 textrTypoApos,
 *                 textrTypoQuotes,
 *                 textrTypoPossPluralsApos,
 *                 textrTypoEllipses,
 *                 textrTypoNumberEnDashes,
 *                 textrTypoEnDashes,
 *             ],
 *             options: {
 *                 locale: 'en-us'
 *             }
 *         }],
 *     ],
 *     rehypePlugins: [
 *         rehypeSlug,
 *     ],
 * }
 * ```
 */
export const MuiMarkdown = ({ mdx: Mdx, className, shortcodes, fallback }: MuiMarkdownProps) => (
    <React.Suspense fallback={fallback || <ChapterSkeleton />}>
        <MarkdownArea className={className}>
            <Mdx components={{ ...muiComponents, ...shortcodes }} />
        </MarkdownArea>
    </React.Suspense>
)

export default MuiMarkdown
