// load type definitions that come with Cypress module
/// <reference types="cypress" />

// Must be declared global to be detected by typescript (allows import/export)
declare global {
    namespace Cypress {
        interface Chainable {
            /**
             * Custom command returning the React history object.
             * 
             * @example cy.history().its('location.pathname')
             */
            history(): Chainable<Element>
        }
    }
}

export {}
