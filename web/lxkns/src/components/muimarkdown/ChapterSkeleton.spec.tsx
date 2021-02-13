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
import { ChapterSkeleton } from './ChapterSkeleton'

describe('ChapterSkeleton', () => {

    it('renders', () => {
        mount(
            <ChapterSkeleton rem={10} />
        )
        cy.waitForReact()
        cy.react('Typography', { props: { variant: 'h4' } })
            .should('have.length', 1)
            .find('.MuiSkeleton-root')
        cy.react('Typography', { props: { variant: 'body1' } })
            .should('have.length', 3)
            .find('.MuiSkeleton-root')
    })

})