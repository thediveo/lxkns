const { addAfterLoader, loaderByName } = require('@craco/craco')

// https://github.com/facebook/create-react-app/pull/11886#issuecomment-1055054685
const ForkTsCheckerWebpackPlugin =
    process.env.TSC_COMPILE_ON_ERROR === 'true'
        ? require('react-dev-utils/ForkTsCheckerWarningWebpackPlugin')
        : require('react-dev-utils/ForkTsCheckerWebpackPlugin');

module.exports = async (env) => {
    const remarkGfm = (await import('remark-gfm')).default
    const remarkImages = (await import('remark-images')).default
    const remarkTextr = (await import('remark-textr')).default
    const rehypeSlug = (await import('rehype-slug')).default
    const textrTypoApos = (await import('typographic-apostrophes')).default
    const textrTypoQuotes = (await import('typographic-quotes')).default
    const textrTypoPossPluralsApos = (await import('typographic-apostrophes-for-possessive-plurals')).default
    const textrTypoEllipses = (await import('typographic-ellipses')).default
    const textrTypoEmDashes = (await import('typographic-em-dashes')).default
    const textrTypoEnDashes = (await import('typographic-en-dashes')).default

    return {
        webpack: {
            configure: (webpackConfig) => {
                addAfterLoader(webpackConfig, loaderByName('babel-loader'), {
                    test: /\.(md|mdx)$/,
                    loader: require.resolve('@mdx-js/loader'),
                    /** @type {import('@mdx-js/loader').Options} */
                    options: {
                        remarkPlugins: [
                            remarkGfm,
                            remarkImages,
                            [remarkTextr, {
                                plugins: [
                                    textrTypoApos,
                                    textrTypoQuotes,
                                    textrTypoPossPluralsApos,
                                    textrTypoEllipses,
                                    // textrTypoEmDashes,
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
                })
                // https://github.com/facebook/create-react-app/pull/11886#issuecomment-1055054685
                webpackConfig.plugins.forEach((plugin) => {
                    if (plugin instanceof ForkTsCheckerWebpackPlugin) {
                        plugin.options.issue.exclude.push({file: '**/src/**/?(*.){spec,test,cy}.*'});
                    }
                })
                return webpackConfig
            }
        }
    }
}
