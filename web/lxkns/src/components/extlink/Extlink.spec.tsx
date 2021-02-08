import React from 'react'
import { mount } from '@cypress/react'
import ExtLink from './ExtLink'

describe('ExtLink', () => {

    it('adorns external links', () => {
        mount(
            <ExtLink href="http://endoftheinternet.null">End of The Internet</ExtLink>
        )
        cy.waitForReact()
        cy.react('ExtLink')
            .find('svg')
            .should('exist')
        cy.react('ExtLink')
            .find('a')
            .contains('End of The Internet')
            .should('have.attr', 'href', 'http://endoftheinternet.null')
            .and('have.attr', 'target', '_blank')
            .and('have.attr', 'rel').then((rel) => {
                expect(rel).matches(/\bnoopener\b/)
                expect(rel).matches(/\bnoreferrer\b/)
            })
    })

})