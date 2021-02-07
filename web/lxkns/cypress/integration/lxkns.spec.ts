describe('lxkns end-to-end', () => {

    beforeEach(() => {
        cy.visit('/')
        cy.waitForReact(1000, '#root')
    })

    it('refreshes', () => {
        cy.react('Refresher').react('IconButton').click()
    })

})