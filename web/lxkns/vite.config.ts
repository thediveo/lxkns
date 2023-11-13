import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import eslint from 'vite-plugin-eslint'
import viteTsconfigPaths from 'vite-tsconfig-paths'
import svgr from 'vite-plugin-svgr'
import mdx from '@mdx-js/rollup'

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

// There's only typographic-em-dashes that covers US typographic style, but no
// need for a full-blown npm module just to get European en dash typography.
const textrTypoEnDashes = (input) => {
    return input
        .replace(/ -- /gim, ' – ')
}

export default defineConfig(() => {
    return {
        base: './',
        server: {
            proxy: {
                '/api': 'http://localhost:5010',
            },
        },
        build: {
            outDir: 'build',
            // https://stackoverflow.com/questions/69260715/skipping-larger-chunks-while-running-npm-run-build
            rollupOptions: {
                output:{
                    manualChunks(id) {
                      if (id.includes('node_modules')) {
                          return id.toString().split('node_modules/')[1].split('/')[0].toString();
                      }
                  }
                }
            },
        },
        plugins: [
            {
                enforce: 'pre',
                ...mdx({
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
                })
            },
            react({
                jsxImportSource: '@emotion/react',
                babel: {
                    plugins: [
                        '@emotion/babel-plugin',
                    ],
                },
            }),
            eslint({
                exclude: ['/virtual:/**', 'node_modules/**'],
            }),
            viteTsconfigPaths(),
            svgr({
                svgrOptions: {
                    icon: true,
                }
            }),
        ],
    }
})
