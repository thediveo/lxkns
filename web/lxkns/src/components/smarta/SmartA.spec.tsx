import React from 'react'
import { BrowserRouter } from 'react-router-dom'
import { mount } from '@cypress/react'
import SmartA from './SmartA'

describe('SmartA', () => {

    it('adorns external links', () => {
        mount(
            <SmartA href="http://nottheendoftheinternet.null">Not the End of The Internet</SmartA>
        )
        cy.waitForReact()
        cy.get('span')
            .should('exist')
            .contains('Not the End of The Internet')
    })

    it('renders internal links plainly', () => {
        mount(
            <BrowserRouter>
                <SmartA href="/help">help</SmartA>
            </BrowserRouter>
        )
        cy.waitForReact()
        cy.get('a')
            .should('exist')
            .and('have.attr', 'href', '/help')
            .contains('help')
    })


})