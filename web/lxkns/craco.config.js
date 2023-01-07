const { addAfterLoader, loaderByName } = require('@craco/craco')

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
                return webpackConfig
            }
        }
    }
}
