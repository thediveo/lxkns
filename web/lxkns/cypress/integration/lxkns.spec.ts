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

/// <reference types="../support" />

describe('lxkns app', () => {

    before(() => {
        cy.log('loads')
        cy.visit('/')
        cy.waitForReact(2000, '#root')

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
        cy.routerNavigate('/about')
        cy.react('About').contains('Version')
    })

    it('lends a helping hand', () => {
        cy.routerNavigate('/help')
        cy.react('HelpViewer').contains('The information in this help')
    })

})

export { }
