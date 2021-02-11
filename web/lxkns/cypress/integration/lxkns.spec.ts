describe('lxkns app', () => {

    before(() => {
        cy.log('loads')
        cy.visit('/')
        cy.waitForReact(1000, '#root')

        cy.log('refreshes')
        cy.react('Refresher')
            .find('button')
            .first().click()
        cy.getReact('NamespaceInfo')
            .nthNode(0)
            .getProps('namespace').then((netns) => {
                expect(netns).has.property('type', 'user')
                expect(netns).has.property('initial', true)

            })
    })

    it('shows about', () => {
        cy.navigate('/about')
        cy.getReact('About')
        // Make sure that the content actually loaded.
        cy.contains('Version v')
    })

    it('lends a helping hand', () => {
        cy.navigate('/help')
        cy.getReact('Help')
        cy.getReact('HelpViewer')
        // Make sure that the first help chapter actually loaded.
        cy.contains('Linux Kernel Namespaces Discovery')
    })

})

export { }
