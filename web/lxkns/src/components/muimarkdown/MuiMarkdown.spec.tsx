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

import React, { ComponentType } from 'react'
import { mount } from '@cypress/react'
import { MuiMarkdown } from './MuiMarkdown'
import pDefer from 'p-defer'

import TestMDX from "!babel-loader!mdx-loader!./MuiMarkdown.spec.mdx"


describe('MuiMarkdown', () => {

    it('renders synchronous MDX', () => {
        mount(<MuiMarkdown mdx={TestMDX} />)
        cy.waitForReact()
        cy.get('#headah')
            .should('have.length', 1)
            .contains('Headah')
        cy.get('strong').contains('text')
    })

    it('renders lazy MDX with default fallback', () => {
        const deferredImportPromise = pDefer()
        const deferredMDX = React.lazy(() =>
            (deferredImportPromise.promise as Promise<{ default: ComponentType<any> }>))

        mount(<MuiMarkdown mdx={deferredMDX} />)
        cy.waitForReact()
        cy.get('.MuiSkeleton-root').should('exist')

        cy.then(() => deferredImportPromise.resolve({ default: TestMDX }))
            .get('#headah')
            .should('have.length', 1)
            .contains('Headah')
            // fallback skeleton should be gone by now.
            .get('.MuiSkeleton-root', { timeout: 100 }).should('not.exist')
    })

    it('renders custom fallback', () => {
        const deferredImportPromise = pDefer()
        const deferredMDX = React.lazy(() =>
            (deferredImportPromise.promise as Promise<{ default: ComponentType<any> }>))

        const MyFallback = () => <span id="myfallback">myfallback</span>

        mount(<MuiMarkdown mdx={deferredMDX} fallback={<MyFallback />} />)
        cy.waitForReact()
        cy.get('#myfallback').should('exist')
            .contains('myfallback')
    })

})