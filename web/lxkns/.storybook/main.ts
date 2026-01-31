// Copyright 2023 Harald Albrecht.
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

import type { StorybookConfig as StorybookViteConfig } from '@storybook/react-vite'
import { mdxConfiguration } from '../src/mdxconfig.ts'
import type { Plugin } from 'vite'

const config: StorybookViteConfig = {
    framework: {
        name: '@storybook/react-vite',
        options: {},
    },

    stories: [
        '../src/**/*.stories.@(ts|tsx)',
        '../src/*.mdx',
    ],

    addons: [
        {
            name: '@storybook/addon-docs',
            options: {
                mdxPluginOptions: {
                    mdxCompileOptions: {
                        ...mdxConfiguration,
                    }
                },
            }
        },
        '@storybook/addon-links',
    ],

    docs: {
        defaultName: 'Description',
    },

    core: {
        disableTelemetry: true,
        disableWhatsNewNotifications: true,
    },

    typescript: {
        check: true,
    },

    async viteFinal(config) {
        // drop the @mdx-js/rollup plugin that we get from the vite
        // configuration, as this otherwise causes problems with the mdx plugin
        // brought in by @storybook/addon-docs.
        config.plugins = config.plugins?.filter(e => (e as Plugin)?.name !== '@mdx-js/rollup')
        return config
    },

}

export default config
