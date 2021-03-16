import React from 'react'
// Copyright 2020 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

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
