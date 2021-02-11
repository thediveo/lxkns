import React from 'react'
import { mount } from '@cypress/react'
import { NamespaceIcon } from './NamespaceIcon'
import { NamespaceType } from 'models/lxkns'

describe('NamespaceIcon', () => {

    it('renders namespace icons', () => {
        mount(<>
            {Object.values(NamespaceType).map(nstype =>
                <NamespaceIcon key={nstype} id={nstype} type={nstype} />
            )}
        </>)
        Object.values(NamespaceType).forEach(nstype => {
            cy.get(`#${nstype}`)
                .get('svg').should('exist')
        })
    })

    it('does not render non-existing namespace type', () => {
        mount(<NamespaceIcon id="icon" type={'foobar' as NamespaceType} />)
        cy.get('#icon').should('not.exist')
    })

})
