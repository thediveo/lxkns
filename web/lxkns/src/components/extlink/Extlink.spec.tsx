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