import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tsconfigPaths from 'vite-tsconfig-paths'
import svgr from 'vite-plugin-svgr'
import mdx from '@mdx-js/rollup'
import path from 'path'

import { mdxConfiguration } from './src/mdxconfig.js'

const srcs = [
    'app',
    'components',
    'hooks',
    'icons',
    'models',
    'theming',
    'utils',
    'views',
]

// https://vite.dev/config/
export default defineConfig({
    base: './',
    build: {
        outDir: 'build'
    },
    server: {
        host: "0.0.0.0",
        port: 3300,
        proxy: {
            '/api': 'http://localhost:5010',
        },
    },
    resolve: {
        alias: Object.fromEntries(
            srcs.map(d => [d, path.resolve(__dirname, `src/${d}`)])
        )
    },
    plugins: [
        {
            enforce: 'pre',
            ...mdx(mdxConfiguration)
        },
        tsconfigPaths(),
        react(),
        svgr({
            svgrOptions: {
                icon: true,
            }
        }),
    ]
})
