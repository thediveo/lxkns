const { addAfterLoader, loaderByName } = require('@craco/craco')
const mdxplagues = require('./mdxplugins.js')

// https://github.com/facebook/create-react-app/pull/11886#issuecomment-1055054685
const ForkTsCheckerWebpackPlugin =
    process.env.TSC_COMPILE_ON_ERROR === 'true'
        ? require('react-dev-utils/ForkTsCheckerWarningWebpackPlugin')
        : require('react-dev-utils/ForkTsCheckerWebpackPlugin');

module.exports = async (env) => {
    const mdxplagueConfig = await mdxplagues()

    return {
        webpack: {
            configure: (webpackConfig) => {
                addAfterLoader(webpackConfig, loaderByName('babel-loader'), {
                    test: /\.(md|mdx)$/,
                    loader: require.resolve('@mdx-js/loader'),
                    /** @type {import('@mdx-js/loader').Options} */
                    options: mdxplagueConfig,
                })
                // https://github.com/facebook/create-react-app/pull/11886#issuecomment-1055054685
                webpackConfig.plugins.forEach((plugin) => {
                    if (plugin instanceof ForkTsCheckerWebpackPlugin) {
                        plugin.options.issue.exclude.push({ file: '**/src/**/?(*.){spec,test,cy}.*' });
                    }
                })
                return webpackConfig
            }
        }
    }
}
