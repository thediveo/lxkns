import remarkGfm from 'remark-gfm'
import remarkImages from 'remark-images'
import remarkTextr from 'remark-textr'
import remarkGEmoji from 'remark-gemoji'

import textrTypoApos from 'typographic-apostrophes'
import textrTypoQuotes from 'typographic-quotes'
import textrTypoPossPluralsApos from 'typographic-apostrophes-for-possessive-plurals'
import textrTypoEllipses from 'typographic-ellipses'
import textrTypoNumberEnDashes from 'typographic-en-dashes'

import rehypeSlug from 'rehype-slug'
import type { Options } from '@mdx-js/rollup'

// There's only typographic-em-dashes that covers US typographic style, but no
// need for a full-blown npm module just to get European en dash typography.
const textrTypoEnDashes = (input: string) => {
    return input
        .replace(/ -- /gim, ' – ')
}

export const mdxConfiguration: Readonly<Options> = {
    remarkPlugins: [
        remarkGfm,
        remarkImages,
        remarkGEmoji,
        [remarkTextr, {
            plugins: [
                textrTypoApos,
                textrTypoQuotes,
                textrTypoPossPluralsApos,
                textrTypoEllipses,
                textrTypoNumberEnDashes,
                textrTypoEnDashes,
            ],
            options: {
                locale: 'en-us'
            }
        }],
    ],
    rehypePlugins: [
        rehypeSlug,
    ],
}