// MDXv2 remark/rehype plugin configuration for sharing between CRA/CRACO and
// React Styleguidist.

// There's only typographic-em-dashes that covers US typographic style, but no
// need for a full-blown npm module just to get European en dash typography.
const textrTypoEnDashes = (input) => {
    return input
        .replace(/ -- /gim, ' – ')
}

module.exports = async () => {
    const remarkGfm = (await import('remark-gfm')).default
    const remarkImages = (await import('remark-images')).default
    const remarkTextr = (await import('remark-textr')).default
    const remarkGEmoji = (await import('remark-gemoji')).default
    const rehypeSlug = (await import('rehype-slug')).default
    const textrTypoApos = (await import('typographic-apostrophes')).default
    const textrTypoQuotes = (await import('typographic-quotes')).default
    const textrTypoPossPluralsApos = (await import('typographic-apostrophes-for-possessive-plurals')).default
    const textrTypoEllipses = (await import('typographic-ellipses')).default
    //const textrTypoEmDashes = (await import('typographic-em-dashes')).default
    const textrTypoNumberEnDashes = (await import('typographic-en-dashes')).default

    return {
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
                    // textrTypoEmDashes,
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
}
